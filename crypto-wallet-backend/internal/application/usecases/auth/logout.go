package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// LogoutUseCase provides a hook for future session invalidation.
type LogoutUseCase struct {
	users repositories.UserRepository
}

// NewLogoutUseCase constructs a LogoutUseCase.
func NewLogoutUseCase(users repositories.UserRepository) *LogoutUseCase {
	return &LogoutUseCase{users: users}
}

// Execute currently validates the user exists. Stateless JWT tokens are invalidated client-side.
func (uc *LogoutUseCase) Execute(ctx context.Context, input dto.LogoutRequest) error {
	if input.UserID == uuid.Nil {
		return utils.NewAppError(
			"VALIDATION_ERROR",
			"user id is required",
			http.StatusBadRequest,
			nil,
			nil,
		)
	}

	_, err := uc.users.GetByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return utils.NewAppError(
				"USER_NOT_FOUND",
				"user not found",
				http.StatusNotFound,
				err,
				nil,
			)
		}
		return err
	}

	return nil
}
