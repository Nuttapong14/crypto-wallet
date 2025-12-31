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

// GetCurrentRatesInput captures query parameters for getting current rates.
type GetCurrentRatesInput struct {
	Symbols []string
}

// GetCurrentRatesUseCase returns current exchange rates for cryptocurrencies.
type GetCurrentRatesUseCase struct {
	repository repositories.RateRepository
	logger     *slog.Logger
}

// NewGetCurrentRatesUseCase constructs a GetCurrentRatesUseCase.
func NewGetCurrentRatesUseCase(repository repositories.RateRepository, logger *slog.Logger) *GetCurrentRatesUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetCurrentRatesUseCase{
		repository: repository,
		logger:     logger,
	}
}

// Execute runs the get current rates workflow.
func (uc *GetCurrentRatesUseCase) Execute(ctx context.Context, input GetCurrentRatesInput) (dto.ExchangeRateList, error) {
	var validation utils.ValidationErrors

	// Normalize and validate symbols
	normalizedSymbols := make([]string, 0)
	if len(input.Symbols) > 0 {
		for _, symbol := range input.Symbols {
			normalized := strings.ToUpper(strings.TrimSpace(symbol))
			if normalized == "" {
				continue
			}
			// Validate symbol is supported
			if !entities.IsSupportedSymbol(normalized) {
				validation.Add("symbols", "contains unsupported symbol: "+normalized)
				continue
			}
			normalizedSymbols = append(normalizedSymbols, normalized)
		}
	}

	if !validation.IsEmpty() {
		return dto.ExchangeRateList{}, utils.NewAppError(
			"VALIDATION_ERROR",
			"invalid get rates request",
			fiber.StatusBadRequest,
			validation,
			map[string]any{"errors": validation},
		)
	}

	// Fetch rates from repository
	var rates []entities.ExchangeRate
	var err error

	if len(normalizedSymbols) > 0 {
		rates, err = uc.repository.GetRatesBySymbols(ctx, normalizedSymbols)
	} else {
		// If no symbols specified, get all rates
		rates, err = uc.repository.GetAllRates(ctx)
	}

	if err != nil {
		uc.logger.Error("Failed to fetch exchange rates", "error", err)
		return dto.ExchangeRateList{}, utils.NewAppError(
			"DATABASE_ERROR",
			"failed to fetch exchange rates",
			fiber.StatusInternalServerError,
			err,
			nil,
		)
	}

	// Map entities to DTOs
	rateDTOs := make([]dto.ExchangeRate, len(rates))
	var mostRecentUpdate time.Time

	for i, rate := range rates {
		rateDTOs[i] = dto.ExchangeRate{
			Symbol:         rate.GetSymbol(),
			PriceUSD:       rate.GetPriceUSD().String(),
			PriceChange24h: rate.GetPriceChange24h().String(),
			Volume24h:      rate.GetVolume24h().String(),
			MarketCap:      rate.GetMarketCap().String(),
			LastUpdated:    rate.GetLastUpdated(),
		}

		// Track the most recent update time
		if rate.GetLastUpdated().After(mostRecentUpdate) {
			mostRecentUpdate = rate.GetLastUpdated()
		}
	}

	return dto.ExchangeRateList{
		Rates:       rateDTOs,
		LastUpdated: mostRecentUpdate,
	}, nil
}
