package services

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
	"github.com/crypto-wallet/backend/internal/infrastructure/blockchain"
	appLogging "github.com/crypto-wallet/backend/internal/infrastructure/logging"
)

var (
	// ErrUnsupportedChain is returned when attempting to operate on an unknown blockchain.
	ErrUnsupportedChain = errors.New("wallet service: unsupported chain")
	// ErrAdapterNotRegistered indicates that no blockchain adapter is configured for the requested chain.
	ErrAdapterNotRegistered = errors.New("wallet service: blockchain adapter not registered")
	// ErrEncryptorNotConfigured indicates that wallet encryption could not be performed.
	ErrEncryptorNotConfigured = errors.New("wallet service: encryption service not configured")
	// ErrWalletNotFound is returned when the requested wallet cannot be located.
	ErrWalletNotFound = errors.New("wallet service: wallet not found")
)

// KeyEncryptor abstracts encryption of private keys for storage.
type KeyEncryptor interface {
	EncryptToString(plaintext, additionalData []byte) (string, error)
}

// WalletService coordinates wallet operations across repositories and blockchain adapters.
type WalletService struct {
	repo      repositories.WalletRepository
	encryptor KeyEncryptor
	adapters  map[entities.Chain]blockchain.BlockchainAdapter
	logger    *slog.Logger
	now       func() time.Time
	retryCfg  blockchain.RetryConfig
}

// WalletServiceConfig configures a WalletService instance.
type WalletServiceConfig struct {
	Repository repositories.WalletRepository
	Encryptor  KeyEncryptor
	Adapters   map[entities.Chain]blockchain.BlockchainAdapter
	Logger     *slog.Logger
	Now        func() time.Time
	Retry      blockchain.RetryConfig
}

// NewWalletService constructs a WalletService.
func NewWalletService(cfg WalletServiceConfig) *WalletService {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	now := cfg.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}

	adapterMap := make(map[entities.Chain]blockchain.BlockchainAdapter, len(cfg.Adapters))
	for chain, adapter := range cfg.Adapters {
		if adapter != nil {
			adapterMap[chain] = adapter
		}
	}

	return &WalletService{
		repo:      cfg.Repository,
		encryptor: cfg.Encryptor,
		adapters:  adapterMap,
		logger:    logger,
		now:       now,
		retryCfg:  cfg.Retry,
	}
}

// CreateWalletParams captures the data required to create a new wallet.
type CreateWalletParams struct {
	UserID uuid.UUID
	Chain  entities.Chain
	Label  string
}

// CreateWallet generates a new blockchain wallet, encrypts the private key, and persists the aggregate.
func (s *WalletService) CreateWallet(ctx context.Context, params CreateWalletParams) (entities.Wallet, error) {
	logger := appLogging.LoggerFromContext(ctx, s.logger).With(
		slog.String("user_id", params.UserID.String()),
		slog.String("chain", string(params.Chain)),
	)
	logger.Debug("wallet creation initiated")
	if params.UserID == uuid.Nil {
		return nil, fmt.Errorf("wallet service: user id is required")
	}

	chain := entities.NormalizeChain(string(params.Chain))
	if chain == "" || !entities.IsSupportedChain(chain) {
		return nil, ErrUnsupportedChain
	}

	adapter, ok := s.adapters[chain]
	if !ok || adapter == nil {
		return nil, ErrAdapterNotRegistered
	}

	if s.encryptor == nil {
		return nil, ErrEncryptorNotConfigured
	}

	generatedWallet, err := blockchain.Retry(ctx, logger, s.retryCfg, "generate_wallet", func(inner context.Context) (*blockchain.Wallet, error) {
		return adapter.GenerateWallet(inner)
	})
	if err != nil {
		return nil, fmt.Errorf("wallet service: generate wallet: %w", err)
	}
	if generatedWallet == nil {
		return nil, fmt.Errorf("wallet service: blockchain adapter returned nil wallet")
	}

	privateKey := strings.TrimSpace(generatedWallet.PrivateKey)
	if privateKey == "" {
		return nil, fmt.Errorf("wallet service: blockchain adapter returned empty private key")
	}

	address := strings.TrimSpace(generatedWallet.Address)
	if address == "" {
		return nil, fmt.Errorf("wallet service: blockchain adapter returned empty address")
	}

	encryptedKey, err := s.encryptor.EncryptToString([]byte(privateKey), []byte(address))
	if err != nil {
		return nil, fmt.Errorf("wallet service: encrypt private key: %w", err)
	}

	label := strings.TrimSpace(params.Label)
	if label == "" {
		label = fmt.Sprintf("%s Wallet", chain)
	}

	now := s.now()

	entity, err := entities.NewWalletEntity(entities.WalletParams{
		UserID:              params.UserID,
		Chain:               chain,
		Address:             address,
		EncryptedPrivateKey: encryptedKey,
		DerivationPath:      strings.TrimSpace(generatedWallet.DerivationPath),
		Label:               label,
		Balance:             decimal.Zero,
		Status:              entities.WalletStatusActive,
		CreatedAt:           now,
		UpdatedAt:           now,
	})
	if err != nil {
		return nil, fmt.Errorf("wallet service: construct entity: %w", err)
	}

	if err := s.repo.Create(ctx, entity); err != nil {
		logger.Error("failed to persist wallet", slog.String("error", err.Error()))
		return nil, fmt.Errorf("wallet service: persist wallet: %w", err)
	}

	logger.Info("wallet created", slog.String("wallet_id", entity.GetID().String()))

	return entity, nil
}

