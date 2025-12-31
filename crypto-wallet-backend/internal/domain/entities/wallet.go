package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// WalletStatus represents the lifecycle status for a wallet.
type WalletStatus string

const (
	WalletStatusActive   WalletStatus = "active"
	WalletStatusArchived WalletStatus = "archived"
)

var (
	errWalletUserIDRequired       = errors.New("wallet user ID is required")
	errWalletAddressRequired      = errors.New("wallet address is required")
	errWalletEncryptedKeyRequired = errors.New("wallet encrypted private key is required")
	errWalletChainInvalid         = errors.New("wallet chain is invalid")
	errWalletStatusInvalid        = errors.New("wallet status is invalid")
	errWalletBalanceNegative      = errors.New("wallet balance cannot be negative")
)

// Wallet exposes the behavior required by the application layer when working with wallet entities.
type Wallet interface {
	Entity
	Identifiable
	Timestamped

	GetUserID() uuid.UUID
	GetChain() Chain
	GetAddress() string
	GetEncryptedPrivateKey() string
	GetDerivationPath() string
	GetLabel() string
	GetBalance() decimal.Decimal
	GetBalanceUpdatedAt() *time.Time
	GetStatus() WalletStatus
	UpdateBalance(amount decimal.Decimal, at time.Time) error
	SetStatus(status WalletStatus) error
	Rename(label string)
	Touch(at time.Time)
}

// WalletEntity is the default implementation of the Wallet interface.
type WalletEntity struct {
	id                  uuid.UUID
	userID              uuid.UUID
	chain               Chain
	address             string
	encryptedPrivateKey string
	derivationPath      string
	label               string
	balance             decimal.Decimal
	balanceUpdatedAt    *time.Time
	status              WalletStatus
	createdAt           time.Time
	updatedAt           time.Time
}

// WalletParams captures the fields required to construct a WalletEntity.
type WalletParams struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	Chain               Chain
	Address             string
	EncryptedPrivateKey string
	DerivationPath      string
	Label               string
	Balance             decimal.Decimal
	BalanceUpdatedAt    *time.Time
	Status              WalletStatus
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// NewWalletEntity validates the supplied parameters and returns a new WalletEntity instance.
func NewWalletEntity(params WalletParams) (*WalletEntity, error) {
	if params.ID == uuid.Nil {
		params.ID = uuid.New()
	}

	if params.CreatedAt.IsZero() {
		params.CreatedAt = time.Now().UTC()
	}

	if params.UpdatedAt.IsZero() {
		params.UpdatedAt = params.CreatedAt
	}

	if params.Status == "" {
		params.Status = WalletStatusActive
	}

	entity := &WalletEntity{
		id:                  params.ID,
		userID:              params.UserID,
		chain:               params.Chain,
		address:             strings.TrimSpace(params.Address),
		encryptedPrivateKey: strings.TrimSpace(params.EncryptedPrivateKey),
		derivationPath:      strings.TrimSpace(params.DerivationPath),
		label:               strings.TrimSpace(params.Label),
		balance:             params.Balance,
		balanceUpdatedAt:    params.BalanceUpdatedAt,
		status:              params.Status,
		createdAt:           params.CreatedAt,
		updatedAt:           params.UpdatedAt,
	}

	if entity.balance.IsZero() && params.Balance.IsZero() {
		entity.balance = decimal.Zero
	}

	if err := entity.Validate(); err != nil {
		return nil, err
	}

	return entity, nil
}

// HydrateWalletEntity creates a WalletEntity without re-validating invariants (used for repository hydration).
func HydrateWalletEntity(params WalletParams) *WalletEntity {
	return &WalletEntity{
		id:                  params.ID,
		userID:              params.UserID,
		chain:               params.Chain,
		address:             strings.TrimSpace(params.Address),
		encryptedPrivateKey: strings.TrimSpace(params.EncryptedPrivateKey),
		derivationPath:      strings.TrimSpace(params.DerivationPath),
		label:               strings.TrimSpace(params.Label),
		balance:             params.Balance,
		balanceUpdatedAt:    params.BalanceUpdatedAt,
		status:              params.Status,
		createdAt:           params.CreatedAt,
		updatedAt:           params.UpdatedAt,
	}
}

// Validate ensures the entity adheres to domain invariants.
func (w *WalletEntity) Validate() error {
	var validationErr error

	if w.userID == uuid.Nil {
		validationErr = errors.Join(validationErr, errWalletUserIDRequired)
	}

	if strings.TrimSpace(w.address) == "" {
		validationErr = errors.Join(validationErr, errWalletAddressRequired)
	}

	if strings.TrimSpace(w.encryptedPrivateKey) == "" {
		validationErr = errors.Join(validationErr, errWalletEncryptedKeyRequired)
	}

	if !IsSupportedChain(w.chain) {
		validationErr = errors.Join(validationErr, errWalletChainInvalid)
	}

	if !isValidWalletStatus(w.status) {
		validationErr = errors.Join(validationErr, errWalletStatusInvalid)
	}

	if w.balance.IsNegative() {
		validationErr = errors.Join(validationErr, errWalletBalanceNegative)
	}

	return validationErr
}

// Getter implementations satisfy the Wallet interface.

func (w *WalletEntity) GetID() uuid.UUID {
	return w.id
}

func (w *WalletEntity) GetUserID() uuid.UUID {
	return w.userID
}

func (w *WalletEntity) GetChain() Chain {
	return w.chain
}

func (w *WalletEntity) GetAddress() string {
	return w.address
}

func (w *WalletEntity) GetEncryptedPrivateKey() string {
	return w.encryptedPrivateKey
}

func (w *WalletEntity) GetDerivationPath() string {
	return w.derivationPath
}

func (w *WalletEntity) GetLabel() string {
	return w.label
}

func (w *WalletEntity) GetBalance() decimal.Decimal {
	return w.balance
}

func (w *WalletEntity) GetBalanceUpdatedAt() *time.Time {
	return w.balanceUpdatedAt
}

func (w *WalletEntity) GetStatus() WalletStatus {
	return w.status
}

func (w *WalletEntity) GetCreatedAt() time.Time {
	return w.createdAt
}

func (w *WalletEntity) GetUpdatedAt() time.Time {
	return w.updatedAt
}

// Domain behavior helpers.

// UpdateBalance sets the current balance and optional updated timestamp.
func (w *WalletEntity) UpdateBalance(amount decimal.Decimal, at time.Time) error {
	if amount.IsNegative() {
		return errWalletBalanceNegative
	}
	w.balance = amount
	if at.IsZero() {
		now := time.Now().UTC()
		w.balanceUpdatedAt = &now
	} else {
		t := at.UTC()
		w.balanceUpdatedAt = &t
	}
	return nil
}

// SetStatus transitions the wallet to a new status when valid.
func (w *WalletEntity) SetStatus(status WalletStatus) error {
	if !isValidWalletStatus(status) {
		return errWalletStatusInvalid
	}
	w.status = status
	return nil
}

// Rename updates the human friendly label for the wallet.
func (w *WalletEntity) Rename(label string) {
	w.label = strings.TrimSpace(label)
}

// Touch refreshes the updatedAt timestamp.
func (w *WalletEntity) Touch(at time.Time) {
	if at.IsZero() {
		w.updatedAt = time.Now().UTC()
		return
	}
	w.updatedAt = at
}

func isValidWalletStatus(status WalletStatus) bool {
	switch status {
	case WalletStatusActive, WalletStatusArchived:
		return true
	default:
		return false
	}
}
