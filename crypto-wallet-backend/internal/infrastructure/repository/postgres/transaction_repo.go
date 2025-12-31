package postgres

import (
    "context"
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgconn"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/shopspring/decimal"

    "github.com/crypto-wallet/backend/internal/domain/entities"
    "github.com/crypto-wallet/backend/internal/domain/repositories"
)

const selectTransactionBase = `
SELECT
    id,
    wallet_id,
    chain,
    tx_hash,
    type,
    amount,
    fee,
    status,
    from_address,
    to_address,
    block_number,
    confirmations,
    error_message,
    metadata,
    created_at,
    confirmed_at,
    updated_at
FROM transactions
`

// PostgresTransactionRepository persists transactions in PostgreSQL.
type PostgresTransactionRepository struct {
    pool *pgxpool.Pool
}

// NewPostgresTransactionRepository constructs the repository.
func NewPostgresTransactionRepository(pool *pgxpool.Pool) *PostgresTransactionRepository {
    return &PostgresTransactionRepository{pool: pool}
}

// GetByID retrieves a transaction by primary key.
func (r *PostgresTransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (entities.Transaction, error) {
    row := r.pool.QueryRow(ctx, selectTransactionBase+" WHERE id = $1", id)
    return scanTransaction(row)
}

// GetByHash retrieves a transaction by chain/hash combination.
func (r *PostgresTransactionRepository) GetByHash(ctx context.Context, chain entities.Chain, hash string) (entities.Transaction, error) {
    row := r.pool.QueryRow(ctx, selectTransactionBase+" WHERE chain = $1 AND tx_hash = $2", chain, hash)
    return scanTransaction(row)
}

// ListByWallet retrieves paginated transactions for a wallet.
func (r *PostgresTransactionRepository) ListByWallet(ctx context.Context, walletID uuid.UUID, opts repositories.ListOptions) ([]entities.Transaction, error) {
    options := opts.WithDefaults()

    sortColumn := "created_at"
    if strings.EqualFold(options.SortBy, "confirmations") {
        sortColumn = "confirmations"
    }

    sortDirection := strings.ToUpper(string(options.SortOrder))
    if sortDirection != "ASC" && sortDirection != "DESC" {
        sortDirection = "DESC"
    }

    query := fmt.Sprintf("%s WHERE wallet_id = $1 ORDER BY %s %s LIMIT $2 OFFSET $3", selectTransactionBase, sortColumn, sortDirection)

    rows, err := r.pool.Query(ctx, query, walletID, options.Limit, options.Offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var results []entities.Transaction
    for rows.Next() {
        tx, scanErr := scanTransaction(rows)
        if scanErr != nil {
            return nil, scanErr
        }
        results = append(results, tx)
    }

    if rows.Err() != nil {
        return nil, rows.Err()
    }

    return results, nil
}

// ListWithFilters returns transactions filtered by multiple attributes with pagination.
func (r *PostgresTransactionRepository) ListWithFilters(ctx context.Context, filter repositories.TransactionFilter, opts repositories.ListOptions) ([]entities.Transaction, int64, error) {
    if r.pool == nil {
        return nil, 0, errors.New("transaction repository: database pool is not configured")
    }

    opts = opts.WithDefaults()

    conditions := make([]string, 0, 6)
    args := make([]any, 0, 6)

    if filter.WalletID != nil {
        conditions = append(conditions, fmt.Sprintf("wallet_id = $%d", len(args)+1))
        args = append(args, *filter.WalletID)
    }

    if filter.Chain != nil && *filter.Chain != "" {
        conditions = append(conditions, fmt.Sprintf("chain = $%d", len(args)+1))
        args = append(args, string(*filter.Chain))
    }

    if filter.Type != nil && *filter.Type != "" {
        conditions = append(conditions, fmt.Sprintf("type = $%d", len(args)+1))
        args = append(args, string(*filter.Type))
    }

    if filter.Status != nil && *filter.Status != "" {
        conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)+1))
        args = append(args, string(*filter.Status))
    }

    if filter.StartDate != nil {
        conditions = append(conditions, fmt.Sprintf("created_at >= $%d", len(args)+1))
        args = append(args, filter.StartDate.UTC())
    }

    if filter.EndDate != nil {
        conditions = append(conditions, fmt.Sprintf("created_at <= $%d", len(args)+1))
        args = append(args, filter.EndDate.UTC())
    }

    whereClause := ""
    if len(conditions) > 0 {
        whereClause = " WHERE " + strings.Join(conditions, " AND ")
    }

    countQuery := "SELECT COUNT(*) FROM transactions" + whereClause
    var total int64
    if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
        return nil, 0, err
    }

    sortColumn := sanitizeTransactionSortColumn(opts.SortBy)
    sortOrder := strings.ToUpper(string(opts.SortOrder))
    if sortOrder != "ASC" {
        sortOrder = "DESC"
    }

    limitPlaceholder := len(args) + 1
    offsetPlaceholder := len(args) + 2

    query := fmt.Sprintf("%s%s ORDER BY %s %s LIMIT $%d OFFSET $%d", selectTransactionBase, whereClause, sortColumn, sortOrder, limitPlaceholder, offsetPlaceholder)
    queryArgs := append(args, opts.Limit, opts.Offset)

    rows, err := r.pool.Query(ctx, query, queryArgs...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    results := make([]entities.Transaction, 0, opts.Limit)
    for rows.Next() {
        tx, scanErr := scanTransaction(rows)
        if scanErr != nil {
            return nil, 0, scanErr
        }
        results = append(results, tx)
    }

    if rows.Err() != nil {
        return nil, 0, rows.Err()
    }

    return results, total, nil
}

