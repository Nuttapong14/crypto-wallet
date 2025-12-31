package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	errAccountUserIDRequired  = errors.New("account user ID is required")
	errAccountBalanceNegative = errors.New("account total balance cannot be negative")
)

// Account models the aggregate portfolio view for a user.
type Account interface {
	Entity
	Identifiable
	Timestamped

	GetUserID() uuid.UUID
	GetTotalBalanceUSD() decimal.Decimal
	GetLastCalculatedAt() *time.Time
	UpdateTotalBalance(amount decimal.Decimal, at time.Time) error
}

// AccountEntity is the concrete implementation of Account.
type AccountEntity struct {
	id               uuid.UUID
	userID           uuid.UUID
	totalBalanceUSD  decimal.Decimal
	lastCalculatedAt *time.Time
	createdAt        time.Time
	updatedAt        time.Time
}

// AccountParams captures the fields required to construct an AccountEntity.
type AccountParams struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	TotalBalanceUSD  decimal.Decimal
	LastCalculatedAt *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewAccountEntity validates the supplied parameters and returns a new AccountEntity instance.
func NewAccountEntity(params AccountParams) (*AccountEntity, error) {
	if params.ID == uuid.Nil {
		params.ID = uuid.New()
	}
	if params.CreatedAt.IsZero() {
		params.CreatedAt = time.Now().UTC()
	}
	if params.UpdatedAt.IsZero() {
		params.UpdatedAt = params.CreatedAt
	}

	entity := &AccountEntity{
		id:               params.ID,
		userID:           params.UserID,
		totalBalanceUSD:  params.TotalBalanceUSD,
		lastCalculatedAt: params.LastCalculatedAt,
		createdAt:        params.CreatedAt,
		updatedAt:        params.UpdatedAt,
	}

	if err := entity.Validate(); err != nil {
		return nil, err
	}

	return entity, nil
}

// HydrateAccountEntity builds an AccountEntity without executing validation (used for persistence hydration).
func HydrateAccountEntity(params AccountParams) *AccountEntity {
	return &AccountEntity{
		id:               params.ID,
		userID:           params.UserID,
		totalBalanceUSD:  params.TotalBalanceUSD,
		lastCalculatedAt: params.LastCalculatedAt,
		createdAt:        params.CreatedAt,
		updatedAt:        params.UpdatedAt,
	}
}

// Validate ensures the account entity adheres to domain invariants.
func (a *AccountEntity) Validate() error {
	var validationErr error

	if a.userID == uuid.Nil {
		validationErr = errors.Join(validationErr, errAccountUserIDRequired)
	}
	if a.totalBalanceUSD.IsNegative() {
		validationErr = errors.Join(validationErr, errAccountBalanceNegative)
	}

	return validationErr
}

// GetID returns the aggregate identifier.
func (a *AccountEntity) GetID() uuid.UUID {
	return a.id
}

// GetUserID returns the owning user identifier.
func (a *AccountEntity) GetUserID() uuid.UUID {
	return a.userID
}

// GetTotalBalanceUSD returns the cached total balance in USD.
func (a *AccountEntity) GetTotalBalanceUSD() decimal.Decimal {
	return a.totalBalanceUSD
}

// GetLastCalculatedAt returns the last calculation timestamp.
func (a *AccountEntity) GetLastCalculatedAt() *time.Time {
	return a.lastCalculatedAt
}

// GetCreatedAt returns the creation timestamp.
func (a *AccountEntity) GetCreatedAt() time.Time {
	return a.createdAt
}

// GetUpdatedAt returns the last update timestamp.
func (a *AccountEntity) GetUpdatedAt() time.Time {
	return a.updatedAt
}

// UpdateTotalBalance replaces the cached balance and updates timestamps.
func (a *AccountEntity) UpdateTotalBalance(amount decimal.Decimal, at time.Time) error {
	if amount.IsNegative() {
		return errAccountBalanceNegative
	}
	a.totalBalanceUSD = amount
	if at.IsZero() {
		now := time.Now().UTC()
		a.lastCalculatedAt = &now
	} else {
		t := at.UTC()
		a.lastCalculatedAt = &t
	}
	a.Touch(at)
	return nil
}

// Touch updates the modification timestamp.
func (a *AccountEntity) Touch(at time.Time) {
	if at.IsZero() {
		a.updatedAt = time.Now().UTC()
		return
	}
	a.updatedAt = at.UTC()
}
