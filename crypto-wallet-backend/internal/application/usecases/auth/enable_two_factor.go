package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
	"github.com/crypto-wallet/backend/internal/infrastructure/security"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// EnableTwoFactorInput encapsulates the data required to enable 2FA.
type EnableTwoFactorInput struct {
	UserID string
	Payload dto.EnableTwoFactorRequest
}

// EnableTwoFactorUseCase verifies the provided code and enables TOTP.
type EnableTwoFactorUseCase struct {
	users  repositories.UserRepository
	logger *slog.Logger
}

// NewEnableTwoFactorUseCase constructs the use case.
func NewEnableTwoFactorUseCase(users repositories.UserRepository, logger *slog.Logger) *EnableTwoFactorUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &EnableTwoFactorUseCase{users: users, logger: logger}
}

// Execute validates the verification code and enables two-factor authentication.
func (uc *EnableTwoFactorUseCase) Execute(ctx context.Context, input EnableTwoFactorInput) (*dto.TwoFactorStatusResponse, error) {
	if uc.users == nil {
		return nil, errors.New("enable 2fa: user repository not configured")
	}

	if errs := input.Payload.Validate(); !errs.IsEmpty() {
		return nil, utils.NewAppError(
			"VALIDATION_ERROR",
			"invalid verification code",
			fiber.StatusBadRequest,
			nil,
			errs.ToDetails(),
		)
	}

	userID, err := uuid.Parse(strings.TrimSpace(input.UserID))
	if err != nil {
		return nil, utils.NewAppError(
			"INVALID_USER_ID",
			"user id must be a valid uuid",
			fiber.StatusBadRequest,
			err,
			nil,
		)
	}

	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	secret := strings.TrimSpace(user.GetTwoFactorSecret())
	if secret == "" {
		return nil, utils.NewAppError(
			"TWO_FACTOR_SETUP_REQUIRED",
			"generate a two-factor secret before enabling",
			fiber.StatusPreconditionFailed,
			nil,
			nil,
		)
	}

	if !security.ValidateTOTP(secret, input.Payload.Code) {
		return nil, utils.NewAppError(
			"TWO_FACTOR_CODE_INVALID",
			"verification code is invalid or expired",
			fiber.StatusUnauthorized,
			nil,
			nil,
		)
	}

	entity, ok := user.(*entities.UserEntity)
	if !ok {
		return nil, errors.New("enable 2fa: unexpected user implementation")
	}

	if err := entity.EnableTwoFactor(secret); err != nil {
		return nil, utils.NewAppError(
			"TWO_FACTOR_ENABLE_FAILED",
			"failed to enable two-factor authentication",
			http.StatusInternalServerError,
			err,
			nil,
		)
	}
	entity.Touch(time.Now().UTC())

	if err := uc.users.Update(ctx, entity); err != nil {
		return nil, err
	}

	return &dto.TwoFactorStatusResponse{Enabled: true}, nil
}
