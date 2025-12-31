package wallet

import (
	"context"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/services"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// CreateWalletInput captures the data required to execute the create wallet use case.
type CreateWalletInput struct {
	UserID string
	Chain  string
	Label  string
}

// CreateWalletUseCase coordinates wallet creation between the transport layer and domain service.
type CreateWalletUseCase struct {
	service Service
	logger  *slog.Logger
}

// NewCreateWalletUseCase constructs a CreateWalletUseCase.
func NewCreateWalletUseCase(service Service, logger *slog.Logger) *CreateWalletUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &CreateWalletUseCase{
		service: service,
		logger:  logger,
	}
}

// Execute performs the wallet creation workflow.
func (uc *CreateWalletUseCase) Execute(ctx context.Context, input CreateWalletInput) (dto.Wallet, error) {
	var validation utils.ValidationErrors

	userID, err := uuid.Parse(strings.TrimSpace(input.UserID))
	if err != nil {
		validation.Add("user_id", "must be a valid UUID")
	}

	chain := entities.NormalizeChain(input.Chain)
	if chain == "" {
		validation.Add("chain", "must be one of BTC, ETH, SOL, XLM")
	}

	if !validation.IsEmpty() {
		return dto.Wallet{}, utils.NewAppError(
			"VALIDATION_ERROR",
			"invalid wallet request",
			fiber.StatusBadRequest,
			validation,
			map[string]any{"errors": validation},
		)
	}

	wallet, err := uc.service.CreateWallet(ctx, services.CreateWalletParams{
		UserID: userID,
		Chain:  chain,
		Label:  input.Label,
	})
	if err != nil {
		return dto.Wallet{}, err
	}

	return mapWalletEntity(wallet), nil
}
