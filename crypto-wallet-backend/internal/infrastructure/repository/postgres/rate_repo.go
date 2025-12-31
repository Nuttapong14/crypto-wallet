package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
)

const exchangeRateSelectColumns = `
SELECT
	id,
	symbol,
	price_usd,
	price_change_24h,
	volume_24h,
	market_cap,
	last_updated,
	created_at
FROM exchange_rates`

const priceHistorySelectColumns = `
SELECT
	id,
	symbol,
	interval,
	timestamp,
	open,
	high,
	low,
	close,
	volume,
	created_at
FROM price_history`

var (
	errNilRatePool         = errors.New("rate repository: database pool is not configured")
	errNilExchangeRate     = errors.New("rate repository: exchange rate entity is required")
	errNilPriceHistory     = errors.New("rate repository: price history entity is required")
	errEmptySymbol         = errors.New("rate repository: symbol is required")
	errEmptySymbolList     = errors.New("rate repository: symbol list is required")
)

// RateRepository persists exchange rate and price history aggregates using PostgreSQL.
type RateRepository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewRateRepository constructs a RateRepository backed by the provided pool.
func NewRateRepository(pool *pgxpool.Pool, logger *slog.Logger) *RateRepository {
	if logger == nil {
		logger = slog.Default()
	}
	return &RateRepository{
		pool:   pool,
		logger: logger,
	}
}

// GetRateBySymbol returns an exchange rate matching the supplied symbol.
func (r *RateRepository) GetRateBySymbol(ctx context.Context, symbol string) (entities.ExchangeRate, error) {
	if r.pool == nil {
		return nil, errNilRatePool
	}

	if strings.TrimSpace(symbol) == "" {
		return nil, errEmptySymbol
	}

	row := r.pool.QueryRow(ctx, exchangeRateSelectColumns+" WHERE symbol = $1", strings.ToUpper(symbol))
	rate, err := r.scanExchangeRate(row)
	if err != nil {
		return nil, mapPGError(err)
	}
	return rate, nil
}

// GetRatesBySymbols returns exchange rates matching the supplied symbols.
func (r *RateRepository) GetRatesBySymbols(ctx context.Context, symbols []string) ([]entities.ExchangeRate, error) {
	if r.pool == nil {
		return nil, errNilRatePool
	}

	if len(symbols) == 0 {
		return nil, errEmptySymbolList
	}

	// Normalize symbols to uppercase
	normalizedSymbols := make([]string, len(symbols))
	for i, sym := range symbols {
		normalizedSymbols[i] = strings.ToUpper(strings.TrimSpace(sym))
	}

	rows, err := r.pool.Query(ctx, exchangeRateSelectColumns+" WHERE symbol = ANY($1)", normalizedSymbols)
	if err != nil {
		return nil, mapPGError(err)
	}
	defer rows.Close()

	results := make([]entities.ExchangeRate, 0, len(symbols))
	for rows.Next() {
		rate, scanErr := r.scanExchangeRate(rows)
		if scanErr != nil {
			return nil, mapPGError(scanErr)
		}
		results = append(results, rate)
	}

	if rows.Err() != nil {
		return nil, mapPGError(rows.Err())
	}

	return results, nil
}

// GetAllRates returns all exchange rates.
func (r *RateRepository) GetAllRates(ctx context.Context) ([]entities.ExchangeRate, error) {
	if r.pool == nil {
		return nil, errNilRatePool
	}

	rows, err := r.pool.Query(ctx, exchangeRateSelectColumns+" ORDER BY market_cap DESC")
	if err != nil {
		return nil, mapPGError(err)
	}
	defer rows.Close()

	results := make([]entities.ExchangeRate, 0)
	for rows.Next() {
		rate, scanErr := r.scanExchangeRate(rows)
		if scanErr != nil {
			return nil, mapPGError(scanErr)
		}
		results = append(results, rate)
	}

	if rows.Err() != nil {
		return nil, mapPGError(rows.Err())
	}

	return results, nil
}

