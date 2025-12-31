package rates

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// GetPriceHistoryInput captures query parameters for getting price history.
type GetPriceHistoryInput struct {
	Symbol   string
	Interval string
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
	SortBy   string
}

// GetPriceHistoryUseCase returns historical price data for a cryptocurrency.
type GetPriceHistoryUseCase struct {
	repository repositories.RateRepository
	logger     *slog.Logger
}

// NewGetPriceHistoryUseCase constructs a GetPriceHistoryUseCase.
func NewGetPriceHistoryUseCase(repository repositories.RateRepository, logger *slog.Logger) *GetPriceHistoryUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetPriceHistoryUseCase{
		repository: repository,
		logger:     logger,
	}
}

// Execute runs the get price history workflow.
func (uc *GetPriceHistoryUseCase) Execute(ctx context.Context, input GetPriceHistoryInput) (dto.PriceHistoryList, error) {
	var validation utils.ValidationErrors

	// Validate and normalize symbol
	symbol := strings.ToUpper(strings.TrimSpace(input.Symbol))
	if symbol == "" {
		validation.Add("symbol", "is required")
	} else if !entities.IsSupportedSymbol(symbol) {
		validation.Add("symbol", "must be one of BTC, ETH, SOL, XLM")
	}

	// Validate interval if provided
	var intervalType entities.IntervalType
	if strings.TrimSpace(input.Interval) != "" {
		interval := entities.IntervalType(strings.ToLower(strings.TrimSpace(input.Interval)))
		switch interval {
		case entities.Interval1m, entities.Interval5m, entities.Interval15m,
			entities.Interval1h, entities.Interval4h, entities.Interval1d, entities.Interval1w:
			intervalType = interval
		default:
			validation.Add("interval", "must be one of 1m, 5m, 15m, 1h, 4h, 1d, 1w")
		}
	}

	// Validate time range
	if input.From != nil && input.To != nil && input.To.Before(*input.From) {
		validation.Add("to", "must be after from")
	}

	// Validate pagination
	if input.Limit < 0 {
		validation.Add("limit", "must be non-negative")
	}
	if input.Offset < 0 {
		validation.Add("offset", "must be non-negative")
	}

	if !validation.IsEmpty() {
		return dto.PriceHistoryList{}, utils.NewAppError(
			"VALIDATION_ERROR",
			"invalid price history request",
			fiber.StatusBadRequest,
			validation,
			map[string]any{"errors": validation},
		)
	}

	// Build filter
	filter := repositories.PriceHistoryFilter{
		Symbol:   symbol,
		Interval: intervalType,
		From:     input.From,
		To:       input.To,
	}

	// Build list options
	opts := repositories.ListOptions{
		Limit:     input.Limit,
		Offset:    input.Offset,
		SortBy:    input.SortBy,
		SortOrder: repositories.SortDescending, // Most recent first by default
	}

	// Set defaults
	if opts.Limit == 0 {
		opts.Limit = 100 // Default to 100 data points
	}

	// Fetch price history from repository
	history, err := uc.repository.ListPriceHistory(ctx, filter, opts)
	if err != nil {
		uc.logger.Error("Failed to fetch price history",
			"symbol", symbol,
			"error", err)
		return dto.PriceHistoryList{}, utils.NewAppError(
			"DATABASE_ERROR",
			"failed to fetch price history",
			fiber.StatusInternalServerError,
			err,
			nil,
		)
	}

	// Map entities to DTOs
	historyDTOs := make([]dto.PriceHistory, len(history))
	for i, h := range history {
		historyDTOs[i] = dto.PriceHistory{
			ID:        h.GetID(),
			Symbol:    h.GetSymbol(),
			Timestamp: h.GetTimestamp(),
			Open:      h.GetOpen().String(),
			High:      h.GetHigh().String(),
			Low:       h.GetLow().String(),
			Close:     h.GetClose().String(),
			Volume:    h.GetVolume().String(),
		}
	}

	intervalStr := "all"
	if intervalType != "" {
		intervalStr = string(intervalType)
	}

	return dto.PriceHistoryList{
		Symbol:   symbol,
		Interval: intervalStr,
		Data:     historyDTOs,
	}, nil
}
