package auth

import (
	"context"
	"errors"
	"log/slog"
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

// DisableTwoFactorInput encapsulates the request to disable 2FA.
type DisableTwoFactorInput struct {
	UserID  string
	Payload dto.DisableTwoFactorRequest
}

// DisableTwoFactorUseCase handles deactivating two-factor authentication.
type DisableTwoFactorUseCase struct {
	users  repositories.UserRepository
	logger *slog.Logger
}

// NewDisableTwoFactorUseCase constructs the use case.
func NewDisableTwoFactorUseCase(users repositories.UserRepository, logger *slog.Logger) *DisableTwoFactorUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &DisableTwoFactorUseCase{users: users, logger: logger}
}

// Execute disables two-factor authentication after optional verification.
func (uc *DisableTwoFactorUseCase) Execute(ctx context.Context, input DisableTwoFactorInput) (*dto.TwoFactorStatusResponse, error) {
	if uc.users == nil {
		return nil, errors.New("disable 2fa: user repository not configured")
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

	if errs := input.Payload.Validate(); !errs.IsEmpty() {
		return nil, utils.NewAppError(
			"VALIDATION_ERROR",
			"invalid verification code",
			fiber.StatusBadRequest,
			nil,
			errs.ToDetails(),
		)
	}

	code := strings.TrimSpace(input.Payload.Code)
	if code != "" {
		secret := strings.TrimSpace(user.GetTwoFactorSecret())
		if secret == "" || !security.ValidateTOTP(secret, code) {
			return nil, utils.NewAppError(
				"TWO_FACTOR_CODE_INVALID",
				"verification code is invalid or expired",
				fiber.StatusUnauthorized,
				nil,
				nil,
			)
		}
	}

	entity, ok := user.(*entities.UserEntity)
	if !ok {
		return nil, errors.New("disable 2fa: unexpected user implementation")
	}

	entity.DisableTwoFactor()
	entity.Touch(time.Now().UTC())

	if err := uc.users.Update(ctx, entity); err != nil {
		return nil, err
	}

	return &dto.TwoFactorStatusResponse{Enabled: false}, nil
}