// UpsertRate inserts or updates an exchange rate (uses INSERT ... ON CONFLICT).
func (r *RateRepository) UpsertRate(ctx context.Context, rate *entities.ExchangeRateEntity) error {
	if r.pool == nil {
		return errNilRatePool
	}
	if rate == nil {
		return errNilExchangeRate
	}

	query := `
INSERT INTO exchange_rates (
	id,
	symbol,
	price_usd,
	price_change_24h,
	volume_24h,
	market_cap,
	last_updated,
	created_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8
)
ON CONFLICT (symbol)
DO UPDATE SET
	price_usd = EXCLUDED.price_usd,
	price_change_24h = EXCLUDED.price_change_24h,
	volume_24h = EXCLUDED.volume_24h,
	market_cap = EXCLUDED.market_cap,
	last_updated = EXCLUDED.last_updated`

	_, err := r.pool.Exec(ctx, query,
		rate.GetID(),
		rate.GetSymbol(),
		rate.GetPriceUSD().String(),
		rate.GetPriceChange24h().String(),
		rate.GetVolume24h().String(),
		rate.GetMarketCap().String(),
		rate.GetLastUpdated().UTC(),
		rate.GetCreatedAt().UTC(),
	)
	if err != nil {
		return mapPGError(err)
	}

	return nil
}

// CreateRate persists the supplied exchange rate entity.
func (r *RateRepository) CreateRate(ctx context.Context, rate *entities.ExchangeRateEntity) error {
	if r.pool == nil {
		return errNilRatePool
	}
	if rate == nil {
		return errNilExchangeRate
	}

	query := `
INSERT INTO exchange_rates (
	id,
	symbol,
	price_usd,
	price_change_24h,
	volume_24h,
	market_cap,
	last_updated,
	created_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8
)`

	_, err := r.pool.Exec(ctx, query,
		rate.GetID(),
		rate.GetSymbol(),
		rate.GetPriceUSD().String(),
		rate.GetPriceChange24h().String(),
		rate.GetVolume24h().String(),
		rate.GetMarketCap().String(),
		rate.GetLastUpdated().UTC(),
		rate.GetCreatedAt().UTC(),
	)
	if err != nil {
		return mapPGError(err)
	}

	return nil
}

