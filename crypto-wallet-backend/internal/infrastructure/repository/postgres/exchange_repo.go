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

const exchangeOperationSelectColumns = `
SELECT
	id,
	user_id,
	from_wallet_id,
	to_wallet_id,
	from_amount,
	to_amount,
	exchange_rate,
	fee_percentage,
	fee_amount,
	status,
	quote_expires_at,
	executed_at,
	from_transaction_id,
	to_transaction_id,
	error_message,
	created_at,
	updated_at
FROM exchange_operations`

const tradingPairSelectColumns = `
SELECT
	id,
	base_symbol,
	quote_symbol,
	exchange_rate,
	inverse_rate,
	fee_percentage,
	min_swap_amount,
	max_swap_amount,
	daily_volume,
	is_active,
	has_liquidity,
	last_updated,
	created_at,
	updated_at
FROM trading_pairs`

var (
	errExchangeNilPool              = errors.New("exchange repository: database pool is not configured")
	errExchangeNilExchangeOperation = errors.New("exchange repository: exchange operation entity is required")
	errExchangeNilTradingPair       = errors.New("exchange repository: trading pair entity is required")
)

// ExchangeOperationRepository persists exchange operation aggregates using PostgreSQL.
type ExchangeOperationRepository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// TradingPairRepository persists trading pair aggregates using PostgreSQL.
type TradingPairRepository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewExchangeOperationRepository constructs an ExchangeOperationRepository backed by the provided pool.
func NewExchangeOperationRepository(pool *pgxpool.Pool, logger *slog.Logger) *ExchangeOperationRepository {
	if logger == nil {
		logger = slog.Default()
	}
	return &ExchangeOperationRepository{
		pool:   pool,
		logger: logger,
	}
}

// NewTradingPairRepository constructs a TradingPairRepository backed by the provided pool.
func NewTradingPairRepository(pool *pgxpool.Pool, logger *slog.Logger) *TradingPairRepository {
	if logger == nil {
		logger = slog.Default()
	}
	return &TradingPairRepository{
		pool:   pool,
		logger: logger,
	}
}

// ExchangeOperationRepository methods

// GetByID returns an exchange operation matching the supplied identifier.
func (r *ExchangeOperationRepository) GetByID(ctx context.Context, id uuid.UUID) (entities.ExchangeOperation, error) {
	if r.pool == nil {
		return nil, errExchangeNilPool
	}

	row := r.pool.QueryRow(ctx, exchangeOperationSelectColumns+" WHERE id = $1", id)
	operation, err := r.scanExchangeOperation(row)
	if err != nil {
		return nil, mapPGError(err)
	}
	return operation, nil
}

// GetByUser returns exchange operations belonging to the specified user with optional filters.
func (r *ExchangeOperationRepository) GetByUser(ctx context.Context, userID uuid.UUID, filter repositories.ExchangeOperationFilter, opts repositories.ListOptions) ([]entities.ExchangeOperation, error) {
	if r.pool == nil {
		return nil, errExchangeNilPool
	}

	opts = opts.WithDefaults()

	sortColumn := sanitizeExchangeOperationSortColumn(opts.SortBy)
	sortOrder := "DESC"
	if opts.SortOrder == repositories.SortAscending {
		sortOrder = "ASC"
	}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(exchangeOperationSelectColumns)
	queryBuilder.WriteString(" WHERE user_id = $1")

	args := []any{userID}
	argIndex := 2

	if filter.Status != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND status = $%d", argIndex))
		args = append(args, string(*filter.Status))
		argIndex++
	}

	if filter.FromWalletID != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND from_wallet_id = $%d", argIndex))
		args = append(args, *filter.FromWalletID)
		argIndex++
	}

	if filter.ToWalletID != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND to_wallet_id = $%d", argIndex))
		args = append(args, *filter.ToWalletID)
		argIndex++
	}

	if filter.DateFrom != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND created_at >= $%d", argIndex))
		args = append(args, filter.DateFrom.UTC())
		argIndex++
	}

	if filter.DateTo != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND created_at <= $%d", argIndex))
		args = append(args, filter.DateTo.UTC())
		argIndex++
	}

	if filter.MinAmount != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND from_amount >= $%d", argIndex))
		args = append(args, filter.MinAmount.String())
		argIndex++
	}

	if filter.MaxAmount != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND from_amount <= $%d", argIndex))
		args = append(args, filter.MaxAmount.String())
		argIndex++
	}

	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY %s %s", sortColumn, sortOrder))
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1))
	args = append(args, opts.Limit, opts.Offset)

	rows, err := r.pool.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, mapPGError(err)
	}
	defer rows.Close()

	results := make([]entities.ExchangeOperation, 0)
	for rows.Next() {
		operation, scanErr := r.scanExchangeOperation(rows)
		if scanErr != nil {
			return nil, mapPGError(scanErr)
		}
		results = append(results, operation)
	}

	if rows.Err() != nil {
		return nil, mapPGError(rows.Err())
	}

	return results, nil
}

