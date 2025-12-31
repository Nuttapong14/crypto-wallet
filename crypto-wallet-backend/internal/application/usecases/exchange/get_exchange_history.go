package exchange

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
	"github.com/crypto-wallet/backend/internal/domain/services"
)

// GetExchangeHistory handles retrieving exchange history for a user.
type GetExchangeHistory struct {
	exchangeService *services.ExchangeService
}

// NewGetExchangeHistory creates a new GetExchangeHistory use case.
func NewGetExchangeHistory(exchangeService *services.ExchangeService) *GetExchangeHistory {
	return &GetExchangeHistory{
		exchangeService: exchangeService,
	}
}

// Execute retrieves exchange history for the specified user with pagination and filtering.
func (uc *GetExchangeHistory) Execute(ctx context.Context, userID uuid.UUID, req *dto.ExchangeHistoryRequest) (*dto.ExchangeHistoryResponse, error) {
	// Validate request
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// Build filter from request
	filter := repositories.ExchangeOperationFilter{}

	if req.Status != nil && *req.Status != "" {
		status := entities.ExchangeStatus(*req.Status)
		if !isValidExchangeStatus(status) {
			return nil, errors.New("invalid status value")
		}
		filter.Status = &status
	}

	if req.FromWalletID != nil && *req.FromWalletID != uuid.Nil {
		filter.FromWalletID = req.FromWalletID
	}

	if req.ToWalletID != nil && *req.ToWalletID != uuid.Nil {
		filter.ToWalletID = req.ToWalletID
	}

	if req.DateFrom != nil {
		filter.DateFrom = req.DateFrom
	}

	if req.DateTo != nil {
		filter.DateTo = req.DateTo
	}

	if req.MinAmount != nil && *req.MinAmount != "" {
		minAmount, err := decimal.NewFromString(*req.MinAmount)
		if err != nil {
			return nil, fmt.Errorf("invalid min amount: %w", err)
		}
		filter.MinAmount = &minAmount
	}

	if req.MaxAmount != nil && *req.MaxAmount != "" {
		maxAmount, err := decimal.NewFromString(*req.MaxAmount)
		if err != nil {
			return nil, fmt.Errorf("invalid max amount: %w", err)
		}
		filter.MaxAmount = &maxAmount
	}

	// Calculate offset
	offset := (req.Page - 1) * req.PageSize

	// Build list options
	opts := repositories.ListOptions{
		Limit:     req.PageSize,
		Offset:    offset,
		SortBy:    "created_at",
		SortOrder: repositories.SortDescending,
	}

	// Get operations and total count
	operations, err := uc.exchangeService.GetUserExchangeHistory(ctx, userID, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange history: %w", err)
	}

	total, err := uc.exchangeService.GetUserExchangeHistoryCount(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange count: %w", err)
	}

	// Convert to response DTO
	operationResponses := make([]dto.ExchangeOperationResponse, len(operations))
	for i, op := range operations {
		operationResponses[i] = dto.ExchangeOperationResponse{
			ID:                op.GetID(),
			UserID:            op.GetUserID(),
			FromWalletID:      op.GetFromWalletID(),
			ToWalletID:        op.GetToWalletID(),
			FromAmount:        op.GetFromAmount(),
			ToAmount:          op.GetToAmount(),
			ExchangeRate:      op.GetExchangeRate(),
			FeePercentage:     op.GetFeePercentage(),
			FeeAmount:         op.GetFeeAmount(),
			Status:            string(op.GetStatus()),
			QuoteExpiresAt:    op.GetQuoteExpiresAt(),
			ExecutedAt:        op.GetExecutedAt(),
			FromTransactionID: op.GetFromTransactionID(),
			ToTransactionID:   op.GetToTransactionID(),
			ErrorMessage:      op.GetErrorMessage(),
			CreatedAt:         op.GetCreatedAt(),
			UpdatedAt:         op.GetUpdatedAt(),
		}
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	response := &dto.ExchangeHistoryResponse{
		Operations: operationResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}

	return response, nil
}

// GetExchangeStats retrieves exchange statistics for a user.
func (uc *GetExchangeHistory) GetStats(ctx context.Context, userID uuid.UUID) (*dto.ExchangeStatsResponse, error) {
	return uc.exchangeService.GetExchangeStats(ctx, userID)
}

// GetActiveTradingPairs retrieves all active trading pairs.
func (uc *GetExchangeHistory) GetActiveTradingPairs(ctx context.Context) (*dto.TradingPairsResponse, error) {
	pairs, err := uc.exchangeService.GetActiveTradingPairs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active trading pairs: %w", err)
	}

	pairResponses := make([]dto.ExchangeRateResponse, len(pairs))
	for i, pair := range pairs {
		pairResponses[i] = dto.ExchangeRateResponse{
			BaseSymbol:    pair.GetBaseSymbol(),
			QuoteSymbol:   pair.GetQuoteSymbol(),
			ExchangeRate:  pair.GetExchangeRate(),
			InverseRate:   pair.GetInverseRate(),
			FeePercentage: pair.GetFeePercentage(),
			MinSwapAmount: pair.GetMinSwapAmount(),
			MaxSwapAmount: pair.GetMaxSwapAmount(),
			IsActive:      pair.IsActive(),
			HasLiquidity:  pair.HasLiquidity(),
			LastUpdated:   pair.GetLastUpdated(),
		}
	}

	response := &dto.TradingPairsResponse{
		Pairs: pairResponses,
	}

	return response, nil
}

// isValidExchangeStatus checks if the exchange status is valid.
func isValidExchangeStatus(status entities.ExchangeStatus) bool {
	switch status {
	case entities.ExchangeStatusPending, entities.ExchangeStatusProcessing, entities.ExchangeStatusCompleted, entities.ExchangeStatusFailed, entities.ExchangeStatusCancelled:
		return true
	default:
		return false
	}
}
