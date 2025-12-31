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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
)

const walletSelectColumns = `
SELECT
	id,
	user_id,
	chain,
	address,
	encrypted_private_key,
	derivation_path,
	label,
	balance,
	balance_updated_at,
	status,
	created_at,
	updated_at
FROM wallets`

var (
	errNilPool   = errors.New("wallet repository: database pool is not configured")
	errNilWallet = errors.New("wallet repository: wallet entity is required")
)

// WalletRepository persists wallet aggregates using PostgreSQL.
type WalletRepository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewWalletRepository constructs a WalletRepository backed by the provided pool.
func NewWalletRepository(pool *pgxpool.Pool, logger *slog.Logger) *WalletRepository {
	if logger == nil {
		logger = slog.Default()
	}
	return &WalletRepository{
		pool:   pool,
		logger: logger,
	}
}

// GetByID returns a wallet matching the supplied identifier.
func (r *WalletRepository) GetByID(ctx context.Context, id uuid.UUID) (entities.Wallet, error) {
	if r.pool == nil {
		return nil, errNilPool
	}

	row := r.pool.QueryRow(ctx, walletSelectColumns+" WHERE id = $1", id)
	wallet, err := r.scanWallet(row)
	if err != nil {
		return nil, mapPGError(err)
	}
	return wallet, nil
}

// GetByAddress returns a wallet that matches the address and chain.
func (r *WalletRepository) GetByAddress(ctx context.Context, chain entities.Chain, address string) (entities.Wallet, error) {
	if r.pool == nil {
		return nil, errNilPool
	}

	row := r.pool.QueryRow(ctx, walletSelectColumns+" WHERE chain = $1 AND address = $2", string(chain), address)
	wallet, err := r.scanWallet(row)
	if err != nil {
		return nil, mapPGError(err)
	}
	return wallet, nil
}

// ListByUser returns wallets belonging to the specified user with optional filters.
func (r *WalletRepository) ListByUser(ctx context.Context, userID uuid.UUID, filter repositories.WalletFilter, opts repositories.ListOptions) ([]entities.Wallet, error) {
	if r.pool == nil {
		return nil, errNilPool
	}

	opts = opts.WithDefaults()

	sortColumn := sanitizeWalletSortColumn(opts.SortBy)
	sortOrder := "DESC"
	if opts.SortOrder == repositories.SortAscending {
		sortOrder = "ASC"
	}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(walletSelectColumns)
	queryBuilder.WriteString(" WHERE user_id = $1")

	args := []any{userID}
	argIndex := 2

	if filter.Chain != nil && *filter.Chain != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND chain = $%d", argIndex))
		args = append(args, string(*filter.Chain))
		argIndex++
	}

	if filter.Status != nil && *filter.Status != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND status = $%d", argIndex))
		args = append(args, string(*filter.Status))
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

	results := make([]entities.Wallet, 0)
	for rows.Next() {
		wallet, scanErr := r.scanWallet(rows)
		if scanErr != nil {
			return nil, mapPGError(scanErr)
		}
		results = append(results, wallet)
	}

	if rows.Err() != nil {
		return nil, mapPGError(rows.Err())
	}

	return results, nil
}

// Create persists the supplied wallet entity.
func (r *WalletRepository) Create(ctx context.Context, wallet *entities.WalletEntity) error {
	if r.pool == nil {
		return errNilPool
	}
	if wallet == nil {
		return errNilWallet
	}

	query := `
INSERT INTO wallets (
	id,
	user_id,
	chain,
	address,
	encrypted_private_key,
	derivation_path,
	label,
	balance,
	balance_updated_at,
	status,
	created_at,
	updated_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)`

	balanceStr := wallet.GetBalance().String()
	var balanceUpdatedAt any
	if ts := wallet.GetBalanceUpdatedAt(); ts != nil {
		balanceUpdatedAt = ts.UTC()
	}

	_, err := r.pool.Exec(ctx, query,
		wallet.GetID(),
		wallet.GetUserID(),
		string(wallet.GetChain()),
		wallet.GetAddress(),
		wallet.GetEncryptedPrivateKey(),
		nullIfEmpty(wallet.GetDerivationPath()),
		nullIfEmpty(wallet.GetLabel()),
		balanceStr,
		balanceUpdatedAt,
		string(wallet.GetStatus()),
		wallet.GetCreatedAt().UTC(),
		wallet.GetUpdatedAt().UTC(),
	)
	if err != nil {
		return mapPGError(err)
	}

	return nil
}