// GetPendingByUser returns pending exchange operations for a user.
func (r *ExchangeOperationRepository) GetPendingByUser(ctx context.Context, userID uuid.UUID) ([]entities.ExchangeOperation, error) {
	if r.pool == nil {
		return nil, errExchangeNilPool
	}

	query := exchangeOperationSelectColumns + " WHERE user_id = $1 AND status = $2"
	rows, err := r.pool.Query(ctx, query, userID, string(entities.ExchangeStatusPending))
	if err != nil {
		return nil, mapPGError(err)
	}
	defer rows.Close()

	results := make([]entities.ExchangeOperation, 0)
	for rows.Next() {
		operation, scanErr := r.scanExchangeOperation(rows)
		if scanErr != nil {
			return nil, mapPGError(scanErr)
		}
		results = append(results, operation)
	}

	if rows.Err() != nil {
		return nil, mapPGError(rows.Err())
	}

	return results, nil
}

// GetExpiredPending returns pending exchange operations that have expired.
func (r *ExchangeOperationRepository) GetExpiredPending(ctx context.Context) ([]entities.ExchangeOperation, error) {
	if r.pool == nil {
		return nil, errExchangeNilPool
	}

	query := exchangeOperationSelectColumns + " WHERE status = $1 AND quote_expires_at <= $2"
	rows, err := r.pool.Query(ctx, query, string(entities.ExchangeStatusPending), time.Now().UTC())
	if err != nil {
		return nil, mapPGError(err)
	}
	defer rows.Close()

	results := make([]entities.ExchangeOperation, 0)
	for rows.Next() {
		operation, scanErr := r.scanExchangeOperation(rows)
		if scanErr != nil {
			return nil, mapPGError(scanErr)
		}
		results = append(results, operation)
	}

	if rows.Err() != nil {
		return nil, mapPGError(rows.Err())
	}

	return results, nil
}

// Create persists the supplied exchange operation entity.
func (r *ExchangeOperationRepository) Create(ctx context.Context, operation *entities.ExchangeOperationEntity) error {
	if r.pool == nil {
		return errExchangeNilPool
	}
	if operation == nil {
		return errExchangeNilExchangeOperation
	}

	query := `
INSERT INTO exchange_operations (
	id,
	user_id,
	from_wallet_id,
	to_wallet_id,
	from_amount,
	to_amount,
	exchange_rate,
	fee_percentage,
	fee_amount,
	status,
	quote_expires_at,
	executed_at,
	from_transaction_id,
	to_transaction_id,
	error_message,
	created_at,
	updated_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
)`

	var executedAt any
	if ts := operation.GetExecutedAt(); ts != nil {
		executedAt = ts.UTC()
	}

	_, err := r.pool.Exec(ctx, query,
		operation.GetID(),
		operation.GetUserID(),
		operation.GetFromWalletID(),
		operation.GetToWalletID(),
		operation.GetFromAmount().String(),
		operation.GetToAmount().String(),
		operation.GetExchangeRate().String(),
		operation.GetFeePercentage().String(),
		operation.GetFeeAmount().String(),
		string(operation.GetStatus()),
		operation.GetQuoteExpiresAt().UTC(),
		executedAt,
		operation.GetFromTransactionID(),
		operation.GetToTransactionID(),
		operation.GetErrorMessage(),
		operation.GetCreatedAt().UTC(),
		operation.GetUpdatedAt().UTC(),
	)
	if err != nil {
		return mapPGError(err)
	}

	return nil
}

