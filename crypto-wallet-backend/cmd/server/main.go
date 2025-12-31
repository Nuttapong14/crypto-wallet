package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberRecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"

	analyticsusecase "github.com/crypto-wallet/backend/internal/application/usecases/analytics"
	authusecase "github.com/crypto-wallet/backend/internal/application/usecases/auth"
	kycusecase "github.com/crypto-wallet/backend/internal/application/usecases/kyc"
	transactionusecase "github.com/crypto-wallet/backend/internal/application/usecases/transaction"
	"github.com/crypto-wallet/backend/internal/application/usecases/wallet"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/services"
	"github.com/crypto-wallet/backend/internal/infrastructure/blockchain"
	"github.com/crypto-wallet/backend/internal/infrastructure/database"
	"github.com/crypto-wallet/backend/internal/infrastructure/external"
	"github.com/crypto-wallet/backend/internal/infrastructure/logging"
	"github.com/crypto-wallet/backend/internal/infrastructure/repository/postgres"
	"github.com/crypto-wallet/backend/internal/infrastructure/security"
	httproutes "github.com/crypto-wallet/backend/internal/interfaces/http"
	"github.com/crypto-wallet/backend/internal/interfaces/http/handlers"
	httpmiddleware "github.com/crypto-wallet/backend/internal/interfaces/http/middleware"
	"github.com/crypto-wallet/backend/pkg/utils"
)

