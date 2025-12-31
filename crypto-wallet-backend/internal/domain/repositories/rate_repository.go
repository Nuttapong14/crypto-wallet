package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/domain/entities"
)

// PriceHistoryFilter captures optional filters when querying price history.
type PriceHistoryFilter struct {
	Symbol   string
	Interval entities.IntervalType
	From     *time.Time
	To       *time.Time
}

// RateRepository defines the persistence contract for exchange rate and price history aggregates.
type RateRepository interface {
	// ExchangeRate operations
	GetRateBySymbol(ctx context.Context, symbol string) (entities.ExchangeRate, error)
	GetRatesBySymbols(ctx context.Context, symbols []string) ([]entities.ExchangeRate, error)
	GetAllRates(ctx context.Context) ([]entities.ExchangeRate, error)
	UpsertRate(ctx context.Context, rate *entities.ExchangeRateEntity) error
	CreateRate(ctx context.Context, rate *entities.ExchangeRateEntity) error
	UpdateRate(ctx context.Context, rate entities.ExchangeRate) error
	DeleteRate(ctx context.Context, symbol string) error

	// PriceHistory operations
	GetPriceHistoryByID(ctx context.Context, id uuid.UUID) (entities.PriceHistory, error)
	ListPriceHistory(ctx context.Context, filter PriceHistoryFilter, opts ListOptions) ([]entities.PriceHistory, error)
	CreatePriceHistory(ctx context.Context, history *entities.PriceHistoryEntity) error
	DeleteOldPriceHistory(ctx context.Context, before time.Time) (int64, error)
}
