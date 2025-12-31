package httpserver

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/interfaces/http/handlers"
	"github.com/crypto-wallet/backend/internal/interfaces/http/middleware"
)

// DefaultAPIPrefix defines the root path for versioned API routes.
const DefaultAPIPrefix = "/api/v1"

// RouteOptions defines dependencies required to register HTTP routes.
type RouteOptions struct {
	Logger             *slog.Logger
	AuthMiddleware     fiber.Handler
	Prefix             string
	AuthHandler        *handlers.AuthHandler
	WalletHandler      *handlers.WalletHandler
	TransactionHandler *handlers.TransactionHandler
	AnalyticsHandler   *handlers.AnalyticsHandler
	KYCHandler         *handlers.KYCHandler
	KYCEnforcer        *middleware.KYCEnforcer
}

// RegisterRoutes wires application endpoints onto the provided Fiber application.
func RegisterRoutes(app *fiber.App, opts RouteOptions) {
	if app == nil {
		return
	}

	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	prefix := opts.Prefix
	if prefix == "" {
		prefix = DefaultAPIPrefix
	}

	// Public endpoints (no authentication required).
	public := app.Group(prefix)
	registerHealthRoutes(public, logger)

	// Secure endpoints (authentication required).
	if opts.AuthMiddleware != nil {
		secure := public.Group("", opts.AuthMiddleware)
		registerSecureRoutes(secure, logger, opts)
	}

	// Root route for quick diagnostics.
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "crypto-wallet-backend",
			"status":  "ok",
			"version": "v1",
			"time":    time.Now().UTC(),
		})
	})

	logger.Info("http routes registered", slog.String("prefix", prefix))
}

func registerHealthRoutes(router fiber.Router, logger *slog.Logger) {
	router.Get("/health", func(c *fiber.Ctx) error {
		logger.Debug("health check invoked")
		return c.JSON(fiber.Map{
			"status":     "ok",
			"timestamp":  time.Now().UTC(),
			"component":  "api",
			"version":    "v1",
			"request_id": c.Get("X-Request-ID"),
		})
	})
}

func registerSecureRoutes(router fiber.Router, logger *slog.Logger, opts RouteOptions) {
	if opts.AuthHandler != nil {
		authGroup := router.Group("/auth")
		authGroup.Post("/register", opts.AuthHandler.Register())
		authGroup.Post("/login", opts.AuthHandler.Login())
		authGroup.Post("/logout", opts.AuthHandler.Logout())
		authGroup.Post("/2fa/setup", opts.AuthHandler.GenerateTwoFactorSetup())
		authGroup.Post("/2fa/enable", opts.AuthHandler.EnableTwoFactor())
		authGroup.Post("/2fa/disable", opts.AuthHandler.DisableTwoFactor())
		logger.Debug("auth routes registered")
	}

	if opts.KYCHandler != nil {
		kycGroup := router.Group("/kyc")
		opts.KYCHandler.Register(kycGroup)
		logger.Debug("kyc routes registered")
	}

	if opts.WalletHandler != nil {
		walletGroup := router.Group("/wallets")
		opts.WalletHandler.Register(walletGroup)
		logger.Debug("wallet routes registered")
	}

	if opts.TransactionHandler != nil {
		txGroup := router.Group("/transactions")
		if opts.KYCEnforcer != nil {
			txGroup.Use(opts.KYCEnforcer.Require(entities.VerificationLevelBasic))
		}
		opts.TransactionHandler.Register(txGroup)
		logger.Debug("transaction routes registered")
	}

	if opts.AnalyticsHandler != nil {
		analyticsGroup := router.Group("/analytics")
		opts.AnalyticsHandler.Register(analyticsGroup)
		logger.Debug("analytics routes registered")
	}

	logger.Debug("secure routes registered")
}
