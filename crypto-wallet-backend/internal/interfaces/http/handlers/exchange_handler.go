package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/application/usecases/exchange"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// ExchangeHandler handles HTTP requests for exchange operations.
type ExchangeHandler struct {
	getExchangeRate    *exchange.GetExchangeRate
	getExchangeHistory *exchange.GetExchangeHistory
	swapTokens         *exchange.SwapTokens
}

// NewExchangeHandler creates a new ExchangeHandler.
func NewExchangeHandler(
	getExchangeRate *exchange.GetExchangeRate,
	getExchangeHistory *exchange.GetExchangeHistory,
	swapTokens *exchange.SwapTokens,
) *ExchangeHandler {
	return &ExchangeHandler{
		getExchangeRate:    getExchangeRate,
		getExchangeHistory: getExchangeHistory,
		swapTokens:         swapTokens,
	}
}

// GetExchangeRate handles GET /api/v1/exchange/rate
func (h *ExchangeHandler) GetExchangeRate(c *fiber.Ctx) error {
	baseSymbol := c.Query("base_symbol")
	quoteSymbol := c.Query("quote_symbol")

	if baseSymbol == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "base_symbol is required")
	}
	if quoteSymbol == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "quote_symbol is required")
	}

	req := &dto.ExchangeRateRequest{
		BaseSymbol:  baseSymbol,
		QuoteSymbol: quoteSymbol,
	}

	response, err := h.getExchangeRate.Execute(c.Context(), req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, response)
}

// GetQuote handles POST /api/v1/exchange/quote
func (h *ExchangeHandler) GetQuote(c *fiber.Ctx) error {
	var req dto.QuoteRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	response, err := h.swapTokens.GetQuote(c.Context(), &req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, response)
}

// ExecuteSwap handles POST /api/v1/exchange/execute
func (h *ExchangeHandler) ExecuteSwap(c *fiber.Ctx) error {
	var req dto.ExecuteSwapRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	response, err := h.swapTokens.ExecuteSwap(c.Context(), &req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, response)
}

// CancelSwap handles POST /api/v1/exchange/cancel
func (h *ExchangeHandler) CancelSwap(c *fiber.Ctx) error {
	var req dto.CancelSwapRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	response, err := h.swapTokens.CancelSwap(c.Context(), &req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, response)
}

// GetExchangeHistory handles GET /api/v1/exchange/history
func (h *ExchangeHandler) GetExchangeHistory(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("userID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	// Parse query parameters
	req := &dto.ExchangeHistoryRequest{
		Page:     c.QueryInt("page", 1),
		PageSize: c.QueryInt("page_size", 10),
	}

	// Parse optional parameters
	if status := c.Query("status"); status != "" {
		req.Status = &status
	}

	if fromWalletID := c.Query("from_wallet_id"); fromWalletID != "" {
		if id, err := uuid.Parse(fromWalletID); err == nil {
			req.FromWalletID = &id
		}
	}

	if toWalletID := c.Query("to_wallet_id"); toWalletID != "" {
		if id, err := uuid.Parse(toWalletID); err == nil {
			req.ToWalletID = &id
		}
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if t, err := time.Parse(time.RFC3339, dateFrom); err == nil {
			req.DateFrom = &t
		}
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		if t, err := time.Parse(time.RFC3339, dateTo); err == nil {
			req.DateTo = &t
		}
	}

	if minAmount := c.Query("min_amount"); minAmount != "" {
		req.MinAmount = &minAmount
	}

	if maxAmount := c.Query("max_amount"); maxAmount != "" {
		req.MaxAmount = &maxAmount
	}

	response, err := h.getExchangeHistory.Execute(c.Context(), userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, response)
}

// GetExchangeStats handles GET /api/v1/exchange/stats/:userID
func (h *ExchangeHandler) GetExchangeStats(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("userID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	response, err := h.getExchangeHistory.GetStats(c.Context(), userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, response)
}

// GetActiveTradingPairs handles GET /api/v1/exchange/pairs
func (h *ExchangeHandler) GetActiveTradingPairs(c *fiber.Ctx) error {
	response, err := h.getExchangeHistory.GetActiveTradingPairs(c.Context())
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, response)
}
