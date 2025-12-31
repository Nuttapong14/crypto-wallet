package handlers

import (
	"errors"
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/application/usecases/auth"
	"github.com/crypto-wallet/backend/internal/interfaces/http/middleware"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// AuthHandler wires authentication use cases to HTTP endpoints.
type AuthHandler struct {
	registerUC      *auth.RegisterUseCase
	loginUC         *auth.LoginUseCase
	logoutUC        *auth.LogoutUseCase
	setup2FAUC      *auth.GenerateTwoFactorSetupUseCase
	enable2FAUC     *auth.EnableTwoFactorUseCase
	disable2FAUC    *auth.DisableTwoFactorUseCase
	twoFactorIssuer string
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(
	registerUC *auth.RegisterUseCase,
	loginUC *auth.LoginUseCase,
	logoutUC *auth.LogoutUseCase,
	setup2FAUC *auth.GenerateTwoFactorSetupUseCase,
	enable2FAUC *auth.EnableTwoFactorUseCase,
	disable2FAUC *auth.DisableTwoFactorUseCase,
	twoFactorIssuer string,
) *AuthHandler {
	return &AuthHandler{
		registerUC:      registerUC,
		loginUC:         loginUC,
		logoutUC:        logoutUC,
		setup2FAUC:      setup2FAUC,
		enable2FAUC:     enable2FAUC,
		disable2FAUC:    disable2FAUC,
		twoFactorIssuer: twoFactorIssuer,
	}
}

// Register handles user registration requests.
func (h *AuthHandler) Register() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payload dto.RegisterRequest
		if err := c.BodyParser(&payload); err != nil {
			resp, status := utils.ToErrorResponse(utils.NewAppError(
				"INVALID_JSON",
				"unable to parse request body",
				fiber.StatusBadRequest,
				err,
				nil,
			))
			return c.Status(status).JSON(resp)
		}

		result, err := h.registerUC.Execute(c.Context(), payload)
		if err != nil {
			resp, status := utils.ToErrorResponse(err)
			return c.Status(status).JSON(resp)
		}

		return c.Status(fiber.StatusCreated).JSON(result)
	}
}

// Login handles user login requests.
func (h *AuthHandler) Login() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payload dto.LoginRequest
		if err := c.BodyParser(&payload); err != nil {
			resp, status := utils.ToErrorResponse(utils.NewAppError(
				"INVALID_JSON",
				"unable to parse request body",
				fiber.StatusBadRequest,
				err,
				nil,
			))
			return c.Status(status).JSON(resp)
		}

		result, err := h.loginUC.Execute(c.Context(), payload)
		if err != nil {
			resp, status := utils.ToErrorResponse(err)
			return c.Status(status).JSON(resp)
		}

		return c.Status(fiber.StatusOK).JSON(result)
	}
}

// Logout handles user logout requests.
func (h *AuthHandler) Logout() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payload dto.LogoutRequest
		if err := c.BodyParser(&payload); err != nil {
			resp, status := utils.ToErrorResponse(utils.NewAppError(
				"INVALID_JSON",
				"unable to parse request body",
				fiber.StatusBadRequest,
				err,
				nil,
			))
			return c.Status(status).JSON(resp)
		}

		if err := h.logoutUC.Execute(c.Context(), payload); err != nil {
			resp, status := utils.ToErrorResponse(err)
			return c.Status(status).JSON(resp)
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

// GenerateTwoFactorSetup initiates the TOTP enrolment process for the authenticated user.
func (h *AuthHandler) GenerateTwoFactorSetup() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if h.setup2FAUC == nil {
			return fiber.NewError(fiber.StatusNotImplemented, "two-factor setup not configured")
		}

		userIDUUID, err := extractUserID(c)
		if err != nil {
			return err
		}

		issuer := c.Query("issuer", h.twoFactorIssuer)
		result, execErr := h.setup2FAUC.Execute(c.UserContext(), auth.GenerateTwoFactorSetupInput{
			UserID: userIDUUID.String(),
			Issuer: issuer,
		})
		if execErr != nil {
			return respondError(c, execErr)
		}

		return c.Status(fiber.StatusOK).JSON(result)
	}
}

// EnableTwoFactor confirms the setup using a verification code.
func (h *AuthHandler) EnableTwoFactor() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if h.enable2FAUC == nil {
			return fiber.NewError(fiber.StatusNotImplemented, "two-factor enable not configured")
		}

		userIDUUID, err := extractUserID(c)
		if err != nil {
			return err
		}

		var payload dto.EnableTwoFactorRequest
		if err := c.BodyParser(&payload); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request payload")
		}

		result, execErr := h.enable2FAUC.Execute(c.UserContext(), auth.EnableTwoFactorInput{
			UserID:  userIDUUID.String(),
			Payload: payload,
		})
		if execErr != nil {
			return respondError(c, execErr)
		}

		return c.Status(fiber.StatusOK).JSON(result)
	}
}

// DisableTwoFactor disables TOTP optionally validating the provided code.
func (h *AuthHandler) DisableTwoFactor() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if h.disable2FAUC == nil {
			return fiber.NewError(fiber.StatusNotImplemented, "two-factor disable not configured")
		}

		userIDUUID, err := extractUserID(c)
		if err != nil {
			return err
		}

		var payload dto.DisableTwoFactorRequest
		if err := c.BodyParser(&payload); err != nil && !errors.Is(err, io.EOF) {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request payload")
		}

		result, execErr := h.disable2FAUC.Execute(c.UserContext(), auth.DisableTwoFactorInput{
			UserID:  userIDUUID.String(),
			Payload: payload,
		})
		if execErr != nil {
			return respondError(c, execErr)
		}

		return c.Status(fiber.StatusOK).JSON(result)
	}
}
