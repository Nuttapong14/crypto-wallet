package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ExchangeStatus enumerates the lifecycle states of an exchange operation.
type ExchangeStatus string

const (
	ExchangeStatusPending    ExchangeStatus = "pending"
	ExchangeStatusProcessing ExchangeStatus = "processing"
	ExchangeStatusCompleted  ExchangeStatus = "completed"
	ExchangeStatusFailed     ExchangeStatus = "failed"
	ExchangeStatusCancelled  ExchangeStatus = "cancelled"
)

var (
	errExchangeUserIDRequired       = errors.New("exchange operation user ID is required")
	errExchangeFromWalletIDRequired = errors.New("exchange operation from wallet ID is required")
	errExchangeToWalletIDRequired   = errors.New("exchange operation to wallet ID is required")
	errExchangeFromAmountInvalid    = errors.New("exchange operation from amount must be positive")
	errExchangeToAmountInvalid      = errors.New("exchange operation to amount must be positive")
	errExchangeRateInvalid          = errors.New("exchange operation rate must be positive")
	errExchangeFeeInvalid           = errors.New("exchange operation fee cannot be negative")
	errExchangeStatusInvalid        = errors.New("exchange operation status is invalid")
	errExchangeQuoteExpired         = errors.New("exchange operation quote has expired")
	errExchangeSameWallets          = errors.New("exchange operation from and to wallets cannot be the same")
	errExchangeInsufficientBalance  = errors.New("exchange operation insufficient balance")
)

// ExchangeOperation exposes the behavior required by the application layer when working with exchange entities.
type ExchangeOperation interface {
	Entity
	Identifiable
	Timestamped

	GetUserID() uuid.UUID
	GetFromWalletID() uuid.UUID
	GetToWalletID() uuid.UUID
	GetFromAmount() decimal.Decimal
	GetToAmount() decimal.Decimal
	GetExchangeRate() decimal.Decimal
	GetFeePercentage() decimal.Decimal
	GetFeeAmount() decimal.Decimal
	GetStatus() ExchangeStatus
	GetFromTransactionID() *uuid.UUID
	GetToTransactionID() *uuid.UUID
	GetQuoteExpiresAt() time.Time
	GetExecutedAt() *time.Time
	GetErrorMessage() string
}

// ExchangeOperationEntity is the default implementation of the ExchangeOperation interface.
type ExchangeOperationEntity struct {
	id                uuid.UUID
	userID            uuid.UUID
	fromWalletID      uuid.UUID
	toWalletID        uuid.UUID
	fromAmount        decimal.Decimal
	toAmount          decimal.Decimal
	exchangeRate      decimal.Decimal
	feePercentage     decimal.Decimal
	feeAmount         decimal.Decimal
	status            ExchangeStatus
	fromTransactionID *uuid.UUID
	toTransactionID   *uuid.UUID
	quoteExpiresAt    time.Time
	executedAt        *time.Time
	errorMessage      string
	createdAt         time.Time
	updatedAt         time.Time
}

// ExchangeOperationParams captures the fields required to construct an ExchangeOperationEntity.
type ExchangeOperationParams struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	FromWalletID      uuid.UUID
	ToWalletID        uuid.UUID
	FromAmount        decimal.Decimal
	ToAmount          decimal.Decimal
	ExchangeRate      decimal.Decimal
	FeePercentage     decimal.Decimal
	FeeAmount         decimal.Decimal
	Status            ExchangeStatus
	FromTransactionID *uuid.UUID
	ToTransactionID   *uuid.UUID
	QuoteExpiresAt    time.Time
	ExecutedAt        *time.Time
	ErrorMessage      string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewExchangeOperationEntity validates the supplied parameters and returns a new ExchangeOperationEntity instance.
