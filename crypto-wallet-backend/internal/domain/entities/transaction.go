package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TransactionType enumerates supported transaction directions and intents.
type TransactionType string

const (
	TransactionTypeSend     TransactionType = "send"
	TransactionTypeReceive  TransactionType = "receive"
	TransactionTypeSwapIn   TransactionType = "swap_in"
	TransactionTypeSwapOut  TransactionType = "swap_out"
)

// TransactionStatus captures the lifecycle of a blockchain transaction.
type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "pending"
	TransactionStatusConfirming TransactionStatus = "confirming"
	TransactionStatusConfirmed  TransactionStatus = "confirmed"
	TransactionStatusFailed     TransactionStatus = "failed"
	TransactionStatusCancelled  TransactionStatus = "cancelled"
)

var (
	errTransactionWalletIDRequired = errors.New("transaction wallet ID is required")
	errTransactionChainInvalid     = errors.New("transaction chain is invalid")
	errTransactionHashRequired     = errors.New("transaction hash is required")
	errTransactionTypeInvalid      = errors.New("transaction type is invalid")
	errTransactionAmountInvalid    = errors.New("transaction amount must be positive")
	errTransactionFeeInvalid       = errors.New("transaction fee cannot be negative")
	errTransactionStatusInvalid    = errors.New("transaction status is invalid")
	errTransactionFromAddress      = errors.New("transaction from address is required")
	errTransactionToAddress        = errors.New("transaction to address is required")
	errTransactionConfirmations    = errors.New("transaction confirmations cannot be negative")
)

// Transaction exposes the behavior required by the application layer when working with transaction entities.
type Transaction interface {
	Entity
	Identifiable
	Timestamped

	GetWalletID() uuid.UUID
	GetChain() Chain
	GetHash() string
	GetType() TransactionType
	GetAmount() decimal.Decimal
	GetFee() decimal.Decimal
	GetStatus() TransactionStatus
	GetFromAddress() string
	GetToAddress() string
	GetBlockNumber() *uint64
	GetConfirmations() int
	GetErrorMessage() string
	GetMetadata() map[string]any
	GetConfirmedAt() *time.Time
}

// TransactionEntity is the default implementation of the Transaction interface.
type TransactionEntity struct {
	id            uuid.UUID
	walletID      uuid.UUID
	chain         Chain
	hash          string
	txType        TransactionType
	amount        decimal.Decimal
	fee           decimal.Decimal
	status        TransactionStatus
	fromAddress   string
	toAddress     string
	blockNumber   *uint64
	confirmations int
	errorMessage  string
	metadata      map[string]any
	createdAt     time.Time
	confirmedAt   *time.Time
	updatedAt     time.Time
}

// TransactionParams captures the fields required to construct a TransactionEntity.
type TransactionParams struct {
	ID            uuid.UUID
	WalletID      uuid.UUID
	Chain         Chain
	Hash          string
	Type          TransactionType
	Amount        decimal.Decimal
	Fee           decimal.Decimal
	Status        TransactionStatus
	FromAddress   string
	ToAddress     string
	BlockNumber   *uint64
	Confirmations int
	ErrorMessage  string
	Metadata      map[string]any
	CreatedAt     time.Time
	ConfirmedAt   *time.Time
	UpdatedAt     time.Time
}

// NewTransactionEntity validates the supplied parameters and returns a new TransactionEntity instance.
func NewTransactionEntity(params TransactionParams) (*TransactionEntity, error) {
	if params.ID == uuid.Nil {
		params.ID = uuid.New()
	}

	if params.CreatedAt.IsZero() {
		params.CreatedAt = time.Now().UTC()
	}

	if params.UpdatedAt.IsZero() {
		params.UpdatedAt = params.CreatedAt
	}

	if params.Metadata == nil {
		params.Metadata = make(map[string]any)
	}

	entity := &TransactionEntity{
		id:            params.ID,
		walletID:      params.WalletID,
		chain:         params.Chain,
		hash:          strings.TrimSpace(params.Hash),
		txType:        params.Type,
		amount:        params.Amount,
		fee:           params.Fee,
		status:        params.Status,
		fromAddress:   strings.TrimSpace(params.FromAddress),
		toAddress:     strings.TrimSpace(params.ToAddress),
		blockNumber:   params.BlockNumber,
		confirmations: params.Confirmations,
		errorMessage:  strings.TrimSpace(params.ErrorMessage),
		metadata:      params.Metadata,
		createdAt:     params.CreatedAt,
		confirmedAt:   params.ConfirmedAt,
		updatedAt:     params.UpdatedAt,
	}

	if entity.status == "" {
		entity.status = TransactionStatusPending
	}

	if err := entity.Validate(); err != nil {
		return nil, err
	}

	return entity, nil
}

// HydrateTransactionEntity creates a TransactionEntity without re-validating invariants (used for repository hydration).
func HydrateTransactionEntity(params TransactionParams) *TransactionEntity {
	if params.Metadata == nil {
		params.Metadata = make(map[string]any)
	}

	return &TransactionEntity{
		id:            params.ID,
		walletID:      params.WalletID,
		chain:         params.Chain,
		hash:          strings.TrimSpace(params.Hash),
		txType:        params.Type,
		amount:        params.Amount,
		fee:           params.Fee,
		status:        params.Status,
		fromAddress:   strings.TrimSpace(params.FromAddress),
		toAddress:     strings.TrimSpace(params.ToAddress),
		blockNumber:   params.BlockNumber,
		confirmations: params.Confirmations,
		errorMessage:  strings.TrimSpace(params.ErrorMessage),
		metadata:      params.Metadata,
		createdAt:     params.CreatedAt,
		confirmedAt:   params.ConfirmedAt,
		updatedAt:     params.UpdatedAt,
	}
}

