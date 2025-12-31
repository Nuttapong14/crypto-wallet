package wallet

import (
	"context"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// ListWalletsInput captures query parameters for listing wallets.
type ListWalletsInput struct {
	UserID    string
	Chain     string
	Status    string
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// ListWalletsUseCase returns wallets for a user with optional filtering.
type ListWalletsUseCase struct {
	service Service
	logger  *slog.Logger
}

// NewListWalletsUseCase constructs a ListWalletsUseCase.
func NewListWalletsUseCase(service Service, logger *slog.Logger) *ListWalletsUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &ListWalletsUseCase{
		service: service,
		logger:  logger,
	}
}

// Execute runs the list wallets workflow.
func (uc *ListWalletsUseCase) Execute(ctx context.Context, input ListWalletsInput) (dto.WalletList, error) {
	var validation utils.ValidationErrors

	userID, err := uuid.Parse(strings.TrimSpace(input.UserID))
	if err != nil {
		validation.Add("user_id", "must be a valid UUID")
	}

	var chainPtr *entities.Chain
	if strings.TrimSpace(input.Chain) != "" {
		chain := entities.NormalizeChain(input.Chain)
		if chain == "" {
			validation.Add("chain", "must be one of BTC, ETH, SOL, XLM")
		} else {
			chainPtr = &chain
		}
	}

	var statusPtr *entities.WalletStatus
	if strings.TrimSpace(input.Status) != "" {
		status := entities.WalletStatus(strings.ToLower(strings.TrimSpace(input.Status)))
		switch status {
		case entities.WalletStatusActive, entities.WalletStatusArchived:
			statusPtr = &status
		default:
			validation.Add("status", "must be either active or archived")
		}
	}

	var sortOrder repositories.SortOrder
	switch strings.ToUpper(strings.TrimSpace(input.SortOrder)) {
	case string(repositories.SortAscending):
		sortOrder = repositories.SortAscending
	case string(repositories.SortDescending):
		sortOrder = repositories.SortDescending
	case "":
		sortOrder = repositories.SortDescending
	default:
		validation.Add("sort_order", "must be ASC or DESC")
	}

	if !validation.IsEmpty() {
		return dto.WalletList{}, utils.NewAppError(
			"VALIDATION_ERROR",
			"invalid wallet list request",
			fiber.StatusBadRequest,
			validation,
			map[string]any{"errors": validation},
		)
	}

	filter := repositories.WalletFilter{
		Chain:  chainPtr,
		Status: statusPtr,
	}

	opts := repositories.ListOptions{
		Limit:     input.Limit,
		Offset:    input.Offset,
		SortBy:    input.SortBy,
		SortOrder: sortOrder,
	}

	wallets, err := uc.service.ListWallets(ctx, userID, filter, opts)
	if err != nil {
		return dto.WalletList{}, err
	}

	return dto.WalletList{
		Wallets: mapWallets(wallets),
		Total:   len(wallets),
	}, nil
}
