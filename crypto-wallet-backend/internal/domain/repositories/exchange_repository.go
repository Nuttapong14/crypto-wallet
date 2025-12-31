package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/crypto-wallet/backend/internal/domain/entities"
)

// ExchangeOperationFilter captures optional filters when listing exchange operations.
type ExchangeOperationFilter struct {
	Status       *entities.ExchangeStatus
	FromWalletID *uuid.UUID
	ToWalletID   *uuid.UUID
	DateFrom     *time.Time
	DateTo       *time.Time
	MinAmount    *decimal.Decimal
	MaxAmount    *decimal.Decimal
}

// TradingPairFilter captures optional filters when listing trading pairs.
type TradingPairFilter struct {
	BaseSymbol   *string
	QuoteSymbol  *string
	IsActive     *bool
	HasLiquidity *bool
}

// ExchangeOperationRepository defines the persistence contract for exchange operation aggregates.
type ExchangeOperationRepository interface {
	// Exchange operations
	GetByID(ctx context.Context, id uuid.UUID) (entities.ExchangeOperation, error)
	GetByUser(ctx context.Context, userID uuid.UUID, filter ExchangeOperationFilter, opts ListOptions) ([]entities.ExchangeOperation, error)
	GetPendingByUser(ctx context.Context, userID uuid.UUID) ([]entities.ExchangeOperation, error)
	GetExpiredPending(ctx context.Context) ([]entities.ExchangeOperation, error)
	Create(ctx context.Context, operation *entities.ExchangeOperationEntity) error
	Update(ctx context.Context, operation entities.ExchangeOperation) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Statistics and analytics
	GetCountByUser(ctx context.Context, userID uuid.UUID, filter ExchangeOperationFilter) (int64, error)
	GetVolumeByUser(ctx context.Context, userID uuid.UUID, filter ExchangeOperationFilter) (decimal.Decimal, error)
}

// TradingPairRepository defines the persistence contract for trading pair aggregates.
type TradingPairRepository interface {
	// Trading pairs
	GetByID(ctx context.Context, id uuid.UUID) (entities.TradingPair, error)
	GetBySymbols(ctx context.Context, baseSymbol, quoteSymbol string) (entities.TradingPair, error)
	List(ctx context.Context, filter TradingPairFilter, opts ListOptions) ([]entities.TradingPair, error)
	GetActivePairs(ctx context.Context) ([]entities.TradingPair, error)
	GetPairsBySymbol(ctx context.Context, symbol string) ([]entities.TradingPair, error)
	Create(ctx context.Context, pair *entities.TradingPairEntity) error
	Update(ctx context.Context, pair entities.TradingPair) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Batch operations for rate updates
	UpdateRates(ctx context.Context, updates map[uuid.UUID]decimal.Decimal) error
	ResetDailyVolumes(ctx context.Context) error

	// Statistics
	GetActiveCount(ctx context.Context) (int64, error)
	GetTotalDailyVolume(ctx context.Context) (decimal.Decimal, error)
}
