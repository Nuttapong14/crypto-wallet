package exchange

import (
	"context"
	"errors"
	"fmt"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/services"
)

// GetExchangeRate handles retrieving the current exchange rate for a trading pair.
type GetExchangeRate struct {
	exchangeService *services.ExchangeService
}

// NewGetExchangeRate creates a new GetExchangeRate use case.
func NewGetExchangeRate(exchangeService *services.ExchangeService) *GetExchangeRate {
	return &GetExchangeRate{
		exchangeService: exchangeService,
	}
}

// Execute retrieves the current exchange rate for the specified trading pair.
func (uc *GetExchangeRate) Execute(ctx context.Context, baseSymbol, quoteSymbol string) (*dto.ExchangeRateResponse, error) {
	// Validate input parameters
	if baseSymbol == "" {
		return nil, errors.New("base symbol is required")
	}
	if quoteSymbol == "" {
		return nil, errors.New("quote symbol is required")
	}
	if baseSymbol == quoteSymbol {
		return nil, errors.New("base and quote symbols cannot be the same")
	}

	// Get the exchange rate from the domain service
	pair, err := uc.exchangeService.GetExchangeRate(ctx, baseSymbol, quoteSymbol)
	if err != nil {
		if errors.Is(err, services.ErrExchangeInvalidTradingPair) {
			return nil, fmt.Errorf("trading pair %s/%s is not available or inactive", baseSymbol, quoteSymbol)
		}
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	// Convert to response DTO
	response := &dto.ExchangeRateResponse{
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

	return response, nil
}
