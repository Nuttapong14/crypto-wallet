package handlers

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/application/usecases/rates"
)

// RateHandler handles HTTP requests for cryptocurrency exchange rates.
type RateHandler struct {
	getCurrentRatesUseCase  *rates.GetCurrentRatesUseCase
	getPriceHistoryUseCase  *rates.GetPriceHistoryUseCase
	logger                  *slog.Logger
}

// NewRateHandler creates a new rate handler.
func NewRateHandler(
	getCurrentRatesUseCase *rates.GetCurrentRatesUseCase,
	getPriceHistoryUseCase *rates.GetPriceHistoryUseCase,
	logger *slog.Logger,
) *RateHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &RateHandler{
		getCurrentRatesUseCase:  getCurrentRatesUseCase,
		getPriceHistoryUseCase:  getPriceHistoryUseCase,
		logger:                  logger,
	}
}

// GetRates handles GET /v1/rates - Get current exchange rates.
func (h *RateHandler) GetRates(c *fiber.Ctx) error {
	// Parse symbols query parameter
	symbolsParam := c.Query("symbols")
	var symbols []string
	if symbolsParam != "" {
		symbols = strings.Split(symbolsParam, ",")
	}

	input := rates.GetCurrentRatesInput{
		Symbols: symbols,
	}

	result, err := h.getCurrentRatesUseCase.Execute(c.Context(), input)
	if err != nil {
		return err
	}

	return c.JSON(result)
}

// GetPriceHistory handles GET /v1/rates/history - Get historical price data.
func (h *RateHandler) GetPriceHistory(c *fiber.Ctx) error {
	var input rates.GetPriceHistoryInput

	// Parse query parameters
	input.Symbol = c.Query("symbol")
	input.Interval = c.Query("interval")
	input.Limit = c.QueryInt("limit", 100)
	input.Offset = c.QueryInt("offset", 0)
	input.SortBy = c.Query("sort_by", "timestamp")

	// Parse time range
	fromStr := c.Query("from")
	if fromStr != "" {
		fromTime, err := time.Parse(time.RFC3339, fromStr)
		if err == nil {
			input.From = &fromTime
		}
	}

	toStr := c.Query("to")
	if toStr != "" {
		toTime, err := time.Parse(time.RFC3339, toStr)
		if err == nil {
			input.To = &toTime
		}
	}

	result, err := h.getPriceHistoryUseCase.Execute(c.Context(), input)
	if err != nil {
		return err
	}

	return c.JSON(result)
}