// Update persists changes to an existing exchange operation entity.
func (r *ExchangeOperationRepository) Update(ctx context.Context, operation entities.ExchangeOperation) error {
	if r.pool == nil {
		return errExchangeNilPool
	}
	if operation == nil {
		return errExchangeNilExchangeOperation
	}

	query := `
UPDATE exchange_operations
SET
	to_amount = $2,
	exchange_rate = $3,
	fee_percentage = $4,
	fee_amount = $5,
	status = $6,
	quote_expires_at = $7,
	executed_at = $8,
	from_transaction_id = $9,
	to_transaction_id = $10,
	error_message = $11,
	updated_at = $12
WHERE id = $1`

	var executedAt any
	if ts := operation.GetExecutedAt(); ts != nil {
		executedAt = ts.UTC()
	}

	cmd, err := r.pool.Exec(ctx, query,
		operation.GetID(),
		operation.GetToAmount().String(),
		operation.GetExchangeRate().String(),
		operation.GetFeePercentage().String(),
		operation.GetFeeAmount().String(),
		string(operation.GetStatus()),
		operation.GetQuoteExpiresAt().UTC(),
		executedAt,
		operation.GetFromTransactionID(),
		operation.GetToTransactionID(),
		operation.GetErrorMessage(),
		operation.GetUpdatedAt().UTC(),
	)
	if err != nil {
		return mapPGError(err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrNotFound
	}

	return nil
}

// Delete removes an exchange operation by its identifier.
func (r *ExchangeOperationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if r.pool == nil {
		return errExchangeNilPool
	}

	cmd, err := r.pool.Exec(ctx, "DELETE FROM exchange_operations WHERE id = $1", id)
	if err != nil {
		return mapPGError(err)
	}

	if cmd.RowsAffected() == 0 {
		return repositories.ErrNotFound
	}

	return nil
}

// GetCountByUser returns the count of exchange operations for a user with optional filters.
func (r *ExchangeOperationRepository) GetCountByUser(ctx context.Context, userID uuid.UUID, filter repositories.ExchangeOperationFilter) (int64, error) {
	if r.pool == nil {
		return 0, errExchangeNilPool
	}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString("SELECT COUNT(*) FROM exchange_operations WHERE user_id = $1")

	args := []any{userID}
	argIndex := 2

	if filter.Status != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND status = $%d", argIndex))
		args = append(args, string(*filter.Status))
		argIndex++
	}

	if filter.FromWalletID != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND from_wallet_id = $%d", argIndex))
		args = append(args, *filter.FromWalletID)
		argIndex++
	}

	if filter.ToWalletID != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND to_wallet_id = $%d", argIndex))
		args = append(args, *filter.ToWalletID)
		argIndex++
	}

	if filter.DateFrom != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND created_at >= $%d", argIndex))
		args = append(args, filter.DateFrom.UTC())
		argIndex++
	}

	if filter.DateTo != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND created_at <= $%d", argIndex))
		args = append(args, filter.DateTo.UTC())
		argIndex++
	}

	if filter.MinAmount != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND from_amount >= $%d", argIndex))
		args = append(args, filter.MinAmount.String())
		argIndex++
	}

	if filter.MaxAmount != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND from_amount <= $%d", argIndex))
		args = append(args, filter.MaxAmount.String())
		argIndex++
	}

	var count int64
	err := r.pool.QueryRow(ctx, queryBuilder.String(), args...).Scan(&count)
	if err != nil {
		return 0, mapPGError(err)
	}

	return count, nil
}

// GetVolumeByUser returns the total volume of exchange operations for a user with optional filters.
func (r *ExchangeOperationRepository) GetVolumeByUser(ctx context.Context, userID uuid.UUID, filter repositories.ExchangeOperationFilter) (decimal.Decimal, error) {
	if r.pool == nil {
		return decimal.Zero, errExchangeNilPool
	}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString("SELECT COALESCE(SUM(from_amount), 0) FROM exchange_operations WHERE user_id = $1")

	args := []any{userID}
	argIndex := 2

	if filter.Status != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND status = $%d", argIndex))
		args = append(args, string(*filter.Status))
		argIndex++
	}

	if filter.FromWalletID != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND from_wallet_id = $%d", argIndex))
		args = append(args, *filter.FromWalletID)
		argIndex++
	}

	if filter.ToWalletID != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND to_wallet_id = $%d", argIndex))
		args = append(args, *filter.ToWalletID)
		argIndex++
	}

	if filter.DateFrom != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND created_at >= $%d", argIndex))
		args = append(args, filter.DateFrom.UTC())
		argIndex++
	}

	if filter.DateTo != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND created_at <= $%d", argIndex))
		args = append(args, filter.DateTo.UTC())
		argIndex++
	}

	var volumeStr string
	err := r.pool.QueryRow(ctx, queryBuilder.String(), args...).Scan(&volumeStr)
	if err != nil {
		return decimal.Zero, mapPGError(err)
	}

	volume, err := decimal.NewFromString(volumeStr)
	if err != nil {
		return decimal.Zero, fmt.Errorf("exchange repository: parse volume: %w", err)
	}

	return volume, nil
}