// UpdateRate persists changes to an existing exchange rate entity.
func (r *RateRepository) UpdateRate(ctx context.Context, rate entities.ExchangeRate) error {
	if r.pool == nil {
		return errNilRatePool
	}
	if rate == nil {
		return errNilExchangeRate
	}

	query := `
UPDATE exchange_rates
SET
	price_usd = $2,
	price_change_24h = $3,
	volume_24h = $4,
	market_cap = $5,
	last_updated = $6
WHERE symbol = $1`

	cmd, err := r.pool.Exec(ctx, query,
		rate.GetSymbol(),
		rate.GetPriceUSD().String(),
		rate.GetPriceChange24h().String(),
		rate.GetVolume24h().String(),
		rate.GetMarketCap().String(),
		rate.GetLastUpdated().UTC(),
	)
	if err != nil {
		return mapPGError(err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrNotFound
	}

	return nil
}

// DeleteRate removes an exchange rate by its symbol.
func (r *RateRepository) DeleteRate(ctx context.Context, symbol string) error {
	if r.pool == nil {
		return errNilRatePool
	}

	if strings.TrimSpace(symbol) == "" {
		return errEmptySymbol
	}

	cmd, err := r.pool.Exec(ctx, "DELETE FROM exchange_rates WHERE symbol = $1", strings.ToUpper(symbol))
	if err != nil {
		return mapPGError(err)
	}

	if cmd.RowsAffected() == 0 {
		return repositories.ErrNotFound
	}

	return nil
}

// GetPriceHistoryByID returns a price history entry matching the supplied identifier.
func (r *RateRepository) GetPriceHistoryByID(ctx context.Context, id uuid.UUID) (entities.PriceHistory, error) {
	if r.pool == nil {
		return nil, errNilRatePool
	}

	row := r.pool.QueryRow(ctx, priceHistorySelectColumns+" WHERE id = $1", id)
	history, err := r.scanPriceHistory(row)
	if err != nil {
		return nil, mapPGError(err)
	}
	return history, nil
}

// ListPriceHistory returns price history entries matching the supplied filter with optional pagination.
func (r *RateRepository) ListPriceHistory(ctx context.Context, filter repositories.PriceHistoryFilter, opts repositories.ListOptions) ([]entities.PriceHistory, error) {
	if r.pool == nil {
		return nil, errNilRatePool
	}

	opts = opts.WithDefaults()

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(priceHistorySelectColumns)
	queryBuilder.WriteString(" WHERE 1=1")

	args := []any{}
	argIndex := 1

	if strings.TrimSpace(filter.Symbol) != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND symbol = $%d", argIndex))
		args = append(args, strings.ToUpper(filter.Symbol))
		argIndex++
	}

	if filter.Interval != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND interval = $%d", argIndex))
		args = append(args, string(filter.Interval))
		argIndex++
	}

	if filter.From != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND timestamp >= $%d", argIndex))
		args = append(args, filter.From.UTC())
		argIndex++
	}

	if filter.To != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND timestamp <= $%d", argIndex))
		args = append(args, filter.To.UTC())
		argIndex++
	}

	sortColumn := sanitizePriceHistorySortColumn(opts.SortBy)
	sortOrder := "DESC"
	if opts.SortOrder == repositories.SortAscending {
		sortOrder = "ASC"
	}

	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY %s %s", sortColumn, sortOrder))
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1))
	args = append(args, opts.Limit, opts.Offset)

	rows, err := r.pool.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, mapPGError(err)
	}
	defer rows.Close()

	results := make([]entities.PriceHistory, 0)
	for rows.Next() {
		history, scanErr := r.scanPriceHistory(rows)
		if scanErr != nil {
			return nil, mapPGError(scanErr)
		}
		results = append(results, history)
	}

	if rows.Err() != nil {
		return nil, mapPGError(rows.Err())
	}

	return results, nil
}

// CreatePriceHistory persists the supplied price history entity.
func (r *RateRepository) CreatePriceHistory(ctx context.Context, history *entities.PriceHistoryEntity) error {
	if r.pool == nil {
		return errNilRatePool
	}
	if history == nil {
		return errNilPriceHistory
	}

	query := `
INSERT INTO price_history (
	id,
	symbol,
	interval,
	timestamp,
	open,
	high,
	low,
	close,
	volume,
	created_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (symbol, interval, timestamp) DO NOTHING`

	_, err := r.pool.Exec(ctx, query,
		history.GetID(),
		history.GetSymbol(),
		string(history.GetInterval()),
		history.GetTimestamp().UTC(),
		history.GetOpen().String(),
		history.GetHigh().String(),
		history.GetLow().String(),
		history.GetClose().String(),
		history.GetVolume().String(),
		history.GetCreatedAt().UTC(),
	)
	if err != nil {
		return mapPGError(err)
	}

	return nil
}

// DeleteOldPriceHistory removes price history entries older than the specified time.
func (r *RateRepository) DeleteOldPriceHistory(ctx context.Context, before time.Time) (int64, error) {
	if r.pool == nil {
		return 0, errNilRatePool
	}

	cmd, err := r.pool.Exec(ctx, "DELETE FROM price_history WHERE timestamp < $1", before.UTC())
	if err != nil {
		return 0, mapPGError(err)
	}

	return cmd.RowsAffected(), nil
}

