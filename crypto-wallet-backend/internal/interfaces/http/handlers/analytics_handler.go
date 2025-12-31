package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/application/dto"
	analyticsusecase "github.com/crypto-wallet/backend/internal/application/usecases/analytics"
	transactionusecase "github.com/crypto-wallet/backend/internal/application/usecases/transaction"
	"github.com/crypto-wallet/backend/internal/interfaces/http/middleware"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// AnalyticsHandlerConfig groups dependencies required by AnalyticsHandler.
type AnalyticsHandlerConfig struct {
	TransactionHistoryUseCase *transactionusecase.GetTransactionHistoryUseCase
	ExportTransactionsUseCase *transactionusecase.ExportTransactionsUseCase
	PortfolioSummaryUseCase   *analyticsusecase.PortfolioSummaryUseCase
	PortfolioPerformanceUseCase *analyticsusecase.PortfolioPerformanceUseCase
}

// AnalyticsHandler handles analytics-oriented HTTP requests.
type AnalyticsHandler struct {
	transactionHistoryUC   *transactionusecase.GetTransactionHistoryUseCase
	exportTransactionsUC   *transactionusecase.ExportTransactionsUseCase
	portfolioSummaryUC     *analyticsusecase.PortfolioSummaryUseCase
	portfolioPerformanceUC *analyticsusecase.PortfolioPerformanceUseCase
}

// NewAnalyticsHandler constructs an AnalyticsHandler instance.
func NewAnalyticsHandler(cfg AnalyticsHandlerConfig) *AnalyticsHandler {
	return &AnalyticsHandler{
		transactionHistoryUC:   cfg.TransactionHistoryUseCase,
		exportTransactionsUC:   cfg.ExportTransactionsUseCase,
		portfolioSummaryUC:     cfg.PortfolioSummaryUseCase,
		portfolioPerformanceUC: cfg.PortfolioPerformanceUseCase,
	}
}

// GetTransactionHistory handles GET /api/v1/analytics/transactions.
func (h *AnalyticsHandler) GetTransactionHistory(c *fiber.Ctx) error {
	if h.transactionHistoryUC == nil {
		return respondError(c, fiber.NewError(fiber.StatusNotImplemented, "transaction history not configured"))
	}

	req := dto.GetTransactionHistoryRequest{
		WalletID:  c.Query("walletId"),
		Chain:     c.Query("chain"),
		Type:      c.Query("type"),
		Status:    c.Query("status"),
		StartDate: c.Query("startDate"),
		EndDate:   c.Query("endDate"),
		Limit:     50,
		Offset:    0,
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	response, err := h.transactionHistoryUC.ExecuteFromRequest(c.UserContext(), req)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(response)
}

// ExportTransactions handles POST /api/v1/analytics/transactions/export.
func (h *AnalyticsHandler) ExportTransactions(c *fiber.Ctx) error {
	if h.exportTransactionsUC == nil {
		return respondError(c, fiber.NewError(fiber.StatusNotImplemented, "transaction export not configured"))
	}

	var req dto.ExportTransactionsRequest
	if err := c.BodyParser(&req); err != nil {
		return respondError(c, utils.NewAppError(
			"INVALID_REQUEST",
			"invalid request body",
			fiber.StatusBadRequest,
			err,
			nil,
		))
	}

	response, err := h.exportTransactionsUC.ExecuteFromRequest(c.UserContext(), req)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(response)
}

// GetTransactionAnalytics handles GET /api/v1/analytics/transactions/summary.
func (h *AnalyticsHandler) GetTransactionAnalytics(c *fiber.Ctx) error {
	return respondError(c, utils.NewAppError(
		"NOT_IMPLEMENTED",
		"transaction analytics is not yet implemented",
		fiber.StatusNotImplemented,
		nil,
		nil,
	))
}

// GetWalletAnalytics handles GET /api/v1/analytics/wallets/:walletId.
func (h *AnalyticsHandler) GetWalletAnalytics(c *fiber.Ctx) error {
	walletIDStr := c.Params("walletId")
	if _, err := uuid.Parse(walletIDStr); err != nil {
		return respondError(c, utils.NewAppError(
			"VALIDATION_ERROR",
			"invalid wallet ID",
			fiber.StatusBadRequest,
			err,
			map[string]any{"walletId": walletIDStr},
		))
	}

	return respondError(c, utils.NewAppError(
		"NOT_IMPLEMENTED",
		"wallet analytics is not yet implemented",
		fiber.StatusNotImplemented,
		nil,
		nil,
	))
}

// GetPortfolioSummary handles GET /api/v1/analytics/portfolio.
func (h *AnalyticsHandler) GetPortfolioSummary(c *fiber.Ctx) error {
	if h.portfolioSummaryUC == nil {
		return respondError(c, fiber.NewError(fiber.StatusNotImplemented, "portfolio summary not configured"))
	}

	userID, err := extractUserID(c)
	if err != nil {
		return respondError(c, err)
	}

	summary, err := h.portfolioSummaryUC.Execute(c.UserContext(), userID)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(summary)
}

// GetPortfolioPerformance handles GET /api/v1/analytics/performance.
func (h *AnalyticsHandler) GetPortfolioPerformance(c *fiber.Ctx) error {
	if h.portfolioPerformanceUC == nil {
		return respondError(c, fiber.NewError(fiber.StatusNotImplemented, "portfolio performance not configured"))
	}

	userID, err := extractUserID(c)
	if err != nil {
		return respondError(c, err)
	}

	period := c.Query("period", "30d")
	performance, err := h.portfolioPerformanceUC.Execute(c.UserContext(), userID, period)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(performance)
}

// Register registers analytics routes.
func (h *AnalyticsHandler) Register(router fiber.Router) {
	if h == nil || router == nil {
		return
	}

	if h.transactionHistoryUC != nil {
		router.Get("/transactions", h.GetTransactionHistory)
	}

	if h.exportTransactionsUC != nil {
		router.Post("/transactions/export", h.ExportTransactions)
	}

	if h.portfolioSummaryUC != nil {
		router.Get("/portfolio", h.GetPortfolioSummary)
	}

	if h.portfolioPerformanceUC != nil {
		router.Get("/performance", h.GetPortfolioPerformance)
	}

	// Placeholder routes for future analytics endpoints.
	router.Get("/transactions/summary", h.GetTransactionAnalytics)
	router.Get("/wallets/:walletId", h.GetWalletAnalytics)
}

func respondError(c *fiber.Ctx, err error) error {
	resp, status := utils.ToErrorResponse(err)
	return c.Status(status).JSON(resp)
}

func extractUserID(c *fiber.Ctx) (uuid.UUID, error) {
	claims := c.Locals(middleware.AuthContextKey)
	switch value := claims.(type) {
	case map[string]any:
		if v, ok := value["user_id"].(string); ok && v != "" {
			return uuid.Parse(v)
		}
		if v, ok := value["sub"].(string); ok && v != "" {
			return uuid.Parse(v)
		}
	case string:
		if value != "" {
			return uuid.Parse(value)
		}
	}

	return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "authentication required")
}