// TradingPairRepository methods

// GetByID returns a trading pair matching the supplied identifier.
func (r *TradingPairRepository) GetByID(ctx context.Context, id uuid.UUID) (entities.TradingPair, error) {
	if r.pool == nil {
		return nil, errExchangeNilPool
	}

	row := r.pool.QueryRow(ctx, tradingPairSelectColumns+" WHERE id = $1", id)
	pair, err := r.scanTradingPair(row)
	if err != nil {
		return nil, mapPGError(err)
	}
	return pair, nil
}

// GetBySymbols returns a trading pair matching the supplied symbols.
func (r *TradingPairRepository) GetBySymbols(ctx context.Context, baseSymbol, quoteSymbol string) (entities.TradingPair, error) {
	if r.pool == nil {
		return nil, errExchangeNilPool
	}

	row := r.pool.QueryRow(ctx, tradingPairSelectColumns+" WHERE base_symbol = $1 AND quote_symbol = $2", baseSymbol, quoteSymbol)
	pair, err := r.scanTradingPair(row)
	if err != nil {
		return nil, mapPGError(err)
	}
	return pair, nil
}

// List returns trading pairs with optional filters.
func (r *TradingPairRepository) List(ctx context.Context, filter repositories.TradingPairFilter, opts repositories.ListOptions) ([]entities.TradingPair, error) {
	if r.pool == nil {
		return nil, errExchangeNilPool
	}

	opts = opts.WithDefaults()

	sortColumn := sanitizeTradingPairSortColumn(opts.SortBy)
	sortOrder := "ASC"
	if opts.SortOrder == repositories.SortDescending {
		sortOrder = "DESC"
	}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(tradingPairSelectColumns)
	queryBuilder.WriteString(" WHERE 1=1")

	args := []any{}
	argIndex := 1

	if filter.BaseSymbol != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND base_symbol = $%d", argIndex))
		args = append(args, *filter.BaseSymbol)
		argIndex++
	}

	if filter.QuoteSymbol != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND quote_symbol = $%d", argIndex))
		args = append(args, *filter.QuoteSymbol)
		argIndex++
	}

	if filter.IsActive != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND is_active = $%d", argIndex))
		args = append(args, *filter.IsActive)
		argIndex++
	}

	if filter.HasLiquidity != nil {
		if *filter.HasLiquidity {
			queryBuilder.WriteString(fmt.Sprintf(" AND liquidity_amount > $%d", argIndex))
			args = append(args, "0")
		} else {
			queryBuilder.WriteString(fmt.Sprintf(" AND liquidity_amount <= $%d", argIndex))
			args = append(args, "0")
		}
		argIndex++
	}

	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY %s %s", sortColumn, sortOrder))
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1))
	args = append(args, opts.Limit, opts.Offset)

	rows, err := r.pool.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, mapPGError(err)
	}
	defer rows.Close()

	results := make([]entities.TradingPair, 0)
	for rows.Next() {
		pair, scanErr := r.scanTradingPair(rows)
		if scanErr != nil {
			return nil, mapPGError(scanErr)
		}
		results = append(results, pair)
	}

	if rows.Err() != nil {
		return nil, mapPGError(rows.Err())
	}

	return results, nil
}

// GetActivePairs returns all active trading pairs.
func (r *TradingPairRepository) GetActivePairs(ctx context.Context) ([]entities.TradingPair, error) {
	if r.pool == nil {
		return nil, errExchangeNilPool
	}

	query := tradingPairSelectColumns + " WHERE is_active = true ORDER BY base_symbol, quote_symbol"
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, mapPGError(err)
	}
	defer rows.Close()

	results := make([]entities.TradingPair, 0)
	for rows.Next() {
		pair, scanErr := r.scanTradingPair(rows)
		if scanErr != nil {
			return nil, mapPGError(scanErr)
		}
		results = append(results, pair)
	}

	if rows.Err() != nil {
		return nil, mapPGError(rows.Err())
	}

	return results, nil
}

