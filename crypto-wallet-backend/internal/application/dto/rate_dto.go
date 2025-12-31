package dto

import (
	"time"

	"github.com/google/uuid"
)

// ExchangeRate represents the current exchange rate for a cryptocurrency.
type ExchangeRate struct {
	Symbol         string    `json:"symbol"`
	PriceUSD       string    `json:"price_usd"`
	PriceChange24h string    `json:"price_change_24h"`
	Volume24h      string    `json:"volume_24h,omitempty"`
	MarketCap      string    `json:"market_cap,omitempty"`
	LastUpdated    time.Time `json:"last_updated"`
}

// ExchangeRateList groups a collection of exchange rates.
type ExchangeRateList struct {
	Rates       []ExchangeRate `json:"rates"`
	LastUpdated time.Time      `json:"last_updated"`
}

// GetRatesRequest models the request for fetching exchange rates.
type GetRatesRequest struct {
	Symbols []string `json:"symbols,omitempty"`
}

// PriceHistory represents historical OHLCV data for a cryptocurrency.
type PriceHistory struct {
	ID        uuid.UUID `json:"id"`
	Symbol    string    `json:"symbol"`
	Timestamp time.Time `json:"timestamp"`
	Open      string    `json:"open"`
	High      string    `json:"high"`
	Low       string    `json:"low"`
	Close     string    `json:"close"`
	Volume    string    `json:"volume"`
}

// PriceHistoryList groups a collection of price history data points.
type PriceHistoryList struct {
	Symbol   string         `json:"symbol"`
	Interval string         `json:"interval"`
	Data     []PriceHistory `json:"data"`
}

// GetPriceHistoryRequest models the request for fetching historical price data.
type GetPriceHistoryRequest struct {
	Symbol   string     `json:"symbol" form:"symbol" validate:"required"`
	Interval string     `json:"interval,omitempty" form:"interval"`
	From     *time.Time `json:"from,omitempty" form:"from"`
	To       *time.Time `json:"to,omitempty" form:"to"`
	Limit    int        `json:"limit,omitempty" form:"limit"`
	Offset   int        `json:"offset,omitempty" form:"offset"`
}
