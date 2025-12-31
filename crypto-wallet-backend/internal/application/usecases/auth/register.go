package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
	"github.com/crypto-wallet/backend/internal/infrastructure/security"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// RegisterUseCase handles user registration lifecycle.
type RegisterUseCase struct {
	users       repositories.UserRepository
	hasher      security.PasswordHasher
	tokenIssuer *security.JWTService
	accessTTL   time.Duration
	refreshTTL  time.Duration
	clock       func() time.Time
}

// NewRegisterUseCase constructs the use case with sane defaults.
func NewRegisterUseCase(
	users repositories.UserRepository,
	hasher security.PasswordHasher,
	tokenIssuer *security.JWTService,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *RegisterUseCase {
	if accessTTL <= 0 {
		accessTTL = time.Hour
	}
	if refreshTTL <= 0 {
		refreshTTL = 24 * time.Hour * 7
	}

	return &RegisterUseCase{
		users:       users,
		hasher:      hasher,
		tokenIssuer: tokenIssuer,
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
		clock:       time.Now,
	}
}

// Execute registers a new user and returns authentication tokens.
func (uc *RegisterUseCase) Execute(ctx context.Context, input dto.RegisterRequest) (*dto.AuthResponse, error) {
	errs := input.Validate()
	if !errs.IsEmpty() {
		return nil, utils.NewAppError(
			"VALIDATION_ERROR",
			"registration payload invalid",
			http.StatusBadRequest,
			nil,
			validationErrorsToDetails(errs),
		)
	}

	if existing, err := uc.users.GetByEmail(ctx, input.Email); err == nil && existing != nil {
		return nil, utils.NewAppError(
			"EMAIL_IN_USE",
			"an account with that email already exists",
			http.StatusConflict,
			repositories.ErrDuplicate,
			nil,
		)
	} else if err != nil && !errors.Is(err, repositories.ErrNotFound) {
		return nil, err
	}

	hashedPassword, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	now := uc.clock().UTC()
	params := entities.UserParams{
		ID:           uuid.New(),
		Email:        strings.TrimSpace(input.Email),
		PasswordHash: hashedPassword,
		FirstName:    strings.TrimSpace(input.FirstName),
		LastName:     strings.TrimSpace(input.LastName),
		PhoneNumber:  strings.TrimSpace(input.PhoneNumber),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	entity, err := entities.NewUserEntity(params)
	if err != nil {
		return nil, err
	}

	if err = uc.users.Create(ctx, entity); err != nil {
		if errors.Is(err, repositories.ErrDuplicate) {
			return nil, utils.NewAppError(
				"EMAIL_IN_USE",
				"an account with that email already exists",
				http.StatusConflict,
				err,
				nil,
			)
		}
		return nil, err
	}

	now = uc.clock()
	accessTokenExpires := now.Add(uc.accessTTL)
	accessToken, err := uc.tokenIssuer.GenerateToken(ctx, entity.GetID().String(), uc.accessTTL, map[string]any{
		"email": entity.GetEmail(),
		"type":  "access",
	})
	if err != nil {
		return nil, err
	}

	refreshTokenExpires := now.Add(uc.refreshTTL)
	refreshToken, err := uc.tokenIssuer.GenerateToken(ctx, entity.GetID().String(), uc.refreshTTL, map[string]any{
		"type": "refresh",
	})
	if err != nil {
		return nil, err
	}

	response := &dto.AuthResponse{
		User: dto.NewAuthUser(entity),
		Tokens: dto.AuthTokens{
			AccessToken:      accessToken,
			RefreshToken:     refreshToken,
			ExpiresAt:        accessTokenExpires,
			RefreshExpiresAt: refreshTokenExpires,
		},
	}

	return response, nil
}

func validationErrorsToDetails(errs utils.ValidationErrors) map[string]any {
	if errs.IsEmpty() {
		return nil
	}
	details := make(map[string]any, len(errs))
	for _, err := range errs {
		details[err.Field] = err.Message
	}
	return details
}
