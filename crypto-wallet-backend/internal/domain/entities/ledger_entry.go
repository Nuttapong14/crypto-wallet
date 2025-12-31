package entities

import (
    "errors"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/shopspring/decimal"
)

// EntryType specifies the double-entry accounting direction.
type EntryType string

const (
    // EntryTypeDebit increases an account balance.
    EntryTypeDebit EntryType = "debit"
    // EntryTypeCredit decreases an account balance.
    EntryTypeCredit EntryType = "credit"
)

var (
    errLedgerAccountRequired  = errors.New("ledger account ID is required")
    errLedgerEntryTypeInvalid = errors.New("ledger entry type is invalid")
    errLedgerAmountInvalid    = errors.New("ledger amount must be positive")
    errLedgerCurrencyRequired = errors.New("ledger currency is required")
    errLedgerDescriptionEmpty = errors.New("ledger description is required")
)

// LedgerEntry exposes ledger entry behaviour.
type LedgerEntry interface {
    Entity
    Identifiable

    GetAccountID() uuid.UUID
    GetTransactionID() *uuid.UUID
    GetEntryType() EntryType
    GetAmount() decimal.Decimal
    GetCurrency() string
    GetDescription() string
    GetBalanceAfter() decimal.Decimal
    GetCreatedAt() time.Time
}

// LedgerEntryEntity implements LedgerEntry.
type LedgerEntryEntity struct {
    id            uuid.UUID
    accountID     uuid.UUID
    transactionID *uuid.UUID
    entryType     EntryType
    amount        decimal.Decimal
    currency      string
    description   string
    balanceAfter  decimal.Decimal
    createdAt     time.Time
}

// LedgerEntryParams captures constructor input.
type LedgerEntryParams struct {
    ID            uuid.UUID
    AccountID     uuid.UUID
    TransactionID *uuid.UUID
    EntryType     EntryType
    Amount        decimal.Decimal
    Currency      string
    Description   string
    BalanceAfter  decimal.Decimal
    CreatedAt     time.Time
}

// NewLedgerEntryEntity validates and returns a ledger entry.
func NewLedgerEntryEntity(params LedgerEntryParams) (*LedgerEntryEntity, error) {
    if params.ID == uuid.Nil {
        params.ID = uuid.New()
    }
    if params.CreatedAt.IsZero() {
        params.CreatedAt = time.Now().UTC()
    }

    entity := &LedgerEntryEntity{
        id:            params.ID,
        accountID:     params.AccountID,
        transactionID: params.TransactionID,
        entryType:     params.EntryType,
        amount:        params.Amount,
        currency:      strings.ToUpper(strings.TrimSpace(params.Currency)),
        description:   strings.TrimSpace(params.Description),
        balanceAfter:  params.BalanceAfter,
        createdAt:     params.CreatedAt,
    }

    if err := entity.Validate(); err != nil {
        return nil, err
    }

    return entity, nil
}

// HydrateLedgerEntryEntity constructs without validation.
func HydrateLedgerEntryEntity(params LedgerEntryParams) *LedgerEntryEntity {
    return &LedgerEntryEntity{
        id:            params.ID,
        accountID:     params.AccountID,
        transactionID: params.TransactionID,
        entryType:     params.EntryType,
        amount:        params.Amount,
        currency:      strings.ToUpper(strings.TrimSpace(params.Currency)),
        description:   strings.TrimSpace(params.Description),
        balanceAfter:  params.BalanceAfter,
        createdAt:     params.CreatedAt,
    }
}

// Validate checks invariants.
func (l *LedgerEntryEntity) Validate() error {
    var validationErr error

    if l.accountID == uuid.Nil {
        validationErr = errors.Join(validationErr, errLedgerAccountRequired)
    }

    if !isValidEntryType(l.entryType) {
        validationErr = errors.Join(validationErr, errLedgerEntryTypeInvalid)
    }

    if l.amount.LessThanOrEqual(decimal.Zero) {
        validationErr = errors.Join(validationErr, errLedgerAmountInvalid)
    }

    if l.currency == "" {
        validationErr = errors.Join(validationErr, errLedgerCurrencyRequired)
    }

    if l.description == "" {
        validationErr = errors.Join(validationErr, errLedgerDescriptionEmpty)
    }

    return validationErr
}

// Getters

func (l *LedgerEntryEntity) GetID() uuid.UUID {
    return l.id
}

func (l *LedgerEntryEntity) GetAccountID() uuid.UUID {
    return l.accountID
}

func (l *LedgerEntryEntity) GetTransactionID() *uuid.UUID {
    return l.transactionID
}

func (l *LedgerEntryEntity) GetEntryType() EntryType {
    return l.entryType
}

func (l *LedgerEntryEntity) GetAmount() decimal.Decimal {
    return l.amount
}

func (l *LedgerEntryEntity) GetCurrency() string {
    return l.currency
}

func (l *LedgerEntryEntity) GetDescription() string {
    return l.description
}

func (l *LedgerEntryEntity) GetBalanceAfter() decimal.Decimal {
    return l.balanceAfter
}

func (l *LedgerEntryEntity) GetCreatedAt() time.Time {
    return l.createdAt
}

func isValidEntryType(entryType EntryType) bool {
    switch entryType {
    case EntryTypeDebit, EntryTypeCredit:
        return true
    default:
        return false
    }
}
