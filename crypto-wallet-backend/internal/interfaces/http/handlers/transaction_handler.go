package handlers

import (
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/crypto-wallet/backend/internal/application/dto"
	usecasetransaction "github.com/crypto-wallet/backend/internal/application/usecases/transaction"
	"github.com/crypto-wallet/backend/internal/interfaces/http/middleware"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// TransactionHandlerConfig configures the transaction HTTP handler.
type TransactionHandlerConfig struct {
	SendUseCase   *usecasetransaction.SendTransactionUseCase
	ListUseCase   *usecasetransaction.ListTransactionsUseCase
	StatusUseCase *usecasetransaction.GetTransactionStatusUseCase
	Logger        *slog.Logger
}

// TransactionHandler exposes transaction-related endpoints.
type TransactionHandler struct {
	sendUC   *usecasetransaction.SendTransactionUseCase
	listUC   *usecasetransaction.ListTransactionsUseCase
	statusUC *usecasetransaction.GetTransactionStatusUseCase
	logger   *slog.Logger
}

// NewTransactionHandler constructs a TransactionHandler.
func NewTransactionHandler(cfg TransactionHandlerConfig) *TransactionHandler {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &TransactionHandler{
		sendUC:   cfg.SendUseCase,
		listUC:   cfg.ListUseCase,
		statusUC: cfg.StatusUseCase,
		logger:   logger,
	}
}

// Register attaches routes to the router.
func (h *TransactionHandler) Register(router fiber.Router) {
	if router == nil {
		return
	}

	router.Post("/", h.handleSend)
	router.Get("/", h.handleList)
	router.Get("/:id", h.handleStatusByID)
	router.Get("/hash/:hash", h.handleStatusByHash)
}

func (h *TransactionHandler) handleSend(c *fiber.Ctx) error {
	if h.sendUC == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "transaction sending not configured")
	}

	userID, err := extractUserID(c)
	if err != nil {
		return err
	}

	var payload dto.SendTransactionRequest
	if err := c.BodyParser(&payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request payload")
	}

	result, err := h.sendUC.Execute(c.UserContext(), usecasetransaction.SendTransactionInput{
		UserID:  userID,
		Payload: payload,
	})
	if err != nil {
		return respondError(c, err)
	}

	return c.Status(fiber.StatusAccepted).JSON(result)
}

func (h *TransactionHandler) handleList(c *fiber.Ctx) error {
	if h.listUC == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "transaction listing not configured")
	}

	walletID := c.Query("walletId")
	if walletID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "walletId query parameter is required")
	}

	limit := parseQueryInt(c, "limit", 50)
	offset := parseQueryInt(c, "offset", 0)

	result, err := h.listUC.Execute(c.UserContext(), usecasetransaction.ListTransactionsInput{
		WalletID:  walletID,
		Status:    c.Query("status"),
		Chain:     c.Query("chain"),
		Limit:     limit,
		Offset:    offset,
		SortBy:    c.Query("sortBy"),
		SortOrder: c.Query("sortOrder"),
	})
	if err != nil {
		return respondError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func (h *TransactionHandler) handleStatusByID(c *fiber.Ctx) error {
	if h.statusUC == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "transaction status not configured")
	}

	result, err := h.statusUC.Execute(c.UserContext(), usecasetransaction.GetTransactionStatusInput{
		TransactionID: c.Params("id"),
	})
	if err != nil {
		return respondError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func (h *TransactionHandler) handleStatusByHash(c *fiber.Ctx) error {
	if h.statusUC == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "transaction status not configured")
	}

	result, err := h.statusUC.Execute(c.UserContext(), usecasetransaction.GetTransactionStatusInput{
		Chain: c.Query("chain"),
		Hash:  c.Params("hash"),
	})
	if err != nil {
		return respondError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func parseQueryInt(c *fiber.Ctx, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
