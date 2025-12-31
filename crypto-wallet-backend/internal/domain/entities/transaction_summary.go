package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	errTransactionSummaryChainRequired      = errors.New("transaction summary chain is required")
	errTransactionSummaryChainInvalid       = errors.New("transaction summary chain is invalid")
	errTransactionSummaryDateRequired       = errors.New("transaction summary date is required")
	errTransactionSummaryCountInvalid       = errors.New("transaction summary count cannot be negative")
	errTransactionSummaryVolumeInvalid      = errors.New("transaction summary volume cannot be negative")
	errTransactionSummaryUniqueCountInvalid = errors.New("transaction summary unique count cannot be negative")
	errTransactionSummaryAverageInvalid     = errors.New("transaction summary average value cannot be negative")
	errTransactionSummarySwapCountInvalid   = errors.New("transaction summary swap count cannot be negative")
	errTransactionSummaryFeesInvalid        = errors.New("transaction summary fees cannot be negative")
)

// TransactionSummary represents aggregated transaction metrics for analytics
type TransactionSummary struct {
	ID                     uuid.UUID `json:"id" db:"id"`
	Chain                  Chain     `json:"chain" db:"chain"`
	Date                   time.Time `json:"date" db:"date"`
	TransactionCount       int       `json:"transaction_count" db:"transaction_count"`
	TotalVolumeUSD         float64   `json:"total_volume_usd" db:"total_volume_usd"`
	UniqueSenders          int       `json:"unique_senders" db:"unique_senders"`
	UniqueReceivers        int       `json:"unique_receivers" db:"unique_receivers"`
	AvgTransactionValueUSD float64   `json:"avg_transaction_value_usd" db:"avg_transaction_value_usd"`
	SwapCount              int       `json:"swap_count" db:"swap_count"`
	TotalFeesUSD           float64   `json:"total_fees_usd" db:"total_fees_usd"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

// TableName returns the database table name for TransactionSummary
func (ts *TransactionSummary) TableName() string {
	return "transaction_summary"
}

// Validate performs basic validation on the TransactionSummary entity
func (ts *TransactionSummary) Validate() error {
	var validationErr error

	if ts.Chain == "" {
		validationErr = errors.Join(validationErr, errTransactionSummaryChainRequired)
	} else if !IsSupportedChain(ts.Chain) {
		validationErr = errors.Join(validationErr, errTransactionSummaryChainInvalid)
	}

	if ts.Date.IsZero() {
		validationErr = errors.Join(validationErr, errTransactionSummaryDateRequired)
	}

	if ts.TransactionCount < 0 {
		validationErr = errors.Join(validationErr, errTransactionSummaryCountInvalid)
	}

	if ts.TotalVolumeUSD < 0 {
		validationErr = errors.Join(validationErr, errTransactionSummaryVolumeInvalid)
	}

	if ts.UniqueSenders < 0 {
		validationErr = errors.Join(validationErr, errTransactionSummaryUniqueCountInvalid)
	}

	if ts.UniqueReceivers < 0 {
		validationErr = errors.Join(validationErr, errTransactionSummaryUniqueCountInvalid)
	}

	if ts.AvgTransactionValueUSD < 0 {
		validationErr = errors.Join(validationErr, errTransactionSummaryAverageInvalid)
	}

	if ts.SwapCount < 0 {
		validationErr = errors.Join(validationErr, errTransactionSummarySwapCountInvalid)
	}

	if ts.TotalFeesUSD < 0 {
		validationErr = errors.Join(validationErr, errTransactionSummaryFeesInvalid)
	}

	return validationErr
}

// GetDailyVolume returns the daily volume in USD for this summary
func (ts *TransactionSummary) GetDailyVolume() float64 {
	return ts.TotalVolumeUSD
}

// GetAverageTransactionSize returns the average transaction size
func (ts *TransactionSummary) GetAverageTransactionSize() float64 {
	if ts.TransactionCount == 0 {
		return 0
	}
	return ts.TotalVolumeUSD / float64(ts.TransactionCount)
}

// GetSwapPercentage returns the percentage of transactions that are swaps
func (ts *TransactionSummary) GetSwapPercentage() float64 {
	if ts.TransactionCount == 0 {
		return 0
	}
	return float64(ts.SwapCount) / float64(ts.TransactionCount) * 100
}

// GetFeePercentage returns the fee percentage relative to total volume
func (ts *TransactionSummary) GetFeePercentage() float64 {
	if ts.TotalVolumeUSD == 0 {
		return 0
	}
	return ts.TotalFeesUSD / ts.TotalVolumeUSD * 100
}

// GetID returns the ID of the transaction summary
func (ts *TransactionSummary) GetID() uuid.UUID {
	return ts.ID
}

// GetChain returns the blockchain chain
func (ts *TransactionSummary) GetChain() Chain {
	return ts.Chain
}

// GetDate returns the date of the summary
func (ts *TransactionSummary) GetDate() time.Time {
	return ts.Date
}

// GetTotalTransactions returns the total number of transactions
func (ts *TransactionSummary) GetTotalTransactions() int64 {
	return int64(ts.TransactionCount)
}

// GetTotalVolume returns the total volume in USD
func (ts *TransactionSummary) GetTotalVolume() float64 {
	return ts.TotalVolumeUSD
}

// GetSwapTransactions returns the number of swap transactions
func (ts *TransactionSummary) GetSwapTransactions() int64 {
	return int64(ts.SwapCount)
}

// GetTotalFees returns the total fees in USD
func (ts *TransactionSummary) GetTotalFees() float64 {
	return ts.TotalFeesUSD
}

// GetAverageFee returns the average fee per transaction
func (ts *TransactionSummary) GetAverageFee() float64 {
	if ts.TransactionCount == 0 {
		return 0
	}
	return ts.TotalFeesUSD / float64(ts.TransactionCount)
}