// ListPending returns transactions awaiting confirmations for monitoring workers.
func (r *PostgresTransactionRepository) ListPending(ctx context.Context, chain entities.Chain, limit int) ([]entities.Transaction, error) {
    if limit <= 0 {
        limit = 100
    }

    query := selectTransactionBase + " WHERE chain = $1 AND status IN ('pending','confirming') ORDER BY created_at ASC LIMIT $2"
    rows, err := r.pool.Query(ctx, query, chain, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var results []entities.Transaction
    for rows.Next() {
        tx, scanErr := scanTransaction(rows)
        if scanErr != nil {
            return nil, scanErr
        }
        results = append(results, tx)
    }
    if rows.Err() != nil {
        return nil, rows.Err()
    }

    return results, nil
}

// Create inserts a new transaction record.
func (r *PostgresTransactionRepository) Create(ctx context.Context, tx *entities.TransactionEntity) error {
    if tx == nil {
        return errors.New("transaction entity is nil")
    }

    metadataJSON, err := json.Marshal(tx.GetMetadata())
    if err != nil {
        return err
    }

    query := `
INSERT INTO transactions (
    id,
    wallet_id,
    chain,
    tx_hash,
    type,
    amount,
    fee,
    status,
    from_address,
    to_address,
    block_number,
    confirmations,
    error_message,
    metadata,
    created_at,
    confirmed_at,
    updated_at
) VALUES (
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17
)
`

    _, err = r.pool.Exec(
        ctx,
        query,
        tx.GetID(),
        tx.GetWalletID(),
        tx.GetChain(),
        tx.GetHash(),
        tx.GetType(),
        tx.GetAmount().String(),
        tx.GetFee().String(),
        tx.GetStatus(),
        tx.GetFromAddress(),
        tx.GetToAddress(),
        nullableUint64(tx.GetBlockNumber()),
        tx.GetConfirmations(),
        tx.GetErrorMessage(),
        metadataJSON,
        tx.GetCreatedAt(),
        tx.GetConfirmedAt(),
        tx.GetUpdatedAt(),
    )
    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23505" {
            return repositories.ErrDuplicate
        }
        return err
    }

    return nil
}