type appConfig struct {
	Host                string
	Port                int
	Environment         string
	LogLevel            string
	LogFormat           string
	JWTSecret           string
	JWTIssuer           string
	JWTAudience         []string
	JWTLeeway           time.Duration
	CORSAllowOrigins    string
	CORSAllowHeaders    string
	CORSAllowMethods    string
	RateLimitEnabled    bool
	RateLimitRequests   int
	RateLimitWindow     time.Duration
	DatabaseDSNs        map[string]string
	WalletEncryptionKey string
	KYCEncryptionKey    string
	TwoFactorIssuer     string
	Blockchain          struct {
		Bitcoin  blockchain.BitcoinConfig
		Ethereum blockchain.EthereumConfig
		Solana   blockchain.SolanaConfig
		Stellar  blockchain.StellarConfig
	}
	KYCProvider struct {
		BaseURL   string
		APIKey    string
		APISecret string
	}
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		slog.Error("failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger, err := logging.NewLogger(logging.Config{
		Level:     cfg.LogLevel,
		Format:    cfg.LogFormat,
		AddSource: cfg.Environment != "production",
	})
	if err != nil {
		slog.Error("failed to initialise logger", slog.String("error", err.Error()))
		os.Exit(1)
	}

	appLogger := logging.WithComponent(logger, "http")

	jwtService, err := security.NewJWTService(security.JWTConfig{
		Secret:   cfg.JWTSecret,
		Issuer:   cfg.JWTIssuer,
		Audience: cfg.JWTAudience,
		Leeway:   cfg.JWTLeeway,
	})
	if err != nil {
		logger.Error("failed to initialise JWT service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	poolManager := database.NewPoolManager(logging.WithComponent(logger, "database"))
	registerDatabasePools(poolManager, cfg)

	var (
		corePool        *pgxpool.Pool
		kycPool         *pgxpool.Pool
		ratesPool       *pgxpool.Pool
		walletHandler   *handlers.WalletHandler
		authHandler     *handlers.AuthHandler
		analyticsHandler *handlers.AnalyticsHandler
		kycHandler      *handlers.KYCHandler
		kycEnforcer     *httpmiddleware.KYCEnforcer
	)

	if pool, err := poolManager.Get("core"); err != nil {
		logger.Warn("core database pool unavailable", slog.String("error", err.Error()))
	} else {
		corePool = pool
	}

	if pool, err := poolManager.Get("kyc"); err != nil {
		logger.Warn("kyc database pool unavailable", slog.String("error", err.Error()))
	} else {
		kycPool = pool
	}

	if pool, err := poolManager.Get("rates"); err != nil {
		logger.Warn("rates database pool unavailable", slog.String("error", err.Error()))
	} else {
		ratesPool = pool
	}

	if corePool != nil {
		walletHandler = buildWalletHandler(cfg, corePool, logger)
		authHandler = buildAuthHandler(cfg, corePool, jwtService, logger)
	}

	if kycPool != nil {
		kycHandler, kycEnforcer = buildKYCComponents(cfg, kycPool, logger)
	}

	analyticsHandler = buildAnalyticsHandler(cfg, corePool, ratesPool, logger)

	app := fiber.New(fiber.Config{
		AppName:      "crypto-wallet-backend",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			resp, status := utils.ToErrorResponse(err)
			return c.Status(status).JSON(resp)
		},
	})

	app.Use(httpmiddleware.NewRequestContextMiddleware(logging.WithComponent(logger, "request")))
	app.Use(httpmiddleware.NewRequestValidationMiddleware(httpmiddleware.RequestValidationConfig{
		MaxBodyBytes: 1 << 20,
		EnforceJSON:  true,
	}))
	app.Use(httpmiddleware.NewLoggingMiddleware(appLogger))
	app.Use(fiberRecover.New())
	app.Use(httpmiddleware.NewCORSMiddleware(httpmiddleware.CORSConfig{
		AllowOrigins:     cfg.CORSAllowOrigins,
		AllowHeaders:     cfg.CORSAllowHeaders,
		AllowMethods:     cfg.CORSAllowMethods,
		AllowCredentials: true,
	}))

	app.Use(httpmiddleware.NewRateLimitMiddleware(httpmiddleware.RateLimitConfig{
		Enabled:     cfg.RateLimitEnabled,
		MaxRequests: cfg.RateLimitRequests,
		Burst:       int(float64(cfg.RateLimitRequests) * 1.5),
		Window:      cfg.RateLimitWindow,
		ExcludePaths: []string{"/api/v1/health", "/"},
	}))

	authMiddleware := httpmiddleware.NewAuthMiddleware(httpmiddleware.AuthConfig{
		JWTService: jwtService,
		Logger:     logging.WithComponent(logger, "auth"),
	})

	httproutes.RegisterRoutes(app, httproutes.RouteOptions{
		Logger:           logging.WithComponent(logger, "routes"),
		AuthMiddleware:   authMiddleware,
		AuthHandler:      authHandler,
		WalletHandler:    walletHandler,
		AnalyticsHandler: analyticsHandler,
		KYCHandler:       kycHandler,
		KYCEnforcer:      kycEnforcer,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		logger.Info("shutdown signal received, stopping server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			logger.Error("error during server shutdown", slog.String("error", err.Error()))
		}
		poolManager.CloseAll()
	}()

	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	logger.Info("starting server", slog.String("address", address), slog.String("environment", cfg.Environment))
	if err := app.Listen(address); err != nil && !errors.Is(err, fiber.ErrServerClosed) {
		logger.Error("server error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}

func loadConfig() (appConfig, error) {
	cfg := appConfig{
		Host:              getEnv("SERVER_HOST", "0.0.0.0"),
		Environment:       strings.ToLower(getEnv("ENVIRONMENT", "development")),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		LogFormat:         getEnv("LOG_FORMAT", "json"),
		JWTSecret:         getEnv("JWT_SECRET", ""),
		JWTIssuer:         getEnv("JWT_ISSUER", "crypto-wallet"),
		CORSAllowOrigins:  getEnv("CORS_ALLOW_ORIGINS", "*"),
		CORSAllowHeaders:  getEnv("CORS_ALLOW_HEADERS", "Authorization,Content-Type,Accept,X-Request-ID"),
		CORSAllowMethods:  getEnv("CORS_ALLOW_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS"),
		RateLimitEnabled:  getEnvAsBool("RATE_LIMIT_ENABLED", true),
		RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnvAsDuration("RATE_LIMIT_WINDOW", time.Minute),
		JWTLeeway:         getEnvAsDuration("JWT_LEEWAY", 30*time.Second),
		DatabaseDSNs: map[string]string{
			"core":  getEnv("CORE_DB_DSN", ""),
			"kyc":   getEnv("KYC_DB_DSN", ""),
			"rates": getEnv("RATES_DB_DSN", ""),
			"audit": getEnv("AUDIT_DB_DSN", ""),
		},
	}

	cfg.WalletEncryptionKey = getEnv("WALLET_ENCRYPTION_KEY", "")
	cfg.KYCEncryptionKey = getEnv("KYC_ENCRYPTION_KEY", "")
	cfg.TwoFactorIssuer = getEnv("TWO_FACTOR_ISSUER", "Atlas Wallet")
	cfg.KYCProvider.BaseURL = getEnv("KYC_PROVIDER_BASE_URL", "")
	cfg.KYCProvider.APIKey = getEnv("KYC_PROVIDER_API_KEY", "")
	cfg.KYCProvider.APISecret = getEnv("KYC_PROVIDER_API_SECRET", "")

	cfg.Blockchain.Bitcoin = blockchain.BitcoinConfig{
		RPCURL:                getEnv("BTC_RPC_URL", ""),
		RPCUser:               getEnv("BTC_RPC_USER", ""),
		RPCPassword:           getEnv("BTC_RPC_PASSWORD", ""),
		Network:               getEnv("BTC_NETWORK", "mainnet"),
		ConfirmationThreshold: getEnvAsInt("BTC_CONFIRMATIONS", 6),
	}

	cfg.Blockchain.Ethereum = blockchain.EthereumConfig{
		RPCURL:                getEnv("ETH_RPC_URL", ""),
		Network:               getEnv("ETH_NETWORK", "mainnet"),
		ChainID:               int64(getEnvAsInt("ETH_CHAIN_ID", 1)),
		ConfirmationThreshold: getEnvAsInt("ETH_CONFIRMATIONS", 12),
	}

	cfg.Blockchain.Solana = blockchain.SolanaConfig{
		RPCURL:                getEnv("SOL_RPC_URL", ""),
		Network:               getEnv("SOL_NETWORK", "mainnet"),
		ConfirmationThreshold: getEnvAsInt("SOL_CONFIRMATIONS", 32),
		Commitment:            getEnv("SOL_COMMITMENT", "finalized"),
	}

	cfg.Blockchain.Stellar = blockchain.StellarConfig{
		HorizonURL:            getEnv("XLM_HORIZON_URL", ""),
		Network:               getEnv("XLM_NETWORK", "public"),
		ConfirmationThreshold: getEnvAsInt("XLM_CONFIRMATIONS", 1),
	}

	port, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return appConfig{}, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}
	cfg.Port = port

	if aud := strings.TrimSpace(os.Getenv("JWT_AUDIENCE")); aud != "" {
		cfg.JWTAudience = splitAndTrim(aud)
	}

	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return appConfig{}, errors.New("JWT_SECRET must be configured")
	}

	return cfg, nil
}

func registerDatabasePools(manager *database.PoolManager, cfg appConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for name, dsn := range cfg.DatabaseDSNs {
		if strings.TrimSpace(dsn) == "" {
			continue
		}
		if err := manager.Register(ctx, name, database.PoolConfig{DSN: dsn}); err != nil {
			slog.Warn("failed to register database pool", slog.String("name", name), slog.String("error", err.Error()))
		}
	}
}

func buildWalletHandler(cfg appConfig, pool *pgxpool.Pool, logger *slog.Logger) *handlers.WalletHandler {
	if pool == nil {
		return nil
	}
	if logger == nil {
		logger = slog.Default()
	}

	componentLogger := logging.WithComponent(logger, "wallet")
	key, err := resolveEncryptionKey(cfg.WalletEncryptionKey, componentLogger)
	if err != nil {
		componentLogger.Error("failed to resolve wallet encryption key", slog.String("error", err.Error()))
		return nil
	}

	encryptor, err := security.NewAESGCMEncryptor(security.AESGCMConfig{Key: key})
	if err != nil {
		componentLogger.Error("failed to initialise wallet encryptor", slog.String("error", err.Error()))
		return nil
	}

	walletRepo := postgres.NewWalletRepository(pool, logging.WithComponent(logger, "wallet-repository"))

	adapters := map[entities.Chain]blockchain.BlockchainAdapter{
		entities.ChainBTC: blockchain.NewBitcoinAdapter(cfg.Blockchain.Bitcoin, logging.WithComponent(logger, "blockchain-btc")),
		entities.ChainETH: blockchain.NewEthereumAdapter(cfg.Blockchain.Ethereum, logging.WithComponent(logger, "blockchain-eth")),
		entities.ChainSOL: blockchain.NewSolanaAdapter(cfg.Blockchain.Solana, logging.WithComponent(logger, "blockchain-sol")),
		entities.ChainXLM: blockchain.NewStellarAdapter(cfg.Blockchain.Stellar, logging.WithComponent(logger, "blockchain-xlm")),
	}

	service := services.NewWalletService(services.WalletServiceConfig{
		Repository: walletRepo,
		Encryptor:  encryptor,
		Adapters:   adapters,
		Logger:     logging.WithComponent(logger, "wallet-service"),
		Retry:      blockchain.RetryConfig{Attempts: 3, Delay: 350 * time.Millisecond},
	})

	createUC := wallet.NewCreateWalletUseCase(service, logging.WithComponent(logger, "wallet-usecase-create"))
	listUC := wallet.NewListWalletsUseCase(service, logging.WithComponent(logger, "wallet-usecase-list"))
	balanceUC := wallet.NewGetWalletBalanceUseCase(service, logging.WithComponent(logger, "wallet-usecase-balance"))

	return handlers.NewWalletHandler(handlers.WalletHandlerConfig{
		CreateUseCase:  createUC,
		ListUseCase:    listUC,
		BalanceUseCase: balanceUC,
		Logger:         logging.WithComponent(logger, "wallet-handler"),
	})
}

func buildAuthHandler(cfg appConfig, pool *pgxpool.Pool, jwtService *security.JWTService, logger *slog.Logger) *handlers.AuthHandler {
    if pool == nil {
        return nil
    }
    if logger == nil {
        logger = slog.Default()
    }

    componentLogger := logging.WithComponent(logger, "auth")

    hasher, err := security.NewBcryptHasher(security.DefaultBCryptCost)
    if err != nil {
        componentLogger.Error("failed to initialise password hasher", slog.String("error", err.Error()))
        return nil
    }

    userRepo := postgres.NewUserRepository(pool, logging.WithComponent(logger, "user-repository"))

    registerUC := authusecase.NewRegisterUseCase(userRepo, hasher, jwtService, 0, 0)
    loginUC := authusecase.NewLoginUseCase(userRepo, hasher, jwtService, 0, 0)
    logoutUC := authusecase.NewLogoutUseCase(userRepo)
    setup2FAUC := authusecase.NewGenerateTwoFactorSetupUseCase(userRepo, logging.WithComponent(logger, "auth-2fa-setup"))
    enable2FAUC := authusecase.NewEnableTwoFactorUseCase(userRepo, logging.WithComponent(logger, "auth-2fa-enable"))
    disable2FAUC := authusecase.NewDisableTwoFactorUseCase(userRepo, logging.WithComponent(logger, "auth-2fa-disable"))

    return handlers.NewAuthHandler(registerUC, loginUC, logoutUC, setup2FAUC, enable2FAUC, disable2FAUC, cfg.TwoFactorIssuer)
}

func buildAnalyticsHandler(cfg appConfig, corePool, ratesPool *pgxpool.Pool, logger *slog.Logger) *handlers.AnalyticsHandler {
	if logger == nil {
		logger = slog.Default()
	}

	var transactionHistoryUC *transactionusecase.GetTransactionHistoryUseCase
	var exportTransactionsUC *transactionusecase.ExportTransactionsUseCase
	var summaryUC *analyticsusecase.PortfolioSummaryUseCase
	var performanceUC *analyticsusecase.PortfolioPerformanceUseCase

	if corePool != nil {
		txRepo := postgres.NewPostgresTransactionRepository(corePool)
		transactionHistoryUC = transactionusecase.NewGetTransactionHistoryUseCase(txRepo, logging.WithComponent(logger, "analytics-transaction-history"))
		exportTransactionsUC = transactionusecase.NewExportTransactionsUseCase(txRepo, logging.WithComponent(logger, "analytics-transaction-export"))
	}

	if corePool != nil && ratesPool != nil {
		walletRepo := postgres.NewWalletRepository(corePool, logging.WithComponent(logger, "analytics-wallet-repository"))
		rateRepo := postgres.NewRateRepository(ratesPool, logging.WithComponent(logger, "analytics-rate-repository"))
		summaryUC = analyticsusecase.NewPortfolioSummaryUseCase(walletRepo, rateRepo, logging.WithComponent(logger, "analytics-portfolio-summary"))
		performanceUC = analyticsusecase.NewPortfolioPerformanceUseCase(walletRepo, rateRepo, logging.WithComponent(logger, "analytics-portfolio-performance"))
	} else if ratesPool == nil {
		logger.Warn("rates database unavailable for analytics handler")
	}

	if transactionHistoryUC == nil && exportTransactionsUC == nil && summaryUC == nil && performanceUC == nil {
		return nil
	}

	return handlers.NewAnalyticsHandler(handlers.AnalyticsHandlerConfig{
		TransactionHistoryUseCase:   transactionHistoryUC,
		ExportTransactionsUseCase:   exportTransactionsUC,
		PortfolioSummaryUseCase:     summaryUC,
		PortfolioPerformanceUseCase: performanceUC,
	})
}

func buildKYCComponents(cfg appConfig, pool *pgxpool.Pool, logger *slog.Logger) (*handlers.KYCHandler, *httpmiddleware.KYCEnforcer) {
    if pool == nil {
        return nil, nil
    }
    if logger == nil {
        logger = slog.Default()
    }

    componentLogger := logging.WithComponent(logger, "kyc")

    key, err := resolveStrictEncryptionKey(cfg.KYCEncryptionKey, componentLogger)
    if err != nil {
        componentLogger.Error("failed to resolve KYC encryption key", slog.String("error", err.Error()))
        return nil, nil
    }

    encryptor, err := security.NewAESGCMEncryptor(security.AESGCMConfig{Key: key})
    if err != nil {
        componentLogger.Error("failed to initialise KYC encryptor", slog.String("error", err.Error()))
        return nil, nil
    }

    repo := postgres.NewKYCRepository(pool, logging.WithComponent(logger, "kyc-repository"))

    var provider external.KYCProviderClient
    if strings.TrimSpace(cfg.KYCProvider.BaseURL) != "" && strings.TrimSpace(cfg.KYCProvider.APIKey) != "" {
        provider, err = external.NewKYCProviderClient(external.KYCProviderConfig{
            BaseURL: cfg.KYCProvider.BaseURL,
            APIKey:  cfg.KYCProvider.APIKey,
            Secret:  cfg.KYCProvider.APISecret,
            Logger:  logging.WithComponent(logger, "kyc-provider"),
        })
        if err != nil {
            componentLogger.Warn("failed to initialise KYC provider client", slog.String("error", err.Error()))
            provider = nil
        }
    }

    submitUC := kycusecase.NewSubmitKYCUseCase(repo, encryptor, provider, logging.WithComponent(logger, "kyc-submit"))
    uploadUC := kycusecase.NewUploadDocumentUseCase(repo, encryptor, provider, logging.WithComponent(logger, "kyc-upload"))
    statusUC := kycusecase.NewGetKYCStatusUseCase(repo, logging.WithComponent(logger, "kyc-status"))

    handler := handlers.NewKYCHandler(handlers.KYCHandlerConfig{
        SubmitUseCase: submitUC,
        UploadUseCase: uploadUC,
        StatusUseCase: statusUC,
        Logger:        logging.WithComponent(logger, "kyc-handler"),
    })

    enforcer := httpmiddleware.NewKYCEnforcer(httpmiddleware.KYCEnforcerConfig{
        Repository: repo,
        Logger:     logging.WithComponent(logger, "kyc-enforcer"),
    })

    return handler, enforcer
}

func resolveEncryptionKey(encoded string, logger *slog.Logger) ([]byte, error) {
	if logger == nil {
		logger = slog.Default()
	}

	trimmed := strings.TrimSpace(encoded)
	if trimmed == "" {
		logger.Warn("WALLET_ENCRYPTION_KEY not configured; generating ephemeral key")
		return security.GenerateRandomKey()
	}

	key, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		logger.Warn("failed to decode WALLET_ENCRYPTION_KEY; generating ephemeral key", slog.String("error", err.Error()))
		return security.GenerateRandomKey()
	}

	if len(key) != security.AES256KeySize {
		logger.Warn("invalid WALLET_ENCRYPTION_KEY length; generating ephemeral key", slog.Int("length", len(key)))
		return security.GenerateRandomKey()
	}

	return key, nil
}

func resolveStrictEncryptionKey(encoded string, logger *slog.Logger) ([]byte, error) {
	if logger == nil {
		logger = slog.Default()
	}

	trimmed := strings.TrimSpace(encoded)
	if trimmed == "" {
		return nil, errors.New("KYC_ENCRYPTION_KEY must be configured")
	}

	key, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return nil, fmt.Errorf("decode kyc encryption key: %w", err)
	}

	if len(key) != security.AES256KeySize {
		return nil, fmt.Errorf("kyc encryption key must be %d bytes", security.AES256KeySize)
	}

	return key, nil
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvAsBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return boolVal
}

func getEnvAsInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return intVal
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return duration
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
