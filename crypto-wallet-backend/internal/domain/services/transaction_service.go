package services

import (
    "errors"
    "fmt"
    "log/slog"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/shopspring/decimal"

    "github.com/crypto-wallet/backend/internal/domain/entities"
)

// TransactionService encapsulates domain logic for transaction lifecycles.
type TransactionService struct {
    logger *slog.Logger
}

// NewTransactionService constructs a TransactionService.
func NewTransactionService(logger *slog.Logger) *TransactionService {
    if logger == nil {
        logger = slog.Default()
    }
    return &TransactionService{logger: logger}
}

// SendParams describes a pending outbound transaction.
type SendParams struct {
    WalletID          uuid.UUID
    Chain             entities.Chain
    FromAddress       string
    ToAddress         string
    Amount            decimal.Decimal
    Fee               decimal.Decimal
    Metadata          map[string]any
    DebitAccountID    uuid.UUID
    CreditAccountID   uuid.UUID
    DebitBalanceAfter decimal.Decimal
    CreditBalanceAfter decimal.Decimal
}

// SendResult aggregates the domain artefacts for a send transaction.
type SendResult struct {
    Transaction  *entities.TransactionEntity
    LedgerDebit  *entities.LedgerEntryEntity
    LedgerCredit *entities.LedgerEntryEntity
}

var (
    errInvalidAmount      = errors.New("amount must be positive")
    errInvalidFee         = errors.New("fee cannot be negative")
    errInvalidAddress     = errors.New("address required")
)

// PrepareSend builds a pending transaction and double-entry ledger items.
func (s *TransactionService) PrepareSend(params SendParams) (*SendResult, error) {
    if params.WalletID == uuid.Nil {
        return nil, fmt.Errorf("wallet id is required")
    }
    if !entities.IsSupportedChain(params.Chain) {
        return nil, fmt.Errorf("invalid chain: %s", params.Chain)
    }
    if strings.TrimSpace(params.FromAddress) == "" || strings.TrimSpace(params.ToAddress) == "" {
        return nil, errInvalidAddress
    }
    if params.Amount.LessThanOrEqual(decimal.Zero) {
        return nil, errInvalidAmount
    }
    if params.Fee.IsNegative() {
        return nil, errInvalidFee
    }

    txParams := entities.TransactionParams{
        WalletID:      params.WalletID,
        Chain:         params.Chain,
        Hash:          fmt.Sprintf("PENDING-%s", uuid.NewString()),
        Type:          entities.TransactionTypeSend,
        Amount:        params.Amount,
        Fee:           params.Fee,
        Status:        entities.TransactionStatusPending,
        FromAddress:   params.FromAddress,
        ToAddress:     params.ToAddress,
        Confirmations: 0,
        Metadata:      params.Metadata,
        CreatedAt:     time.Now().UTC(),
        UpdatedAt:     time.Now().UTC(),
    }

    transaction, err := entities.NewTransactionEntity(txParams)
    if err != nil {
        return nil, err
    }

    var (
        debitEntry  *entities.LedgerEntryEntity
        creditEntry *entities.LedgerEntryEntity
    )

    if params.DebitAccountID != uuid.Nil && params.CreditAccountID != uuid.Nil {
        entry, err := entities.NewLedgerEntryEntity(entities.LedgerEntryParams{
            AccountID:     params.DebitAccountID,
            TransactionID: pointerUUID(transaction.GetID()),
            EntryType:     entities.EntryTypeCredit,
            Amount:        params.Amount.Add(params.Fee),
            Currency:      string(params.Chain),
            Description:   fmt.Sprintf("Send %s to %s", params.Amount.String(), params.ToAddress),
            BalanceAfter:  params.DebitBalanceAfter,
        })
        if err != nil {
            return nil, err
        }
        debitEntry = entry

        entry, err = entities.NewLedgerEntryEntity(entities.LedgerEntryParams{
            AccountID:     params.CreditAccountID,
            TransactionID: pointerUUID(transaction.GetID()),
            EntryType:     entities.EntryTypeDebit,
            Amount:        params.Amount,
            Currency:      string(params.Chain),
            Description:   fmt.Sprintf("Receive from %s", params.FromAddress),
            BalanceAfter:  params.CreditBalanceAfter,
        })
        if err != nil {
            return nil, err
        }
        creditEntry = entry
    } else if params.DebitAccountID != uuid.Nil || params.CreditAccountID != uuid.Nil {
        s.logger.Warn("incomplete ledger parameters, skipping ledger entry creation")
    }

    return &SendResult{
        Transaction:  transaction,
        LedgerDebit:  debitEntry,
        LedgerCredit: creditEntry,
    }, nil
}

func pointerUUID(id uuid.UUID) *uuid.UUID {
    if id == uuid.Nil {
        return nil
    }
    value := id
    return &value
}
