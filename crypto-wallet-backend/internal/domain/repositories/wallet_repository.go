package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/domain/entities"
)

// WalletFilter captures optional filters when listing wallets.
type WalletFilter struct {
	Chain  *entities.Chain
	Status *entities.WalletStatus
}

// WalletRepository defines the persistence contract for wallet aggregates.
type WalletRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (entities.Wallet, error)
	GetByAddress(ctx context.Context, chain entities.Chain, address string) (entities.Wallet, error)
	ListByUser(ctx context.Context, userID uuid.UUID, filter WalletFilter, opts ListOptions) ([]entities.Wallet, error)
	Create(ctx context.Context, wallet *entities.WalletEntity) error
	Update(ctx context.Context, wallet entities.Wallet) error
	Delete(ctx context.Context, id uuid.UUID) error
}