func (r *RateRepository) scanExchangeRate(row pgx.Row) (entities.ExchangeRate, error) {
	var (
		id              uuid.UUID
		symbol          string
		priceUSDStr     string
		priceChange24hStr string
		volume24hStr    string
		marketCapStr    string
		lastUpdated     time.Time
		createdAt       time.Time
	)

	err := row.Scan(
		&id,
		&symbol,
		&priceUSDStr,
		&priceChange24hStr,
		&volume24hStr,
		&marketCapStr,
		&lastUpdated,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}

	priceUSD, err := decimal.NewFromString(priceUSDStr)
	if err != nil {
		return nil, fmt.Errorf("rate repository: parse price_usd: %w", err)
	}

	priceChange24h := decimal.Zero
	if strings.TrimSpace(priceChange24hStr) != "" {
		dec, decErr := decimal.NewFromString(priceChange24hStr)
		if decErr != nil {
			return nil, fmt.Errorf("rate repository: parse price_change_24h: %w", decErr)
		}
		priceChange24h = dec
	}

	volume24h := decimal.Zero
	if strings.TrimSpace(volume24hStr) != "" {
		dec, decErr := decimal.NewFromString(volume24hStr)
		if decErr != nil {
			return nil, fmt.Errorf("rate repository: parse volume_24h: %w", decErr)
		}
		volume24h = dec
	}

	marketCap := decimal.Zero
	if strings.TrimSpace(marketCapStr) != "" {
		dec, decErr := decimal.NewFromString(marketCapStr)
		if decErr != nil {
			return nil, fmt.Errorf("rate repository: parse market_cap: %w", decErr)
		}
		marketCap = dec
	}

	rate := entities.HydrateExchangeRateEntity(entities.ExchangeRateParams{
		ID:             id,
		Symbol:         symbol,
		PriceUSD:       priceUSD,
		PriceChange24h: priceChange24h,
		Volume24h:      volume24h,
		MarketCap:      marketCap,
		LastUpdated:    lastUpdated.UTC(),
		CreatedAt:      createdAt.UTC(),
		UpdatedAt:      time.Now().UTC(), // Set current time since DB doesn't have updated_at
	})

	return rate, nil
}

func (r *RateRepository) scanPriceHistory(row pgx.Row) (entities.PriceHistory, error) {
	var (
		id           uuid.UUID
		symbol       string
		intervalStr  string
		timestamp    time.Time
		openStr      string
		highStr      string
		lowStr       string
		closeStr     string
		volumeStr    string
		createdAt    time.Time
	)

	err := row.Scan(
		&id,
		&symbol,
		&intervalStr,
		&timestamp,
		&openStr,
		&highStr,
		&lowStr,
		&closeStr,
		&volumeStr,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}

	open, err := decimal.NewFromString(openStr)
	if err != nil {
		return nil, fmt.Errorf("rate repository: parse open: %w", err)
	}

	high, err := decimal.NewFromString(highStr)
	if err != nil {
		return nil, fmt.Errorf("rate repository: parse high: %w", err)
	}

	low, err := decimal.NewFromString(lowStr)
	if err != nil {
		return nil, fmt.Errorf("rate repository: parse low: %w", err)
	}

	close, err := decimal.NewFromString(closeStr)
	if err != nil {
		return nil, fmt.Errorf("rate repository: parse close: %w", err)
	}

	volume, err := decimal.NewFromString(volumeStr)
	if err != nil {
		return nil, fmt.Errorf("rate repository: parse volume: %w", err)
	}

	history := entities.HydratePriceHistoryEntity(entities.PriceHistoryParams{
		ID:        id,
		Symbol:    symbol,
		Interval:  entities.IntervalType(intervalStr),
		Timestamp: timestamp.UTC(),
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
		CreatedAt: createdAt.UTC(),
		UpdatedAt: time.Now().UTC(), // Set current time since DB doesn't have updated_at
	})

	return history, nil
}

func sanitizePriceHistorySortColumn(sortBy string) string {
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "symbol":
		return "symbol"
	case "interval":
		return "interval"
	case "timestamp":
		return "timestamp"
	case "close":
		return "close"
	case "volume":
		return "volume"
	default:
		return "timestamp"
	}
}
