package dto

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// formatFloat formats a float64 as a string with 2 decimal places
func formatFloat(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

// GetTransactionHistoryRequest captures query parameters for transaction history.
type GetTransactionHistoryRequest struct {
	WalletID  string `json:"walletId,omitempty"`
	Chain     string `json:"chain,omitempty"`
	Type      string `json:"type,omitempty"`
	Status    string `json:"status,omitempty"`
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

// Validate enforces request invariants.
func (r GetTransactionHistoryRequest) Validate() utils.ValidationErrors {
	errs := utils.ValidationErrors{}

	if r.WalletID != "" {
		utils.RequireUUID(&errs, "walletId", r.WalletID)
	}

	if r.Chain != "" {
		utils.Require(&errs, "chain", r.Chain)
	}

	if r.Type != "" {
		utils.Require(&errs, "type", r.Type)
	}

	if r.Status != "" {
		utils.Require(&errs, "status", r.Status)
	}

	if r.StartDate != "" {
		if _, err := time.Parse(time.RFC3339, r.StartDate); err != nil {
			errs.Add("startDate", "must be a valid RFC3339 datetime")
		}
	}

	if r.EndDate != "" {
		if _, err := time.Parse(time.RFC3339, r.EndDate); err != nil {
			errs.Add("endDate", "must be a valid RFC3339 datetime")
		}
	}

	if r.Limit <= 0 {
		r.Limit = 50 // default limit
	} else if r.Limit > 1000 {
		errs.Add("limit", "cannot exceed 1000")
	}

	if r.Offset < 0 {
		errs.Add("offset", "cannot be negative")
	}

	return errs
}

// ExportTransactionsRequest captures query parameters for transaction export.
type ExportTransactionsRequest struct {
	WalletID  string `json:"walletId,omitempty"`
	Chain     string `json:"chain,omitempty"`
	Type      string `json:"type,omitempty"`
	Status    string `json:"status,omitempty"`
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
	Format    string `json:"format"` // csv, json
}

// Validate enforces request invariants.
func (r ExportTransactionsRequest) Validate() utils.ValidationErrors {
	errs := utils.ValidationErrors{}

	if r.WalletID != "" {
		utils.RequireUUID(&errs, "walletId", r.WalletID)
	}

	if r.Chain != "" {
		utils.Require(&errs, "chain", r.Chain)
	}

	if r.Type != "" {
		utils.Require(&errs, "type", r.Type)
	}

	if r.Status != "" {
		utils.Require(&errs, "status", r.Status)
	}

	if r.StartDate != "" {
		if _, err := time.Parse(time.RFC3339, r.StartDate); err != nil {
			errs.Add("startDate", "must be a valid RFC3339 datetime")
		}
	}

	if r.EndDate != "" {
		if _, err := time.Parse(time.RFC3339, r.EndDate); err != nil {
			errs.Add("endDate", "must be a valid RFC3339 datetime")
		}
	}

	if r.Format == "" {
		r.Format = "csv" // default format
	} else if r.Format != "csv" && r.Format != "json" {
		errs.Add("format", "must be either 'csv' or 'json'")
	}

	return errs
}

// GetTransactionAnalyticsRequest captures query parameters for transaction analytics.
type GetTransactionAnalyticsRequest struct {
	WalletID  string `json:"walletId,omitempty"`
	Chain     string `json:"chain,omitempty"`
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
	Period    string `json:"period"` // daily, weekly, monthly
}

// Validate enforces request invariants.
func (r GetTransactionAnalyticsRequest) Validate() utils.ValidationErrors {
	errs := utils.ValidationErrors{}

	if r.WalletID != "" {
		utils.RequireUUID(&errs, "walletId", r.WalletID)
	}

	if r.Chain != "" {
		utils.Require(&errs, "chain", r.Chain)
	}

	if r.StartDate != "" {
		if _, err := time.Parse(time.RFC3339, r.StartDate); err != nil {
			errs.Add("startDate", "must be a valid RFC3339 datetime")
		}
	}

	if r.EndDate != "" {
		if _, err := time.Parse(time.RFC3339, r.EndDate); err != nil {
			errs.Add("endDate", "must be a valid RFC3339 datetime")
		}
	}

	if r.Period == "" {
		r.Period = "daily" // default period
	} else if r.Period != "daily" && r.Period != "weekly" && r.Period != "monthly" {
		errs.Add("period", "must be either 'daily', 'weekly', or 'monthly'")
	}

	return errs
}

// TransactionAnalyticsResponse provides aggregated transaction analytics.
type TransactionAnalyticsResponse struct {
	WalletID               *uuid.UUID               `json:"walletId,omitempty"`
	Chain                  *string                  `json:"chain,omitempty"`
	Period                 string                   `json:"period"`
	StartDate              string                   `json:"startDate"`
	EndDate                string                   `json:"endDate"`
	TotalTransactions      int64                    `json:"totalTransactions"`
	TotalVolume            string                   `json:"totalVolume"`
	AverageTransactionSize string                   `json:"averageTransactionSize"`
	SwapTransactions       int64                    `json:"swapTransactions"`
	SwapPercentage         float64                  `json:"swapPercentage"`
	TotalFees              string                   `json:"totalFees"`
	AverageFee             string                   `json:"averageFee"`
	FeePercentage          float64                  `json:"feePercentage"`
	DailyData              []DailyAnalyticsResponse `json:"dailyData,omitempty"`
}

// DailyAnalyticsResponse provides daily breakdown of analytics.
type DailyAnalyticsResponse struct {
	Date                   string  `json:"date"`
	TransactionCount       int64   `json:"transactionCount"`
	Volume                 string  `json:"volume"`
	AverageTransactionSize string  `json:"averageTransactionSize"`
	SwapCount              int64   `json:"swapCount"`
	SwapPercentage         float64 `json:"swapPercentage"`
	TotalFees              string  `json:"totalFees"`
	AverageFee             string  `json:"averageFee"`
	FeePercentage          float64 `json:"feePercentage"`
}

// NewTransactionAnalyticsResponse maps domain entities to API response.
func NewTransactionAnalyticsResponse(summary entities.TransactionSummary, dailyData []entities.TransactionSummary, period string, startDate, endDate time.Time) TransactionAnalyticsResponse {
	chain := string(summary.GetChain())
	response := TransactionAnalyticsResponse{
		Chain:                  &chain,
		Period:                 period,
		StartDate:              startDate.UTC().Format(time.RFC3339Nano),
		EndDate:                endDate.UTC().Format(time.RFC3339Nano),
		TotalTransactions:      summary.GetTotalTransactions(),
		TotalVolume:            formatFloat(summary.GetTotalVolume()),
		AverageTransactionSize: formatFloat(summary.GetAverageTransactionSize()),
		SwapTransactions:       summary.GetSwapTransactions(),
		SwapPercentage:         summary.GetSwapPercentage(),
		TotalFees:              formatFloat(summary.GetTotalFees()),
		AverageFee:             formatFloat(summary.GetAverageFee()),
		FeePercentage:          summary.GetFeePercentage(),
	}

	if len(dailyData) > 0 {
		response.DailyData = make([]DailyAnalyticsResponse, len(dailyData))
		for i, daily := range dailyData {
			response.DailyData[i] = DailyAnalyticsResponse{
				Date:                   daily.GetDate().Format(time.RFC3339Nano),
				TransactionCount:       daily.GetTotalTransactions(),
				Volume:                 formatFloat(daily.GetTotalVolume()),
				AverageTransactionSize: formatFloat(daily.GetAverageTransactionSize()),
				SwapCount:              daily.GetSwapTransactions(),
				SwapPercentage:         daily.GetSwapPercentage(),
				TotalFees:              formatFloat(daily.GetTotalFees()),
				AverageFee:             formatFloat(daily.GetAverageFee()),
				FeePercentage:          daily.GetFeePercentage(),
			}
		}
	}

	return response
}

// ExportResponse provides export download information.
type ExportResponse struct {
	DownloadURL string `json:"downloadUrl"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	Format      string `json:"format"`
	CreatedAt   string `json:"createdAt"`
}

// PortfolioAsset represents allocation information for a single asset.
type PortfolioAsset struct {
	Symbol     string `json:"symbol"`
	Balance    string `json:"balance"`
	BalanceUSD string `json:"balance_usd"`
	Percentage string `json:"percentage"`
}

// PortfolioSummary aggregates portfolio value and allocation details.
type PortfolioSummary struct {
	TotalBalanceUSD            string           `json:"total_balance_usd"`
	TotalChange24h             string           `json:"total_change_24h"`
	TotalChangePercentage24h   string           `json:"total_change_percentage_24h"`
	Assets                     []PortfolioAsset `json:"assets"`
}

// PortfolioPerformancePoint represents a historical portfolio value datapoint.
type PortfolioPerformancePoint struct {
	Timestamp string `json:"timestamp"`
	ValueUSD  string `json:"value_usd"`
}

// PortfolioPerformance summarises historical portfolio performance for a selected period.
type PortfolioPerformance struct {
	Period             string                       `json:"period"`
	InitialValueUSD    string                       `json:"initial_value_usd"`
	FinalValueUSD      string                       `json:"final_value_usd"`
	GainLossUSD        string                       `json:"gain_loss_usd"`
	GainLossPercentage string                       `json:"gain_loss_percentage"`
	DataPoints         []PortfolioPerformancePoint  `json:"data_points"`
}
