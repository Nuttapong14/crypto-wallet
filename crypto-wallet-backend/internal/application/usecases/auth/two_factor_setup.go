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

// GenerateTwoFactorSetupInput encapsulates the input for generating a setup payload.
type GenerateTwoFactorSetupInput struct {
	UserID string
	Issuer string
}

// GenerateTwoFactorSetupUseCase generates a shared secret and provisioning URI for 2FA.
type GenerateTwoFactorSetupUseCase struct {
	users  repositories.UserRepository
	logger *slog.Logger
}

// NewGenerateTwoFactorSetupUseCase constructs the use case.
func NewGenerateTwoFactorSetupUseCase(users repositories.UserRepository, logger *slog.Logger) *GenerateTwoFactorSetupUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &GenerateTwoFactorSetupUseCase{users: users, logger: logger}
}

// Execute generates a new TOTP secret for the provided user.
func (uc *GenerateTwoFactorSetupUseCase) Execute(ctx context.Context, input GenerateTwoFactorSetupInput) (*dto.TwoFactorSetupResponse, error) {
	if uc.users == nil {
		return nil, errors.New("two-factor setup: user repository not configured")
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

	secret, err := security.GenerateTOTPSecret()
	if err != nil {
		return nil, utils.NewAppError(
			"TWO_FACTOR_SETUP_FAILED",
			"unable to generate two-factor secret",
			http.StatusInternalServerError,
			err,
			nil,
		)
	}

	issuer := strings.TrimSpace(input.Issuer)
	if issuer == "" {
		issuer = "Atlas Wallet"
	}

	entity, ok := user.(*entities.UserEntity)
	if !ok {
		return nil, errors.New("two-factor setup: unexpected user implementation")
	}

	entity.DisableTwoFactor()
	if err := entity.SetTwoFactorSecret(secret); err != nil {
		return nil, utils.NewAppError(
			"TWO_FACTOR_SETUP_FAILED",
			"unable to assign two-factor secret",
			http.StatusInternalServerError,
			err,
			nil,
		)
	}

	entity.Touch(time.Now().UTC())

	if err := uc.users.Update(ctx, entity); err != nil {
		return nil, err
	}

	otpauthURL := security.GenerateTOTPURI(secret, user.GetEmail(), issuer)
	return &dto.TwoFactorSetupResponse{Secret: secret, OtpauthURL: otpauthURL}, nil
}