func NewExchangeOperationEntity(params ExchangeOperationParams) (*ExchangeOperationEntity, error) {
	if params.ID == uuid.Nil {
		params.ID = uuid.New()
	}

	if params.CreatedAt.IsZero() {
		params.CreatedAt = time.Now().UTC()
	}

	if params.UpdatedAt.IsZero() {
		params.UpdatedAt = params.CreatedAt
	}

	if params.QuoteExpiresAt.IsZero() {
		// Default quote expiration: 60 seconds from creation
		params.QuoteExpiresAt = params.CreatedAt.Add(60 * time.Second)
	}

	entity := &ExchangeOperationEntity{
		id:                params.ID,
		userID:            params.UserID,
		fromWalletID:      params.FromWalletID,
		toWalletID:        params.ToWalletID,
		fromAmount:        params.FromAmount,
		toAmount:          params.ToAmount,
		exchangeRate:      params.ExchangeRate,
		feePercentage:     params.FeePercentage,
		feeAmount:         params.FeeAmount,
		status:            params.Status,
		fromTransactionID: params.FromTransactionID,
		toTransactionID:   params.ToTransactionID,
		quoteExpiresAt:    params.QuoteExpiresAt,
		executedAt:        params.ExecutedAt,
		errorMessage:      strings.TrimSpace(params.ErrorMessage),
		createdAt:         params.CreatedAt,
		updatedAt:         params.UpdatedAt,
	}

	if entity.status == "" {
		entity.status = ExchangeStatusPending
	}

	if err := entity.Validate(); err != nil {
		return nil, err
	}

	return entity, nil
}

// HydrateExchangeOperationEntity creates an ExchangeOperationEntity without re-validating invariants (used for repository hydration).
func HydrateExchangeOperationEntity(params ExchangeOperationParams) *ExchangeOperationEntity {
	return &ExchangeOperationEntity{
		id:                params.ID,
		userID:            params.UserID,
		fromWalletID:      params.FromWalletID,
		toWalletID:        params.ToWalletID,
		fromAmount:        params.FromAmount,
		toAmount:          params.ToAmount,
		exchangeRate:      params.ExchangeRate,
		feePercentage:     params.FeePercentage,
		feeAmount:         params.FeeAmount,
		status:            params.Status,
		fromTransactionID: params.FromTransactionID,
		toTransactionID:   params.ToTransactionID,
		quoteExpiresAt:    params.QuoteExpiresAt,
		executedAt:        params.ExecutedAt,
		errorMessage:      strings.TrimSpace(params.ErrorMessage),
		createdAt:         params.CreatedAt,
		updatedAt:         params.UpdatedAt,
	}
}

// Validate ensures the entity adheres to domain invariants.
func (e *ExchangeOperationEntity) Validate() error {
	var validationErr error

	if e.userID == uuid.Nil {
		validationErr = errors.Join(validationErr, errExchangeUserIDRequired)
	}

	if e.fromWalletID == uuid.Nil {
		validationErr = errors.Join(validationErr, errExchangeFromWalletIDRequired)
	}

	if e.toWalletID == uuid.Nil {
		validationErr = errors.Join(validationErr, errExchangeToWalletIDRequired)
	}

	if e.fromWalletID == e.toWalletID {
		validationErr = errors.Join(validationErr, errExchangeSameWallets)
	}

	if e.fromAmount.LessThanOrEqual(decimal.Zero) {
		validationErr = errors.Join(validationErr, errExchangeFromAmountInvalid)
	}

	if e.toAmount.LessThanOrEqual(decimal.Zero) {
		validationErr = errors.Join(validationErr, errExchangeToAmountInvalid)
	}

	if e.exchangeRate.LessThanOrEqual(decimal.Zero) {
		validationErr = errors.Join(validationErr, errExchangeRateInvalid)
	}

	if e.feePercentage.IsNegative() {
		validationErr = errors.Join(validationErr, errExchangeFeeInvalid)
	}

	if e.feeAmount.IsNegative() {
		validationErr = errors.Join(validationErr, errExchangeFeeInvalid)
	}

	if !isValidExchangeStatus(e.status) {
		validationErr = errors.Join(validationErr, errExchangeStatusInvalid)
	}

	return validationErr
}

// Getter implementations satisfy the ExchangeOperation interface.

func (e *ExchangeOperationEntity) GetID() uuid.UUID {
	return e.id
}

func (e *ExchangeOperationEntity) GetUserID() uuid.UUID {
	return e.userID
}

func (e *ExchangeOperationEntity) GetFromWalletID() uuid.UUID {
	return e.fromWalletID
}

func (e *ExchangeOperationEntity) GetToWalletID() uuid.UUID {
	return e.toWalletID
}

func (e *ExchangeOperationEntity) GetFromAmount() decimal.Decimal {
	return e.fromAmount
}

