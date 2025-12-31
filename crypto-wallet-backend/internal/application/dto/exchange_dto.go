package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ExchangeRateResponse represents the response for getting exchange rates.
type ExchangeRateResponse struct {
	BaseSymbol    string           `json:"base_symbol"`
	QuoteSymbol   string           `json:"quote_symbol"`
	ExchangeRate  decimal.Decimal  `json:"exchange_rate"`
	InverseRate   decimal.Decimal  `json:"inverse_rate"`
	FeePercentage decimal.Decimal  `json:"fee_percentage"`
	MinSwapAmount decimal.Decimal  `json:"min_swap_amount"`
	MaxSwapAmount *decimal.Decimal `json:"max_swap_amount,omitempty"`
	IsActive      bool             `json:"is_active"`
	HasLiquidity  bool             `json:"has_liquidity"`
	LastUpdated   time.Time        `json:"last_updated"`
}

// QuoteRequest represents the request for getting an exchange quote.
type QuoteRequest struct {
	FromWalletID uuid.UUID `json:"from_wallet_id" validate:"required"`
	ToWalletID   uuid.UUID `json:"to_wallet_id" validate:"required"`
	FromAmount   string    `json:"from_amount" validate:"required,numeric"`
}

// QuoteResponse represents the response for an exchange quote.
type QuoteResponse struct {
	OperationID    uuid.UUID       `json:"operation_id"`
	FromWalletID   uuid.UUID       `json:"from_wallet_id"`
	ToWalletID     uuid.UUID       `json:"to_wallet_id"`
	FromAmount     decimal.Decimal `json:"from_amount"`
	ToAmount       decimal.Decimal `json:"to_amount"`
	ExchangeRate   decimal.Decimal `json:"exchange_rate"`
	FeePercentage  decimal.Decimal `json:"fee_percentage"`
	FeeAmount      decimal.Decimal `json:"fee_amount"`
	QuoteExpiresAt time.Time       `json:"quote_expires_at"`
	ExpiresIn      int             `json:"expires_in_seconds"` // Seconds until expiration
}

// ExecuteExchangeRequest represents the request to execute an exchange.
type ExecuteExchangeRequest struct {
	OperationID uuid.UUID `json:"operation_id" validate:"required"`
}

// ExecuteExchangeResponse represents the response after executing an exchange.
type ExecuteExchangeResponse struct {
	OperationID       uuid.UUID       `json:"operation_id"`
	Status            string          `json:"status"`
	FromWalletID      uuid.UUID       `json:"from_wallet_id"`
	ToWalletID        uuid.UUID       `json:"to_wallet_id"`
	FromAmount        decimal.Decimal `json:"from_amount"`
	ToAmount          decimal.Decimal `json:"to_amount"`
	ExchangeRate      decimal.Decimal `json:"exchange_rate"`
	FeeAmount         decimal.Decimal `json:"fee_amount"`
	ExecutedAt        *time.Time      `json:"executed_at,omitempty"`
	FromTransactionID *uuid.UUID      `json:"from_transaction_id,omitempty"`
	ToTransactionID   *uuid.UUID      `json:"to_transaction_id,omitempty"`
	ErrorMessage      string          `json:"error_message,omitempty"`
}

// CancelExchangeRequest represents the request to cancel an exchange.
type CancelExchangeRequest struct {
	OperationID uuid.UUID `json:"operation_id" validate:"required"`
	Reason      string    `json:"reason,omitempty"`
}

// CancelExchangeResponse represents the response after canceling an exchange.
type CancelExchangeResponse struct {
	OperationID uuid.UUID `json:"operation_id"`
	Status      string    `json:"status"`
	CancelledAt time.Time `json:"cancelled_at"`
	Reason      string    `json:"reason,omitempty"`
}

// ExchangeOperationResponse represents a single exchange operation in the history.
type ExchangeOperationResponse struct {
	ID                uuid.UUID       `json:"id"`
	UserID            uuid.UUID       `json:"user_id"`
	FromWalletID      uuid.UUID       `json:"from_wallet_id"`
	ToWalletID        uuid.UUID       `json:"to_wallet_id"`
	FromAmount        decimal.Decimal `json:"from_amount"`
	ToAmount          decimal.Decimal `json:"to_amount"`
	ExchangeRate      decimal.Decimal `json:"exchange_rate"`
	FeePercentage     decimal.Decimal `json:"fee_percentage"`
	FeeAmount         decimal.Decimal `json:"fee_amount"`
	Status            string          `json:"status"`
	QuoteExpiresAt    time.Time       `json:"quote_expires_at"`
	ExecutedAt        *time.Time      `json:"executed_at,omitempty"`
	FromTransactionID *uuid.UUID      `json:"from_transaction_id,omitempty"`
	ToTransactionID   *uuid.UUID      `json:"to_transaction_id,omitempty"`
	ErrorMessage      string          `json:"error_message,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// ExchangeHistoryRequest represents the request for getting exchange history.
type ExchangeHistoryRequest struct {
	Status       *string    `json:"status,omitempty" validate:"omitempty,oneof=pending processing completed failed cancelled"`
	FromWalletID *uuid.UUID `json:"from_wallet_id,omitempty"`
	ToWalletID   *uuid.UUID `json:"to_wallet_id,omitempty"`
	DateFrom     *time.Time `json:"date_from,omitempty"`
	DateTo       *time.Time `json:"date_to,omitempty"`
	MinAmount    *string    `json:"min_amount,omitempty" validate:"omitempty,numeric"`
	MaxAmount    *string    `json:"max_amount,omitempty" validate:"omitempty,numeric"`
	Page         int        `json:"page" validate:"min=1"`
	PageSize     int        `json:"page_size" validate:"min=1,max=100"`
}

// ExchangeHistoryResponse represents the paginated response for exchange history.
type ExchangeHistoryResponse struct {
	Operations []ExchangeOperationResponse `json:"operations"`
	Total      int64                       `json:"total"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"page_size"`
	TotalPages int                         `json:"total_pages"`
}

// TradingPairsResponse represents the response for getting active trading pairs.
type TradingPairsResponse struct {
	Pairs []ExchangeRateResponse `json:"pairs"`
}

// ExchangeStatsResponse represents exchange statistics for a user.
type ExchangeStatsResponse struct {
	TotalOperations int64           `json:"total_operations"`
	TotalVolume     decimal.Decimal `json:"total_volume"`
	CompletedCount  int64           `json:"completed_count"`
	FailedCount     int64           `json:"failed_count"`
	PendingCount    int64           `json:"pending_count"`
}

// ValidationError represents a validation error response.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorResponse represents a standard error response.
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details []ValidationError `json:"details,omitempty"`
}

// SuccessResponse represents a standard success response.
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