// GetPairsBySymbol returns trading pairs that contain the specified symbol.
func (r *TradingPairRepository) GetPairsBySymbol(ctx context.Context, symbol string) ([]entities.TradingPair, error) {
	if r.pool == nil {
		return nil, errExchangeNilPool
	}

	query := tradingPairSelectColumns + " WHERE base_symbol = $1 OR quote_symbol = $1 ORDER BY base_symbol, quote_symbol"
	rows, err := r.pool.Query(ctx, query, symbol)
	if err != nil {
		return nil, mapPGError(err)
	}
	defer rows.Close()

	results := make([]entities.TradingPair, 0)
	for rows.Next() {
		pair, scanErr := r.scanTradingPair(rows)
		if scanErr != nil {
			return nil, mapPGError(scanErr)
		}
		results = append(results, pair)
	}

	if rows.Err() != nil {
		return nil, mapPGError(rows.Err())
	}

	return results, nil
}

// Create persists the supplied trading pair entity.
func (r *TradingPairRepository) Create(ctx context.Context, pair *entities.TradingPairEntity) error {
	if r.pool == nil {
		return errExchangeNilPool
	}
	if pair == nil {
		return errExchangeNilTradingPair
	}

	query := `
INSERT INTO trading_pairs (
	id,
	base_symbol,
	quote_symbol,
	exchange_rate,
	inverse_rate,
	fee_percentage,
	min_swap_amount,
	max_swap_amount,
	daily_volume,
	is_active,
	has_liquidity,
	last_updated,
	created_at,
	updated_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)`

	_, err := r.pool.Exec(ctx, query,
		pair.GetID(),
		pair.GetBaseSymbol(),
		pair.GetQuoteSymbol(),
		pair.GetExchangeRate().String(),
		pair.GetInverseRate().String(),
		pair.GetFeePercentage().String(),
		pair.GetMinSwapAmount().String(),
		func() string {
			if max := pair.GetMaxSwapAmount(); max != nil {
				return max.String()
			}
			return ""
		}(),
		pair.GetDailyVolume().String(),
		pair.IsActive(),
		pair.HasLiquidity(),
		pair.GetLastUpdated().UTC(),
		pair.GetCreatedAt().UTC(),
		pair.GetUpdatedAt().UTC(),
	)
	if err != nil {
		return mapPGError(err)
	}

	return nil
}

// Update persists changes to an existing trading pair entity.
func (r *TradingPairRepository) Update(ctx context.Context, pair entities.TradingPair) error {
	if r.pool == nil {
		return errExchangeNilPool
	}
	if pair == nil {
		return errExchangeNilTradingPair
	}

	query := `
UPDATE trading_pairs
SET
	exchange_rate = $2,
	inverse_rate = $3,
	fee_percentage = $4,
	min_swap_amount = $5,
	max_swap_amount = $6,
	daily_volume = $7,
	is_active = $8,
	has_liquidity = $9,
	last_updated = $10,
	updated_at = $11
WHERE id = $1`

	cmd, err := r.pool.Exec(ctx, query,
		pair.GetID(),
		pair.GetExchangeRate().String(),
		pair.GetInverseRate().String(),
		pair.GetFeePercentage().String(),
		pair.GetMinSwapAmount().String(),
		func() string {
			if max := pair.GetMaxSwapAmount(); max != nil {
				return max.String()
			}
			return ""
		}(),
		pair.GetDailyVolume().String(),
		pair.IsActive(),
		pair.HasLiquidity(),
		pair.GetLastUpdated().UTC(),
		pair.GetUpdatedAt().UTC(),
	)
	if err != nil {
		return mapPGError(err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrNotFound
	}

	return nil
}

// Delete removes a trading pair by its identifier.
func (r *TradingPairRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if r.pool == nil {
		return errExchangeNilPool
	}

	cmd, err := r.pool.Exec(ctx, "DELETE FROM trading_pairs WHERE id = $1", id)
	if err != nil {
		return mapPGError(err)
	}

	if cmd.RowsAffected() == 0 {
		return repositories.ErrNotFound
	}

	return nil
}

// UpdateRates updates exchange rates for multiple trading pairs in a single transaction.
func (r *TradingPairRepository) UpdateRates(ctx context.Context, updates map[uuid.UUID]decimal.Decimal) error {
	if r.pool == nil {
		return errExchangeNilPool
	}

	if len(updates) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return mapPGError(err)
	}
	defer tx.Rollback(ctx)

	query := "UPDATE trading_pairs SET exchange_rate = $2, updated_at = $3 WHERE id = $1"
	now := time.Now().UTC()

	for id, rate := range updates {
		_, err := tx.Exec(ctx, query, id, rate.String(), now)
		if err != nil {
			return mapPGError(err)
		}
	}

	return tx.Commit(ctx)
}

