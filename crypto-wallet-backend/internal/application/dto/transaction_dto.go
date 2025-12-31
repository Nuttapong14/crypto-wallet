package dto

import (
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/shopspring/decimal"

    "github.com/crypto-wallet/backend/internal/domain/entities"
    "github.com/crypto-wallet/backend/pkg/utils"
)

// SendTransactionRequest captures the payload required to initiate a blockchain transfer.
type SendTransactionRequest struct {
    WalletID   string            `json:"walletId"`
    Chain      string            `json:"chain"`
    ToAddress  string            `json:"toAddress"`
    Amount     string            `json:"amount"`
    Fee        string            `json:"fee,omitempty"`
    Memo       string            `json:"memo,omitempty"`
    Metadata   map[string]any    `json:"metadata,omitempty"`
}

// Validate enforces request invariants.
func (r SendTransactionRequest) Validate() utils.ValidationErrors {
    errs := utils.ValidationErrors{}
    utils.RequireUUID(&errs, "walletId", r.WalletID)
    utils.Require(&errs, "chain", r.Chain)
    utils.Require(&errs, "toAddress", r.ToAddress)
    utils.Require(&errs, "amount", r.Amount)

    if _, err := decimal.NewFromString(r.Amount); err != nil {
        errs.Add("amount", "must be a valid decimal string")
    }
    if strings.TrimSpace(r.Fee) != "" {
        if fee, err := decimal.NewFromString(r.Fee); err != nil {
            errs.Add("fee", "must be a valid decimal string")
        } else if fee.IsNegative() {
            errs.Add("fee", "cannot be negative")
        }
    }

    return errs
}

// TransactionStatusResponse provides transaction status details.
type TransactionStatusResponse struct {
    ID            uuid.UUID         `json:"id"`
    WalletID      uuid.UUID         `json:"walletId"`
    Chain         string            `json:"chain"`
    Hash          string            `json:"hash"`
    Type          string            `json:"type"`
    Amount        string            `json:"amount"`
    Fee           string            `json:"fee"`
    Status        string            `json:"status"`
    Confirmations int               `json:"confirmations"`
    FromAddress   string            `json:"fromAddress"`
    ToAddress     string            `json:"toAddress"`
    BlockNumber   *uint64           `json:"blockNumber,omitempty"`
    ErrorMessage  string            `json:"errorMessage,omitempty"`
    Metadata      map[string]any    `json:"metadata,omitempty"`
    CreatedAt     string            `json:"createdAt"`
    ConfirmedAt   *string           `json:"confirmedAt,omitempty"`
    UpdatedAt     string            `json:"updatedAt"`
}

// NewTransactionStatusResponse maps a domain entity to API response.
func NewTransactionStatusResponse(tx entities.Transaction) TransactionStatusResponse {
    confirmedAt := tx.GetConfirmedAt()
    var confirmedAtStr *string
    if confirmedAt != nil {
        value := confirmedAt.UTC().Format(time.RFC3339Nano)
        confirmedAtStr = &value
    }

    blockNumber := tx.GetBlockNumber()

    return TransactionStatusResponse{
        ID:            tx.GetID(),
        WalletID:      tx.GetWalletID(),
        Chain:         string(tx.GetChain()),
        Hash:          tx.GetHash(),
        Type:          string(tx.GetType()),
        Amount:        tx.GetAmount().String(),
        Fee:           tx.GetFee().String(),
        Status:        string(tx.GetStatus()),
        Confirmations: tx.GetConfirmations(),
        FromAddress:   tx.GetFromAddress(),
        ToAddress:     tx.GetToAddress(),
        BlockNumber:   blockNumber,
        ErrorMessage:  tx.GetErrorMessage(),
        Metadata:      tx.GetMetadata(),
        CreatedAt:     tx.GetCreatedAt().UTC().Format(time.RFC3339Nano),
        ConfirmedAt:   confirmedAtStr,
        UpdatedAt:     tx.GetUpdatedAt().UTC().Format(time.RFC3339Nano),
    }
}

// TransactionListResponse aggregates paginated transactions.
type TransactionListResponse struct {
    Items      []TransactionStatusResponse `json:"items"`
    Total      int64                       `json:"total"`
    Limit      int                         `json:"limit"`
    Offset     int                         `json:"offset"`
}