// Update persists transaction status changes.
func (r *PostgresTransactionRepository) Update(ctx context.Context, tx entities.Transaction) error {
    if tx == nil {
        return errors.New("transaction entity is nil")
    }

    metadataJSON, err := json.Marshal(tx.GetMetadata())
    if err != nil {
        return err
    }

    query := `
UPDATE transactions SET
    chain = $1,
    tx_hash = $2,
    type = $3,
    amount = $4,
    fee = $5,
    status = $6,
    from_address = $7,
    to_address = $8,
    block_number = $9,
    confirmations = $10,
    error_message = $11,
    metadata = $12,
    confirmed_at = $13,
    updated_at = $14
WHERE id = $15
`

    cmd, err := r.pool.Exec(
        ctx,
        query,
        tx.GetChain(),
        tx.GetHash(),
        tx.GetType(),
        tx.GetAmount().String(),
        tx.GetFee().String(),
        tx.GetStatus(),
        tx.GetFromAddress(),
        tx.GetToAddress(),
        nullableUint64(tx.GetBlockNumber()),
        tx.GetConfirmations(),
        tx.GetErrorMessage(),
        metadataJSON,
        tx.GetConfirmedAt(),
        time.Now().UTC(),
        tx.GetID(),
    )
    if err != nil {
        return err
    }
    if cmd.RowsAffected() == 0 {
        return repositories.ErrNotFound
    }

    return nil
}

func scanTransaction(row pgx.Row) (entities.Transaction, error) {
    var (
        id uuid.UUID
        walletID uuid.UUID
        chain string
        hash string
        txType string
        amountStr string
        feeStr string
        status string
        fromAddress string
        toAddress string
        blockNumber sql.NullInt64
        confirmations int
        errorMessage sql.NullString
        metadataBytes []byte
        createdAt time.Time
        confirmedAt sql.NullTime
        updatedAt time.Time
    )

    if err := row.Scan(
        &id,
        &walletID,
        &chain,
        &hash,
        &txType,
        &amountStr,
        &feeStr,
        &status,
        &fromAddress,
        &toAddress,
        &blockNumber,
        &confirmations,
        &errorMessage,
        &metadataBytes,
        &createdAt,
        &confirmedAt,
        &updatedAt,
    ); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, repositories.ErrNotFound
        }
        return nil, err
    }

    amount, err := decimal.NewFromString(amountStr)
    if err != nil {
        return nil, fmt.Errorf("parse amount: %w", err)
    }
    fee, err := decimal.NewFromString(feeStr)
    if err != nil {
        return nil, fmt.Errorf("parse fee: %w", err)
    }

    metadata := map[string]any{}
    if len(metadataBytes) > 0 {
        if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
            return nil, fmt.Errorf("parse metadata: %w", err)
        }
    }

    params := entities.TransactionParams{
        ID:            id,
        WalletID:      walletID,
        Chain:         entities.Chain(chain),
        Hash:          hash,
        Type:          entities.TransactionType(txType),
        Amount:        amount,
        Fee:           fee,
        Status:        entities.TransactionStatus(status),
        FromAddress:   fromAddress,
        ToAddress:     toAddress,
        BlockNumber:   nullableUint64FromSQL(blockNumber),
        Confirmations: confirmations,
        ErrorMessage:  errorMessage.String,
        Metadata:      metadata,
        CreatedAt:     createdAt,
        ConfirmedAt:   nullableTimePtr(confirmedAt),
        UpdatedAt:     updatedAt,
    }

    return entities.HydrateTransactionEntity(params), nil
}

func nullableUint64(value *uint64) any {
    if value == nil {
        return nil
    }
    return *value
}

func nullableUint64FromSQL(value sql.NullInt64) *uint64 {
    if !value.Valid {
        return nil
    }
    v := uint64(value.Int64)
    return &v
}

func nullableTimePtr(value sql.NullTime) *time.Time {
    if !value.Valid {
        return nil
    }
    t := value.Time
    return &t
}

func sanitizeTransactionSortColumn(sortBy string) string {
    switch strings.ToLower(strings.TrimSpace(sortBy)) {
    case "amount":
        return "amount"
    case "chain":
        return "chain"
    case "status":
        return "status"
    case "confirmations":
        return "confirmations"
    case "created_at":
        fallthrough
    default:
        return "created_at"
    }
}