// Update persists changes to an existing wallet entity.
func (r *WalletRepository) Update(ctx context.Context, wallet entities.Wallet) error {
	if r.pool == nil {
		return errNilPool
	}
	if wallet == nil {
		return errNilWallet
	}

	query := `
UPDATE wallets
SET
	encrypted_private_key = $2,
	derivation_path = $3,
	label = $4,
	balance = $5,
	balance_updated_at = $6,
	status = $7,
	updated_at = $8
WHERE id = $1`

	var balanceUpdatedAt any
	if ts := wallet.GetBalanceUpdatedAt(); ts != nil {
		balanceUpdatedAt = ts.UTC()
	}

	cmd, err := r.pool.Exec(ctx, query,
		wallet.GetID(),
		wallet.GetEncryptedPrivateKey(),
		nullIfEmpty(wallet.GetDerivationPath()),
		nullIfEmpty(wallet.GetLabel()),
		wallet.GetBalance().String(),
		balanceUpdatedAt,
		string(wallet.GetStatus()),
		wallet.GetUpdatedAt().UTC(),
	)
	if err != nil {
		return mapPGError(err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrNotFound
	}

	return nil
}

// Delete removes a wallet by its identifier.
func (r *WalletRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if r.pool == nil {
		return errNilPool
	}

	cmd, err := r.pool.Exec(ctx, "DELETE FROM wallets WHERE id = $1", id)
	if err != nil {
		return mapPGError(err)
	}

	if cmd.RowsAffected() == 0 {
		return repositories.ErrNotFound
	}

	return nil
}

func (r *WalletRepository) scanWallet(row pgx.Row) (entities.Wallet, error) {
	var (
		id                 uuid.UUID
		userID             uuid.UUID
		chainValue         string
		address            string
		encryptedKey       string
		derivationPathText pgtype.Text
		labelText          pgtype.Text
		balanceNumeric     string
		balanceUpdatedAt   pgtype.Timestamptz
		statusValue        string
		createdAt          time.Time
		updatedAt          time.Time
	)

	err := row.Scan(
		&id,
		&userID,
		&chainValue,
		&address,
		&encryptedKey,
		&derivationPathText,
		&labelText,
		&balanceNumeric,
		&balanceUpdatedAt,
		&statusValue,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	balance := decimal.Zero
	if strings.TrimSpace(balanceNumeric) != "" {
		dec, decErr := decimal.NewFromString(balanceNumeric)
		if decErr != nil {
			return nil, fmt.Errorf("wallet repository: parse balance: %w", decErr)
		}
		balance = dec
	}

	var derivationPath string
	if derivationPathText.Valid {
		derivationPath = derivationPathText.String
	}

	label := ""
	if labelText.Valid {
		label = labelText.String
	}

	var balanceAt *time.Time
	if balanceUpdatedAt.Valid {
		t := balanceUpdatedAt.Time.UTC()
		balanceAt = &t
	}

	wallet := entities.HydrateWalletEntity(entities.WalletParams{
		ID:                  id,
		UserID:              userID,
		Chain:               entities.Chain(chainValue),
		Address:             address,
		EncryptedPrivateKey: encryptedKey,
		DerivationPath:      derivationPath,
		Label:               label,
		Balance:             balance,
		BalanceUpdatedAt:    balanceAt,
		Status:              entities.WalletStatus(statusValue),
		CreatedAt:           createdAt.UTC(),
		UpdatedAt:           updatedAt.UTC(),
	})

	return wallet, nil
}

func sanitizeWalletSortColumn(sortBy string) string {
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "label":
		return "label"
	case "chain":
		return "chain"
	case "balance":
		return "balance"
	case "created_at":
		return "created_at"
	default:
		return "created_at"
	}
}

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func mapPGError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return repositories.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return repositories.ErrDuplicate
		case "23503":
			return fmt.Errorf("wallet repository: foreign key violation: %w", err)
		default:
			return fmt.Errorf("wallet repository: db error (%s): %w", pgErr.Code, err)
		}
	}

	return err
}
