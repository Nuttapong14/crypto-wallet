package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/domain/entities"
)

// UserRepository defines the persistence contract for user aggregates.
type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (entities.User, error)
	GetByEmail(ctx context.Context, email string) (entities.User, error)
	List(ctx context.Context, opts ListOptions) ([]entities.User, error)
	Create(ctx context.Context, user *entities.UserEntity) error
	Update(ctx context.Context, user entities.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}
