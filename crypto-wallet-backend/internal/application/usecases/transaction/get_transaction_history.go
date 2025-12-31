package transaction

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// GetTransactionHistoryInput captures filter parameters for transaction history.
type GetTransactionHistoryInput struct {
	WalletID  *uuid.UUID
	Chain     *entities.Chain
	Type      *entities.TransactionType
	Status    *entities.TransactionStatus
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// GetTransactionHistoryUseCase handles comprehensive transaction history retrieval with filtering.
type GetTransactionHistoryUseCase struct {
	transactions TransactionRepo
	logger       *slog.Logger
}

// NewGetTransactionHistoryUseCase constructs the use case.
func NewGetTransactionHistoryUseCase(repo TransactionRepo, logger *slog.Logger) *GetTransactionHistoryUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetTransactionHistoryUseCase{transactions: repo, logger: logger}
}

// Execute returns a paginated and filtered list of transactions.
func (uc *GetTransactionHistoryUseCase) Execute(ctx context.Context, input GetTransactionHistoryInput) (dto.TransactionListResponse, error) {
	// Build filter options
	filter := repositories.TransactionFilter{
		WalletID:  input.WalletID,
		Chain:     input.Chain,
		Type:      input.Type,
		Status:    input.Status,
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
	}

	// Build list options with defaults
	opts := repositories.ListOptions{
		Limit:     input.Limit,
		Offset:    input.Offset,
		SortBy:    input.SortBy,
		SortOrder: repositories.SortOrder(strings.ToUpper(strings.TrimSpace(input.SortOrder))),
	}
	opts = opts.WithDefaults()

	// Get transactions with filters
	transactions, total, err := uc.transactions.ListWithFilters(ctx, filter, opts)
	if err != nil {
		uc.logger.Error("failed to list transactions with filters", "error", err)
		return dto.TransactionListResponse{}, utils.NewAppError(
			"DATABASE_ERROR",
			"Failed to retrieve transaction history",
			fiber.StatusInternalServerError,
			nil,
			map[string]any{"error": err.Error()},
		)
	}

	// Map to response DTO
	response := dto.TransactionListResponse{
		Items:  mapTransactions(transactions),
		Total:  total,
		Limit:  opts.Limit,
		Offset: opts.Offset,
	}

	uc.logger.Info("successfully retrieved transaction history",
		"count", len(transactions),
		"total", total,
		"limit", opts.Limit,
		"offset", opts.Offset,
	)

	return response, nil
}

// ExecuteFromRequest executes the use case from a DTO request.
func (uc *GetTransactionHistoryUseCase) ExecuteFromRequest(ctx context.Context, req dto.GetTransactionHistoryRequest) (dto.TransactionListResponse, error) {
	// Validate request
	if errs := req.Validate(); len(errs) > 0 {
		return dto.TransactionListResponse{}, utils.NewAppError(
			"VALIDATION_ERROR",
			"Invalid request parameters",
			fiber.StatusBadRequest,
			errs,
			map[string]any{"errors": errs},
		)
	}

	// Parse optional wallet ID
	var walletID *uuid.UUID
	if req.WalletID != "" {
		parsed, err := uuid.Parse(req.WalletID)
		if err != nil {
			return dto.TransactionListResponse{}, utils.NewAppError(
				"VALIDATION_ERROR",
				"Invalid wallet ID format",
				fiber.StatusBadRequest,
				nil,
				map[string]any{"error": err.Error()},
			)
		}
		walletID = &parsed
	}

	// Parse optional chain
	var chain *entities.Chain
	if req.Chain != "" {
		c := entities.Chain(req.Chain)
		chain = &c
	}

	// Parse optional transaction type
	var txType *entities.TransactionType
	if req.Type != "" {
		t := entities.TransactionType(req.Type)
		txType = &t
	}

	// Parse optional status
	var status *entities.TransactionStatus
	if req.Status != "" {
		s := entities.TransactionStatus(req.Status)
		status = &s
	}

	// Parse optional dates
	var startDate, endDate *time.Time
	if req.StartDate != "" {
		t, err := time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			return dto.TransactionListResponse{}, utils.NewAppError(
				"VALIDATION_ERROR",
				"Invalid start date format",
				fiber.StatusBadRequest,
				nil,
				map[string]any{"error": err.Error()},
			)
		}
		startDate = &t
	}

	if req.EndDate != "" {
		t, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			return dto.TransactionListResponse{}, utils.NewAppError(
				"VALIDATION_ERROR",
				"Invalid end date format",
				fiber.StatusBadRequest,
				nil,
				map[string]any{"error": err.Error()},
			)
		}
		endDate = &t
	}

	input := GetTransactionHistoryInput{
		WalletID:  walletID,
		Chain:     chain,
		Type:      txType,
		Status:    status,
		StartDate: startDate,
		EndDate:   endDate,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	return uc.Execute(ctx, input)
}