// ListWallets returns all wallets for a user respecting the provided filters.
func (s *WalletService) ListWallets(ctx context.Context, userID uuid.UUID, filter repositories.WalletFilter, opts repositories.ListOptions) ([]entities.Wallet, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("wallet service: user id is required")
	}
	logger := appLogging.LoggerFromContext(ctx, s.logger).With(slog.String("user_id", userID.String()))
	results, err := s.repo.ListByUser(ctx, userID, filter, opts)
	if err != nil {
		logger.Error("failed to list wallets", slog.String("error", err.Error()))
		return nil, err
	}
	logger.Debug("wallets fetched", slog.Int("count", len(results)))
	return results, nil
}

// GetWalletByID returns a wallet by identifier.
func (s *WalletService) GetWalletByID(ctx context.Context, id uuid.UUID) (entities.Wallet, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("wallet service: wallet id is required")
	}
	logger := appLogging.LoggerFromContext(ctx, s.logger).With(slog.String("wallet_id", id.String()))
	wallet, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			logger.Warn("wallet not found")
			return nil, ErrWalletNotFound
		}
		logger.Error("failed to load wallet", slog.String("error", err.Error()))
		return nil, err
	}
	logger.Debug("wallet retrieved", slog.String("chain", string(wallet.GetChain())))
	return wallet, nil
}

// RefreshWalletBalance pulls the latest balance from the blockchain and persists it.
func (s *WalletService) RefreshWalletBalance(ctx context.Context, walletID uuid.UUID) (entities.Wallet, *blockchain.Balance, error) {
	if walletID == uuid.Nil {
		return nil, nil, fmt.Errorf("wallet service: wallet id is required")
	}

	logger := appLogging.LoggerFromContext(ctx, s.logger).With(slog.String("wallet_id", walletID.String()))

	wallet, err := s.repo.GetByID(ctx, walletID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			logger.Warn("wallet not found during balance refresh")
			return nil, nil, ErrWalletNotFound
		}
		logger.Error("failed to load wallet for balance refresh", slog.String("error", err.Error()))
		return nil, nil, err
	}

	chain := wallet.GetChain()
	adapter, ok := s.adapters[chain]
	if !ok || adapter == nil {
		logger.Error("blockchain adapter missing")
		return nil, nil, ErrAdapterNotRegistered
	}

	balance, err := blockchain.Retry(ctx, logger, s.retryCfg, "get_balance", func(inner context.Context) (*blockchain.Balance, error) {
		return adapter.GetBalance(inner, wallet.GetAddress())
	})
	if err != nil {
		logger.Error("failed to query blockchain balance", slog.String("error", err.Error()))
		return nil, nil, fmt.Errorf("wallet service: get balance: %w", err)
	}
	if balance == nil {
		logger.Warn("blockchain returned empty balance payload")
		return wallet, nil, nil
	}

	balanceValue := decimal.Zero
	balanceString := strings.TrimSpace(balance.Balance)
	if balanceString != "" {
		balanceValue, err = decimal.NewFromString(balanceString)
		if err != nil {
			return nil, nil, fmt.Errorf("wallet service: parse balance: %w", err)
		}
	}

	lastUpdated := balance.LastUpdated
	if lastUpdated.IsZero() {
		lastUpdated = s.now()
	}

	if err := wallet.UpdateBalance(balanceValue, lastUpdated); err != nil {
		logger.Error("failed to update wallet balance", slog.String("error", err.Error()))
		return nil, nil, fmt.Errorf("wallet service: update balance: %w", err)
	}
	wallet.Touch(s.now())

	if err := s.repo.Update(ctx, wallet); err != nil {
		logger.Error("failed to persist wallet balance", slog.String("error", err.Error()))
		return nil, nil, fmt.Errorf("wallet service: persist balance: %w", err)
	}

	logger.Info("wallet balance refreshed",
		slog.String("chain", string(wallet.GetChain())),
		slog.String("address", wallet.GetAddress()),
	)

	return wallet, balance, nil
}

// DecryptPrivateKey attempts to decrypt a previously stored private key using the configured encryptor.
func (s *WalletService) DecryptPrivateKey(encrypted string, address string) (string, error) {
	if s.encryptor == nil {
		return "", ErrEncryptorNotConfigured
	}
	if strings.TrimSpace(encrypted) == "" {
		return "", fmt.Errorf("wallet service: encrypted payload is empty")
	}
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("wallet service: decode payload: %w", err)
	}

	// The encryptor interface only exposes EncryptToString. To support decrypting without
	// introducing a broader interface, we rely on type assertion where available.
	if decryptor, ok := s.encryptor.(interface {
		Decrypt([]byte, []byte) ([]byte, error)
	}); ok {
		plaintext, err := decryptor.Decrypt(data, []byte(address))
		if err != nil {
			return "", fmt.Errorf("wallet service: decrypt payload: %w", err)
		}
		return string(plaintext), nil
	}

	if decryptor, ok := s.encryptor.(interface {
		DecryptString(string, []byte) ([]byte, error)
	}); ok {
		plaintext, err := decryptor.DecryptString(encrypted, []byte(address))
		if err != nil {
			return "", fmt.Errorf("wallet service: decrypt payload: %w", err)
		}
		return string(plaintext), nil
	}

	return "", fmt.Errorf("wallet service: encryptor does not support decryption")
}
