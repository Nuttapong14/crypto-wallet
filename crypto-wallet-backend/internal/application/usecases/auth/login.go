package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
	"github.com/crypto-wallet/backend/internal/infrastructure/security"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// LoginUseCase authenticates an existing user.
type LoginUseCase struct {
	users       repositories.UserRepository
	hasher      security.PasswordHasher
	tokenIssuer *security.JWTService
	accessTTL   time.Duration
	refreshTTL  time.Duration
	clock       func() time.Time
}

// NewLoginUseCase creates a new login use case instance.
func NewLoginUseCase(
	users repositories.UserRepository,
	hasher security.PasswordHasher,
	tokenIssuer *security.JWTService,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *LoginUseCase {
	if accessTTL <= 0 {
		accessTTL = time.Hour
	}
	if refreshTTL <= 0 {
		refreshTTL = 24 * time.Hour * 7
	}

	return &LoginUseCase{
		users:       users,
		hasher:      hasher,
		tokenIssuer: tokenIssuer,
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
		clock:       time.Now,
	}
}

// Execute validates credentials and returns authentication tokens.
func (uc *LoginUseCase) Execute(ctx context.Context, input dto.LoginRequest) (*dto.AuthResponse, error) {
	errs := input.Validate()
	if !errs.IsEmpty() {
		return nil, utils.NewAppError(
			"VALIDATION_ERROR",
			"login payload invalid",
			http.StatusBadRequest,
			nil,
			validationErrorsToDetails(errs),
		)
	}

	user, err := uc.users.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, invalidCredentialsError()
		}
		return nil, err
	}

	if err := uc.hasher.Compare(user.GetPasswordHash(), input.Password); err != nil {
		return nil, invalidCredentialsError()
	}

	now := uc.clock().UTC()
	if entity, ok := user.(*entities.UserEntity); ok {
		entity.UpdateLastLogin(now)
		entity.Touch(now)
		if updateErr := uc.users.Update(ctx, entity); updateErr != nil {
			return nil, updateErr
		}
	}

	accessTokenExpires := uc.clock().Add(uc.accessTTL)
	accessToken, err := uc.tokenIssuer.GenerateToken(ctx, user.GetID().String(), uc.accessTTL, map[string]any{
		"email": user.GetEmail(),
		"type":  "access",
	})
	if err != nil {
		return nil, err
	}

	refreshTokenExpires := uc.clock().Add(uc.refreshTTL)
	refreshToken, err := uc.tokenIssuer.GenerateToken(ctx, user.GetID().String(), uc.refreshTTL, map[string]any{
		"type": "refresh",
	})
	if err != nil {
		return nil, err
	}

	response := &dto.AuthResponse{
		User: dto.NewAuthUser(user),
		Tokens: dto.AuthTokens{
			AccessToken:      accessToken,
			RefreshToken:     refreshToken,
			ExpiresAt:        accessTokenExpires,
			RefreshExpiresAt: refreshTokenExpires,
		},
	}

	return response, nil
}

func invalidCredentialsError() error {
	return utils.NewAppError(
		"INVALID_CREDENTIALS",
		"incorrect email or password",
		http.StatusUnauthorized,
		nil,
		nil,
	)
}