// ResetDailyVolumes resets daily volumes for all trading pairs.
func (r *TradingPairRepository) ResetDailyVolumes(ctx context.Context) error {
	if r.pool == nil {
		return errExchangeNilPool
	}

	query := "UPDATE trading_pairs SET daily_volume = 0, updated_at = $1"
	_, err := r.pool.Exec(ctx, query, time.Now().UTC())
	if err != nil {
		return mapPGError(err)
	}

	return nil
}

// GetActiveCount returns the count of active trading pairs.
func (r *TradingPairRepository) GetActiveCount(ctx context.Context) (int64, error) {
	if r.pool == nil {
		return 0, errExchangeNilPool
	}

	var count int64
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM trading_pairs WHERE is_active = true").Scan(&count)
	if err != nil {
		return 0, mapPGError(err)
	}

	return count, nil
}

// GetTotalDailyVolume returns the total daily volume across all trading pairs.
func (r *TradingPairRepository) GetTotalDailyVolume(ctx context.Context) (decimal.Decimal, error) {
	if r.pool == nil {
		return decimal.Zero, errExchangeNilPool
	}

	var volumeStr string
	err := r.pool.QueryRow(ctx, "SELECT COALESCE(SUM(daily_volume), 0) FROM trading_pairs").Scan(&volumeStr)
	if err != nil {
		return decimal.Zero, mapPGError(err)
	}

	volume, err := decimal.NewFromString(volumeStr)
	if err != nil {
		return decimal.Zero, fmt.Errorf("exchange repository: parse total daily volume: %w", err)
	}

	return volume, nil
}

// Scan methods

