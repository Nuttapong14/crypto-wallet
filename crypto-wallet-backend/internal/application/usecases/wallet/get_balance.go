package wallet

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// GetWalletBalanceInput captures parameters for retrieving a wallet balance.
type GetWalletBalanceInput struct {
	WalletID string
}

// GetWalletBalanceUseCase refreshes and returns the balance for a wallet.
type GetWalletBalanceUseCase struct {
	service Service
	logger  *slog.Logger
}

// NewGetWalletBalanceUseCase constructs a GetWalletBalanceUseCase.
func NewGetWalletBalanceUseCase(service Service, logger *slog.Logger) *GetWalletBalanceUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetWalletBalanceUseCase{
		service: service,
		logger:  logger,
	}
}

// Execute runs the get balance workflow.
func (uc *GetWalletBalanceUseCase) Execute(ctx context.Context, input GetWalletBalanceInput) (dto.WalletBalance, error) {
	var validation utils.ValidationErrors

	walletID, err := uuid.Parse(strings.TrimSpace(input.WalletID))
	if err != nil {
		validation.Add("wallet_id", "must be a valid UUID")
	}

	if !validation.IsEmpty() {
		return dto.WalletBalance{}, utils.NewAppError(
			"VALIDATION_ERROR",
			"invalid wallet id",
			fiber.StatusBadRequest,
			validation,
			map[string]any{"errors": validation},
		)
	}

	wallet, balance, err := uc.service.RefreshWalletBalance(ctx, walletID)
	if err != nil {
		return dto.WalletBalance{}, err
	}

	result := mapBalance(wallet, balance, "")
	if result.LastUpdated.IsZero() {
		result.LastUpdated = time.Now().UTC()
	}

	return result, nil
}