// Validate ensures the entity adheres to domain invariants.
func (t *TransactionEntity) Validate() error {
	var validationErr error

	if t.walletID == uuid.Nil {
		validationErr = errors.Join(validationErr, errTransactionWalletIDRequired)
	}

	if !isValidChain(t.chain) {
		validationErr = errors.Join(validationErr, errTransactionChainInvalid)
	}

	if t.hash == "" {
		validationErr = errors.Join(validationErr, errTransactionHashRequired)
	}

	if !isValidTransactionType(t.txType) {
		validationErr = errors.Join(validationErr, errTransactionTypeInvalid)
	}

	if t.amount.LessThanOrEqual(decimal.Zero) {
		validationErr = errors.Join(validationErr, errTransactionAmountInvalid)
	}

	if t.fee.IsNegative() {
		validationErr = errors.Join(validationErr, errTransactionFeeInvalid)
	}

	if !isValidTransactionStatus(t.status) {
		validationErr = errors.Join(validationErr, errTransactionStatusInvalid)
	}

	if t.fromAddress == "" {
		validationErr = errors.Join(validationErr, errTransactionFromAddress)
	}

	if t.toAddress == "" {
		validationErr = errors.Join(validationErr, errTransactionToAddress)
	}

	if t.confirmations < 0 {
		validationErr = errors.Join(validationErr, errTransactionConfirmations)
	}

	return validationErr
}

// Getter implementations satisfy the Transaction interface.

func (t *TransactionEntity) GetID() uuid.UUID {
	return t.id
}

func (t *TransactionEntity) GetWalletID() uuid.UUID {
	return t.walletID
}

func (t *TransactionEntity) GetChain() Chain {
	return t.chain
}

func (t *TransactionEntity) GetHash() string {
	return t.hash
}

func (t *TransactionEntity) GetType() TransactionType {
	return t.txType
}

func (t *TransactionEntity) GetAmount() decimal.Decimal {
	return t.amount
}

func (t *TransactionEntity) GetFee() decimal.Decimal {
	return t.fee
}

func (t *TransactionEntity) GetStatus() TransactionStatus {
	return t.status
}

func (t *TransactionEntity) GetFromAddress() string {
	return t.fromAddress
}

func (t *TransactionEntity) GetToAddress() string {
	return t.toAddress
}

func (t *TransactionEntity) GetBlockNumber() *uint64 {
	return t.blockNumber
}

func (t *TransactionEntity) GetConfirmations() int {
	return t.confirmations
}

func (t *TransactionEntity) GetErrorMessage() string {
	return t.errorMessage
}

func (t *TransactionEntity) GetMetadata() map[string]any {
	return t.metadata
}

func (t *TransactionEntity) GetCreatedAt() time.Time {
	return t.createdAt
}

func (t *TransactionEntity) GetConfirmedAt() *time.Time {
	return t.confirmedAt
}

func (t *TransactionEntity) GetUpdatedAt() time.Time {
	return t.updatedAt
}

// Domain behavior helpers.

// SetStatus transitions the transaction to the provided status when valid.
func (t *TransactionEntity) SetStatus(status TransactionStatus) error {
	if !isValidTransactionStatus(status) {
		return errTransactionStatusInvalid
	}
	t.status = status
	return nil
}

// SetConfirmations updates the confirmation count; zero or positive values are allowed.
func (t *TransactionEntity) SetConfirmations(confirmations int) error {
	if confirmations < 0 {
		return errTransactionConfirmations
	}
	t.confirmations = confirmations
	return nil
}

// SetHash updates the transaction hash once obtained from the blockchain.
func (t *TransactionEntity) SetHash(hash string) error {
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return errTransactionHashRequired
	}
	t.hash = hash
	return nil
}

// MergeMetadata merges blockchain metadata into the transaction record.
func (t *TransactionEntity) MergeMetadata(values map[string]any) {
	if values == nil {
		return
	}
	if t.metadata == nil {
		t.metadata = make(map[string]any)
	}
	for key, value := range values {
		t.metadata[key] = value
	}
}

// SetBlockNumber records the block number that included the transaction.
func (t *TransactionEntity) SetBlockNumber(number uint64) {
	t.blockNumber = &number
}

// SetErrorMessage records an error message associated with the transaction.
func (t *TransactionEntity) SetErrorMessage(message string) {
	t.errorMessage = strings.TrimSpace(message)
}

// MarkConfirmed sets the status to confirmed, updates confirmations, and records the timestamp.
func (t *TransactionEntity) MarkConfirmed(confirmations int, at time.Time) error {
	if err := t.SetStatus(TransactionStatusConfirmed); err != nil {
		return err
	}
	if err := t.SetConfirmations(confirmations); err != nil {
		return err
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}
	t.confirmedAt = &at
	return nil
}

// Touch refreshes the updatedAt timestamp.
func (t *TransactionEntity) Touch(at time.Time) {
	if at.IsZero() {
		t.updatedAt = time.Now().UTC()
		return
	}
	t.updatedAt = at
}

func isValidChain(chain Chain) bool {
	return IsSupportedChain(chain)
}

func isValidTransactionType(txType TransactionType) bool {
	switch txType {
	case TransactionTypeSend, TransactionTypeReceive, TransactionTypeSwapIn, TransactionTypeSwapOut:
		return true
	default:
		return false
	}
}

func isValidTransactionStatus(status TransactionStatus) bool {
	switch status {
	case TransactionStatusPending, TransactionStatusConfirming, TransactionStatusConfirmed, TransactionStatusFailed, TransactionStatusCancelled:
		return true
	default:
		return false
	}
}
