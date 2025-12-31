package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/domain/entities"
)

// TransactionFilter captures optional filters when listing transactions.
type TransactionFilter struct {
	WalletID  *uuid.UUID
	Chain     *entities.Chain
	Type      *entities.TransactionType
	Status    *entities.TransactionStatus
	StartDate *time.Time
	EndDate   *time.Time
}

// TransactionRepository defines the persistence contract for transaction aggregates.
type TransactionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (entities.Transaction, error)
	GetByHash(ctx context.Context, chain entities.Chain, hash string) (entities.Transaction, error)
	ListByWallet(ctx context.Context, walletID uuid.UUID, opts ListOptions) ([]entities.Transaction, error)
	ListWithFilters(ctx context.Context, filter TransactionFilter, opts ListOptions) ([]entities.Transaction, int64, error)
	ListPending(ctx context.Context, chain entities.Chain, limit int) ([]entities.Transaction, error)
	Create(ctx context.Context, tx *entities.TransactionEntity) error
	Update(ctx context.Context, tx entities.Transaction) error
}