func (e *ExchangeOperationEntity) GetToAmount() decimal.Decimal {
	return e.toAmount
}

func (e *ExchangeOperationEntity) GetExchangeRate() decimal.Decimal {
	return e.exchangeRate
}

func (e *ExchangeOperationEntity) GetFeePercentage() decimal.Decimal {
	return e.feePercentage
}

func (e *ExchangeOperationEntity) GetFeeAmount() decimal.Decimal {
	return e.feeAmount
}

func (e *ExchangeOperationEntity) GetStatus() ExchangeStatus {
	return e.status
}

func (e *ExchangeOperationEntity) GetFromTransactionID() *uuid.UUID {
	return e.fromTransactionID
}

func (e *ExchangeOperationEntity) GetToTransactionID() *uuid.UUID {
	return e.toTransactionID
}

func (e *ExchangeOperationEntity) GetQuoteExpiresAt() time.Time {
	return e.quoteExpiresAt
}

func (e *ExchangeOperationEntity) GetExecutedAt() *time.Time {
	return e.executedAt
}

func (e *ExchangeOperationEntity) GetErrorMessage() string {
	return e.errorMessage
}

func (e *ExchangeOperationEntity) GetCreatedAt() time.Time {
	return e.createdAt
}

func (e *ExchangeOperationEntity) GetUpdatedAt() time.Time {
	return e.updatedAt
}

// Domain behavior helpers.

// SetStatus transitions the exchange operation to the provided status when valid.
func (e *ExchangeOperationEntity) SetStatus(status ExchangeStatus) error {
	if !isValidExchangeStatus(status) {
		return errExchangeStatusInvalid
	}
	e.status = status
	return nil
}

// SetFromTransactionID records the transaction ID for the debit (from) transaction.
func (e *ExchangeOperationEntity) SetFromTransactionID(txID uuid.UUID) {
	e.fromTransactionID = &txID
}

// SetToTransactionID records the transaction ID for the credit (to) transaction.
func (e *ExchangeOperationEntity) SetToTransactionID(txID uuid.UUID) {
	e.toTransactionID = &txID
}

// MarkProcessing sets the status to processing and validates that the quote hasn't expired.
func (e *ExchangeOperationEntity) MarkProcessing() error {
	if time.Now().UTC().After(e.quoteExpiresAt) {
		return errExchangeQuoteExpired
	}
	return e.SetStatus(ExchangeStatusProcessing)
}

// MarkCompleted sets the status to completed and records the execution timestamp.
func (e *ExchangeOperationEntity) MarkCompleted(at time.Time) error {
	if err := e.SetStatus(ExchangeStatusCompleted); err != nil {
		return err
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}
	e.executedAt = &at
	return nil
}

// MarkFailed sets the status to failed and records an error message.
func (e *ExchangeOperationEntity) MarkFailed(message string) error {
	if err := e.SetStatus(ExchangeStatusFailed); err != nil {
		return err
	}
	e.errorMessage = strings.TrimSpace(message)
	return nil
}

// MarkCancelled sets the status to cancelled.
func (e *ExchangeOperationEntity) MarkCancelled() error {
	return e.SetStatus(ExchangeStatusCancelled)
}

// IsQuoteExpired checks if the quote has expired.
func (e *ExchangeOperationEntity) IsQuoteExpired() bool {
	return time.Now().UTC().After(e.quoteExpiresAt)
}

// ExtendQuoteExpiration extends the quote expiration time by the specified duration.
func (e *ExchangeOperationEntity) ExtendQuoteExpiration(duration time.Duration) {
	e.quoteExpiresAt = e.quoteExpiresAt.Add(duration)
}

// SetErrorMessage records an error message associated with the exchange operation.
func (e *ExchangeOperationEntity) SetErrorMessage(message string) {
	e.errorMessage = strings.TrimSpace(message)
}

// Touch refreshes the updatedAt timestamp.
func (e *ExchangeOperationEntity) Touch(at time.Time) {
	if at.IsZero() {
		e.updatedAt = time.Now().UTC()
		return
	}
	e.updatedAt = at
}

func isValidExchangeStatus(status ExchangeStatus) bool {
	switch status {
	case ExchangeStatusPending, ExchangeStatusProcessing, ExchangeStatusCompleted, ExchangeStatusFailed, ExchangeStatusCancelled:
		return true
	default:
		return false
	}
}
