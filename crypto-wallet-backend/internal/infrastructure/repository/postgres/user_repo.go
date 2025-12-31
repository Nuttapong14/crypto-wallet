package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
)

const selectUserBase = `
SELECT
	id,
	email,
	password_hash,
	first_name,
	last_name,
	phone_number,
	status,
	preferred_currency,
	two_factor_enabled,
	ttwo_factor_secret,
	email_verified,
	email_verified_at,
	last_login_at,
	created_at,
	updated_at
FROM users
`

var allowedUserSortColumns = map[string]string{
	"created_at": "created_at",
	"email":      "email",
	"status":     "status",
	"last_login": "last_login_at",
}

// PostgresUserRepository implements repositories.UserRepository for PostgreSQL.
type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresUserRepository returns a PostgreSQL-backed user repository.
func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	row := r.pool.QueryRow(ctx, selectUserBase+" WHERE id = $1", id)
	return scanUser(row)
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (entities.User, error) {
	row := r.pool.QueryRow(ctx, selectUserBase+" WHERE LOWER(email) = LOWER($1)", email)
	return scanUser(row)
}

func (r *PostgresUserRepository) List(ctx context.Context, opts repositories.ListOptions) ([]entities.User, error) {
	options := opts.WithDefaults()
	sortColumn, ok := allowedUserSortColumns[strings.ToLower(options.SortBy)]
	if !ok {
		sortColumn = "created_at"
	}

	sortDirection := strings.ToUpper(string(options.SortOrder))
	if sortDirection != "ASC" && sortDirection != "DESC" {
		sortDirection = "DESC"
	}

	query := fmt.Sprintf("%s ORDER BY %s %s LIMIT $1 OFFSET $2", selectUserBase, sortColumn, sortDirection)
	rows, err := r.pool.Query(ctx, query, options.Limit, options.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entities.User
	for rows.Next() {
		user, scanErr := scanUser(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		users = append(users, user)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return users, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *entities.UserEntity) error {
	if user == nil {
		return errors.New("repository: user entity is nil")
	}

	query := `
INSERT INTO users (
	id,
	email,
	password_hash,
	first_name,
	last_name,
	phone_number,
	status,
	preferred_currency,
	two_factor_enabled,
	two_factor_secret,
	email_verified,
	email_verified_at,
	last_login_at,
	created_at,
	updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15
)
`

	_, err := r.pool.Exec(
		ctx,
		query,
		user.GetID(),
		user.GetEmail(),
		user.GetPasswordHash(),
		user.GetFirstName(),
		user.GetLastName(),
		user.GetPhoneNumber(),
		user.GetStatus(),
		user.GetPreferredCurrency(),
		user.IsTwoFactorEnabled(),
		nullableString(user.GetTwoFactorSecret()),
		user.IsEmailVerified(),
		user.GetEmailVerifiedAt(),
		user.GetLastLoginAt(),
		user.GetCreatedAt(),
		user.GetUpdatedAt(),
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

func (r *PostgresUserRepository) Update(ctx context.Context, user entities.User) error {
	query := `
UPDATE users SET
	email = $1,
	password_hash = $2,
	first_name = $3,
	last_name = $4,
	phone_number = $5,
	status = $6,
	preferred_currency = $7,
	two_factor_enabled = $8,
	ttwo_factor_secret = $9,
	email_verified = $10,
	email_verified_at = $11,
	last_login_at = $12,
	updated_at = $13
WHERE id = $14
`

	cmd, err := r.pool.Exec(
		ctx,
		query,
		user.GetEmail(),
		user.GetPasswordHash(),
		user.GetFirstName(),
		user.GetLastName(),
		user.GetPhoneNumber(),
		user.GetStatus(),
		user.GetPreferredCurrency(),
		user.IsTwoFactorEnabled(),
		nullableString(user.GetTwoFactorSecret()),
		user.IsEmailVerified(),
		user.GetEmailVerifiedAt(),
		user.GetLastLoginAt(),
		time.Now().UTC(),
		user.GetID(),
	)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrNotFound
	}

	return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	cmd, err := r.pool.Exec(ctx, "UPDATE users SET status = 'deleted', updated_at = $1 WHERE id = $2", time.Now().UTC(), id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

func scanUser(row pgx.Row) (entities.User, error) {
	var (
		id              uuid.UUID
		email           string
		passwordHash    string
		firstName       sql.NullString
		lastName        sql.NullString
		phone           sql.NullString
		status          string
		currency        string
		twoFactor       bool
		twoFactorSecret sql.NullString
		emailVerified   bool
		emailVerifiedAt sql.NullTime
		lastLoginAt     sql.NullTime
		createdAt       time.Time
		updatedAt       time.Time
	)

	err := row.Scan(
		&id,
		&email,
		&passwordHash,
		&firstName,
		&lastName,
		&phone,
		&status,
		&currency,
		&twoFactor,
		&twoFactorSecret,
		&emailVerified,
		&emailVerifiedAt,
		&lastLoginAt,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}

	params := entities.UserParams{
		ID:                id,
		Email:             email,
		PasswordHash:      passwordHash,
		FirstName:         firstName.String,
		LastName:          lastName.String,
		PhoneNumber:       phone.String,
		Status:            entities.UserStatus(status),
		PreferredCurrency: entities.CurrencyCode(currency),
		TwoFactorEnabled:  twoFactor,
		TwoFactorSecret:   twoFactorSecret.String,
		EmailVerified:     emailVerified,
		EmailVerifiedAt:   nullableTimePtr(emailVerifiedAt),
		LastLoginAt:       nullableTimePtr(lastLoginAt),
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}

	return entities.HydrateUserEntity(params), nil
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