func (r *ExchangeOperationRepository) scanExchangeOperation(row pgx.Row) (entities.ExchangeOperation, error) {
	var (
		id                uuid.UUID
		userID            uuid.UUID
		fromWalletID      uuid.UUID
		toWalletID        uuid.UUID
		fromAmountStr     string
		toAmountStr       string
		exchangeRateStr   string
		feePercentageStr  string
		feeAmountStr      string
		statusValue       string
		quoteExpiresAt    time.Time
		executedAt        *time.Time
		fromTransactionID *uuid.UUID
		toTransactionID   *uuid.UUID
		errorMessage      string
		createdAt         time.Time
		updatedAt         time.Time
	)

	err := row.Scan(
		&id,
		&userID,
		&fromWalletID,
		&toWalletID,
		&fromAmountStr,
		&toAmountStr,
		&exchangeRateStr,
		&feePercentageStr,
		&feeAmountStr,
		&statusValue,
		&quoteExpiresAt,
		&executedAt,
		&fromTransactionID,
		&toTransactionID,
		&errorMessage,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	fromAmount, err := decimal.NewFromString(fromAmountStr)
	if err != nil {
		return nil, fmt.Errorf("exchange repository: parse from_amount: %w", err)
	}

	toAmount, err := decimal.NewFromString(toAmountStr)
	if err != nil {
		return nil, fmt.Errorf("exchange repository: parse to_amount: %w", err)
	}

	exchangeRate, err := decimal.NewFromString(exchangeRateStr)
	if err != nil {
		return nil, fmt.Errorf("exchange repository: parse exchange_rate: %w", err)
	}

	feePercentage, err := decimal.NewFromString(feePercentageStr)
	if err != nil {
		return nil, fmt.Errorf("exchange repository: parse fee_percentage: %w", err)
	}

	feeAmount, err := decimal.NewFromString(feeAmountStr)
	if err != nil {
		return nil, fmt.Errorf("exchange repository: parse fee_amount: %w", err)
	}

	var executedAtPtr *time.Time
	if executedAt != nil {
		t := executedAt.UTC()
		executedAtPtr = &t
	}

	operation := entities.HydrateExchangeOperationEntity(entities.ExchangeOperationParams{
		ID:                id,
		UserID:            userID,
		FromWalletID:      fromWalletID,
		ToWalletID:        toWalletID,
		FromAmount:        fromAmount,
		ToAmount:          toAmount,
		ExchangeRate:      exchangeRate,
		FeePercentage:     feePercentage,
		FeeAmount:         feeAmount,
		Status:            entities.ExchangeStatus(statusValue),
		FromTransactionID: fromTransactionID,
		ToTransactionID:   toTransactionID,
		QuoteExpiresAt:    quoteExpiresAt.UTC(),
		ExecutedAt:        executedAtPtr,
		ErrorMessage:      errorMessage,
		CreatedAt:         createdAt.UTC(),
		UpdatedAt:         updatedAt.UTC(),
	})

	return operation, nil
}

func (r *TradingPairRepository) scanTradingPair(row pgx.Row) (entities.TradingPair, error) {
	var (
		id               uuid.UUID
		baseSymbol       string
		quoteSymbol      string
		exchangeRateStr  string
		inverseRateStr   string
		feePercentageStr string
		isActive         bool
		minSwapAmountStr string
		maxSwapAmountStr *string
		dailyVolumeStr   string
		hasLiquidity     bool
		lastUpdated      time.Time
		createdAt        time.Time
		updatedAt        time.Time
	)

	err := row.Scan(
		&id,
		&baseSymbol,
		&quoteSymbol,
		&exchangeRateStr,
		&inverseRateStr,
		&feePercentageStr,
		&isActive,
		&minSwapAmountStr,
		&maxSwapAmountStr,
		&dailyVolumeStr,
		&hasLiquidity,
		&lastUpdated,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	exchangeRate, err := decimal.NewFromString(exchangeRateStr)
	if err != nil {
		return nil, fmt.Errorf("exchange repository: parse exchange_rate: %w", err)
	}

	inverseRate, err := decimal.NewFromString(inverseRateStr)
	if err != nil {
		return nil, fmt.Errorf("exchange repository: parse inverse_rate: %w", err)
	}

	feePercentage, err := decimal.NewFromString(feePercentageStr)
	if err != nil {
		return nil, fmt.Errorf("exchange repository: parse fee_percentage: %w", err)
	}

	minSwapAmount, err := decimal.NewFromString(minSwapAmountStr)
	if err != nil {
		return nil, fmt.Errorf("exchange repository: parse min_swap_amount: %w", err)
	}

	var maxSwapAmount *decimal.Decimal
	if maxSwapAmountStr != nil && *maxSwapAmountStr != "" {
		max, err := decimal.NewFromString(*maxSwapAmountStr)
		if err != nil {
			return nil, fmt.Errorf("exchange repository: parse max_swap_amount: %w", err)
		}
		maxSwapAmount = &max
	}

	dailyVolume, err := decimal.NewFromString(dailyVolumeStr)
	if err != nil {
		return nil, fmt.Errorf("exchange repository: parse daily_volume: %w", err)
	}

	pair := entities.HydrateTradingPairEntity(entities.TradingPairParams{
		ID:            id,
		BaseSymbol:    baseSymbol,
		QuoteSymbol:   quoteSymbol,
		ExchangeRate:  exchangeRate,
		InverseRate:   inverseRate,
		FeePercentage: feePercentage,
		MinSwapAmount: minSwapAmount,
		MaxSwapAmount: maxSwapAmount,
		DailyVolume:   dailyVolume,
		IsActive:      isActive,
		HasLiquidity:  hasLiquidity,
		LastUpdated:   lastUpdated,
		CreatedAt:     createdAt.UTC(),
		UpdatedAt:     updatedAt.UTC(),
	})

	return pair, nil
}

// Helper functions

func sanitizeExchangeOperationSortColumn(sortBy string) string {
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "from_amount":
		return "from_amount"
	case "to_amount":
		return "to_amount"
	case "exchange_rate":
		return "exchange_rate"
	case "status":
		return "status"
	case "quote_expires_at":
		return "quote_expires_at"
	case "completed_at":
		return "completed_at"
	case "created_at":
		return "created_at"
	default:
		return "created_at"
	}
}

func sanitizeTradingPairSortColumn(sortBy string) string {
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "base_symbol":
		return "base_symbol"
	case "quote_symbol":
		return "quote_symbol"
	case "exchange_rate":
		return "exchange_rate"
	case "fee_percentage":
		return "fee_percentage"
	case "daily_volume":
		return "daily_volume"
	case "liquidity_amount":
		return "liquidity_amount"
	case "created_at":
		return "created_at"
	default:
		return "base_symbol"
	}
}
