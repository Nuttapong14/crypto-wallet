package handlers

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/crypto-wallet/backend/internal/application/dto"
	usecasewallet "github.com/crypto-wallet/backend/internal/application/usecases/wallet"
	"github.com/crypto-wallet/backend/internal/interfaces/http/middleware"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// WalletHandlerConfig configures the wallet HTTP handler.
type WalletHandlerConfig struct {
	CreateUseCase  *usecasewallet.CreateWalletUseCase
	ListUseCase    *usecasewallet.ListWalletsUseCase
	BalanceUseCase *usecasewallet.GetWalletBalanceUseCase
	Logger         *slog.Logger
}

// WalletHandler exposes wallet-related endpoints.
type WalletHandler struct {
	createUseCase  *usecasewallet.CreateWalletUseCase
	listUseCase    *usecasewallet.ListWalletsUseCase
	balanceUseCase *usecasewallet.GetWalletBalanceUseCase
	logger         *slog.Logger
}

// NewWalletHandler constructs a WalletHandler.
func NewWalletHandler(cfg WalletHandlerConfig) *WalletHandler {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &WalletHandler{
		createUseCase:  cfg.CreateUseCase,
		listUseCase:    cfg.ListUseCase,
		balanceUseCase: cfg.BalanceUseCase,
		logger:         logger,
	}
}

// Register wires wallet routes into the provided router.
func (h *WalletHandler) Register(router fiber.Router) {
	if router == nil {
		return
	}

	router.Get("/", h.handleListWallets)
	router.Post("/", h.handleCreateWallet)
	router.Get("/:id/balance", h.handleGetBalance)
}

func (h *WalletHandler) handleListWallets(c *fiber.Ctx) error {
	if h.listUseCase == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "wallet listing not configured")
	}

	userID, err := h.extractUserID(c)
	if err != nil {
		return h.respondError(c, err)
	}

	limit := parseIntWithDefault(c.Query("limit"), 0)
	offset := parseIntWithDefault(c.Query("offset"), 0)

	input := usecasewallet.ListWalletsInput{
		UserID:    userID,
		Chain:     c.Query("chain"),
		Status:    c.Query("status"),
		Limit:     limit,
		Offset:    offset,
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	result, err := h.listUseCase.Execute(c.UserContext(), input)
	if err != nil {
		return h.respondError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func (h *WalletHandler) handleCreateWallet(c *fiber.Ctx) error {
	if h.createUseCase == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "wallet creation not configured")
	}

	userID, err := h.extractUserID(c)
	if err != nil {
		return h.respondError(c, err)
	}

	var payload dto.CreateWalletRequest
	if err := c.BodyParser(&payload); err != nil {
		return h.respondError(c, fiber.NewError(fiber.StatusBadRequest, "invalid request payload"))
	}

	result, err := h.createUseCase.Execute(c.UserContext(), usecasewallet.CreateWalletInput{
		UserID: userID,
		Chain:  payload.Chain,
		Label:  payload.Label,
	})
	if err != nil {
		return h.respondError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *WalletHandler) handleGetBalance(c *fiber.Ctx) error {
	if h.balanceUseCase == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "wallet balance not configured")
	}

	input := usecasewallet.GetWalletBalanceInput{
		WalletID: c.Params("id"),
	}

	result, err := h.balanceUseCase.Execute(c.UserContext(), input)
	if err != nil {
		return h.respondError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func (h *WalletHandler) respondError(c *fiber.Ctx, err error) error {
	resp, status := utils.ToErrorResponse(err)
	return c.Status(status).JSON(resp)
}

func (h *WalletHandler) extractUserID(c *fiber.Ctx) (string, error) {
	claims := c.Locals(middleware.AuthContextKey)
	if claims == nil {
		return "", fiber.NewError(fiber.StatusUnauthorized, "authentication required")
	}

	switch value := claims.(type) {
	case map[string]any:
		if id, ok := getString(value["user_id"]); ok {
			return id, nil
		}
		if id, ok := getString(value["sub"]); ok {
			return id, nil
		}
	case []any:
		for _, item := range value {
			if str, ok := getString(item); ok {
				return str, nil
			}
		}
	default:
		if str, ok := getString(value); ok {
			return str, nil
		}
	}

	return "", fiber.NewError(fiber.StatusUnauthorized, "user identifier missing from token")
}

func getString(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return "", false
		}
		return v, true
	default:
		return "", false
	}
}

func parseIntWithDefault(value string, fallback int) int {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
