package external

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (
	coinGeckoAPIBaseURL     = "https://api.coingecko.com/api/v3"
	defaultRequestTimeout   = 10 * time.Second
	defaultRetryAttempts    = 3
	defaultRetryDelay       = 1 * time.Second
)

var (
	ErrInvalidAPIKey      = errors.New("coingecko: invalid API key")
	ErrRateLimitExceeded  = errors.New("coingecko: rate limit exceeded")
	ErrCoinNotFound       = errors.New("coingecko: coin not found")
	ErrInvalidResponse    = errors.New("coingecko: invalid response format")
	ErrNetworkTimeout     = errors.New("coingecko: network timeout")
)

// CoinGeckoSymbolMap maps our internal symbols to CoinGecko coin IDs.
var CoinGeckoSymbolMap = map[string]string{
	"BTC": "bitcoin",
	"ETH": "ethereum",
	"SOL": "solana",
	"XLM": "stellar",
}

// CoinGeckoPriceData represents the price data returned by CoinGecko API.
type CoinGeckoPriceData struct {
	Symbol                string
	PriceUSD              decimal.Decimal
	PriceChange24h        decimal.Decimal
	PriceChangePercent24h decimal.Decimal
	Volume24h             decimal.Decimal
	MarketCap             decimal.Decimal
	LastUpdated           time.Time
}

// CoinGeckoClient provides methods for interacting with the CoinGecko API.
type CoinGeckoClient interface {
	// GetPrices fetches current prices for the specified symbols.
	GetPrices(ctx context.Context, symbols []string) (map[string]*CoinGeckoPriceData, error)

	// GetPrice fetches current price for a single symbol.
	GetPrice(ctx context.Context, symbol string) (*CoinGeckoPriceData, error)

	// GetHistoricalPrices fetches historical OHLCV data for a symbol.
	GetHistoricalPrices(ctx context.Context, symbol string, days int) ([]OHLCVData, error)
}

// OHLCVData represents Open-High-Low-Close-Volume candle data.
type OHLCVData struct {
	Timestamp time.Time
	Open      decimal.Decimal
	High      decimal.Decimal
	Low       decimal.Decimal
	Close     decimal.Decimal
	Volume    decimal.Decimal
}

// coinGeckoClientImpl implements the CoinGeckoClient interface.
type coinGeckoClientImpl struct {
	httpClient    *http.Client
	apiKey        string
	logger        *slog.Logger
	retryAttempts int
	retryDelay    time.Duration
}

// CoinGeckoConfig holds configuration for the CoinGecko client.
type CoinGeckoConfig struct {
	APIKey        string
	Timeout       time.Duration
	RetryAttempts int
	RetryDelay    time.Duration
	Logger        *slog.Logger
}

// NewCoinGeckoClient creates a new CoinGecko API client.
func NewCoinGeckoClient(config CoinGeckoConfig) CoinGeckoClient {
	if config.Timeout == 0 {
		config.Timeout = defaultRequestTimeout
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = defaultRetryAttempts
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = defaultRetryDelay
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	return &coinGeckoClientImpl{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		apiKey:        config.APIKey,
		logger:        config.Logger,
		retryAttempts: config.RetryAttempts,
		retryDelay:    config.RetryDelay,
	}
}

// GetPrices fetches current prices for multiple symbols.
func (c *coinGeckoClientImpl) GetPrices(ctx context.Context, symbols []string) (map[string]*CoinGeckoPriceData, error) {
	if len(symbols) == 0 {
		return make(map[string]*CoinGeckoPriceData), nil
	}

	// Convert symbols to CoinGecko coin IDs
	coinIDs := make([]string, 0, len(symbols))
	symbolToCoinID := make(map[string]string)

	for _, symbol := range symbols {
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if coinID, ok := CoinGeckoSymbolMap[symbol]; ok {
			coinIDs = append(coinIDs, coinID)
			symbolToCoinID[coinID] = symbol
		} else {
			c.logger.Warn("Unknown symbol for CoinGecko", "symbol", symbol)
		}
	}

	if len(coinIDs) == 0 {
		return make(map[string]*CoinGeckoPriceData), nil
	}

	// Build API URL
	ids := strings.Join(coinIDs, ",")
	apiURL := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd&include_market_cap=true&include_24hr_vol=true&include_24hr_change=true",
		coinGeckoAPIBaseURL, url.QueryEscape(ids))

	// Make API request with retry logic
	var response map[string]map[string]interface{}
	err := c.doRequestWithRetry(ctx, apiURL, &response)
	if err != nil {
		return nil, err
	}

	// Parse response
	results := make(map[string]*CoinGeckoPriceData)
	now := time.Now().UTC()

	for coinID, data := range response {
		symbol, ok := symbolToCoinID[coinID]
		if !ok {
			continue
		}

		priceData, err := c.parseSimplePriceResponse(symbol, data, now)
		if err != nil {
			c.logger.Error("Failed to parse price data", "symbol", symbol, "error", err)
			continue
		}

		results[symbol] = priceData
	}

	return results, nil
}

// GetPrice fetches current price for a single symbol.
func (c *coinGeckoClientImpl) GetPrice(ctx context.Context, symbol string) (*CoinGeckoPriceData, error) {
	prices, err := c.GetPrices(ctx, []string{symbol})
	if err != nil {
		return nil, err
	}

	priceData, ok := prices[strings.ToUpper(symbol)]
	if !ok {
		return nil, ErrCoinNotFound
	}

	return priceData, nil
}

// GetHistoricalPrices fetches OHLC data for a symbol.
func (c *coinGeckoClientImpl) GetHistoricalPrices(ctx context.Context, symbol string, days int) ([]OHLCVData, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	coinID, ok := CoinGeckoSymbolMap[symbol]
	if !ok {
		return nil, fmt.Errorf("unknown symbol: %s", symbol)
	}

	// CoinGecko OHLC endpoint
	apiURL := fmt.Sprintf("%s/coins/%s/ohlc?vs_currency=usd&days=%d",
		coinGeckoAPIBaseURL, coinID, days)

	// Response is array of [timestamp, open, high, low, close]
	var response [][]float64
	err := c.doRequestWithRetry(ctx, apiURL, &response)
	if err != nil {
		return nil, err
	}

	// Parse OHLC data
	results := make([]OHLCVData, 0, len(response))
	for _, candle := range response {
		if len(candle) != 5 {
			continue
		}

		ohlcv := OHLCVData{
			Timestamp: time.Unix(int64(candle[0])/1000, 0).UTC(),
			Open:      decimal.NewFromFloat(candle[1]),
			High:      decimal.NewFromFloat(candle[2]),
			Low:       decimal.NewFromFloat(candle[3]),
			Close:     decimal.NewFromFloat(candle[4]),
			Volume:    decimal.Zero, // OHLC endpoint doesn't include volume
		}

		results = append(results, ohlcv)
	}

	return results, nil
}

func (c *coinGeckoClientImpl) doRequestWithRetry(ctx context.Context, url string, response interface{}) error {
	var lastErr error

	for attempt := 0; attempt < c.retryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.retryDelay):
			}
		}

		err := c.doRequest(ctx, url, response)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on certain errors
		if errors.Is(err, ErrInvalidAPIKey) || errors.Is(err, ErrCoinNotFound) {
			return err
		}

		c.logger.Warn("CoinGecko API request failed, retrying",
			"attempt", attempt+1,
			"max_attempts", c.retryAttempts,
			"error", err)
	}

	return fmt.Errorf("failed after %d attempts: %w", c.retryAttempts, lastErr)
}

func (c *coinGeckoClientImpl) doRequest(ctx context.Context, url string, response interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Add API key if provided
	if c.apiKey != "" {
		req.Header.Set("x-cg-pro-api-key", c.apiKey)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return ErrNetworkTimeout
		}
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	switch resp.StatusCode {
	case http.StatusOK:
		// Success, continue parsing
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrInvalidAPIKey
	case http.StatusTooManyRequests:
		return ErrRateLimitExceeded
	case http.StatusNotFound:
		return ErrCoinNotFound
	default:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	return nil
}

func (c *coinGeckoClientImpl) parseSimplePriceResponse(symbol string, data map[string]interface{}, timestamp time.Time) (*CoinGeckoPriceData, error) {
	priceData := &CoinGeckoPriceData{
		Symbol:      symbol,
		LastUpdated: timestamp,
	}

	// Parse USD price
	if usdPrice, ok := data["usd"].(float64); ok {
		priceData.PriceUSD = decimal.NewFromFloat(usdPrice)
	} else {
		return nil, fmt.Errorf("missing usd price")
	}

	// Parse 24h change percentage
	if change24h, ok := data["usd_24h_change"].(float64); ok {
		priceData.PriceChangePercent24h = decimal.NewFromFloat(change24h)
		// Calculate absolute change from percentage
		priceData.PriceChange24h = priceData.PriceUSD.Mul(priceData.PriceChangePercent24h).Div(decimal.NewFromInt(100))
	}

	// Parse 24h volume
	if volume24h, ok := data["usd_24h_vol"].(float64); ok {
		priceData.Volume24h = decimal.NewFromFloat(volume24h)
	}

	// Parse market cap
	if marketCap, ok := data["usd_market_cap"].(float64); ok {
		priceData.MarketCap = decimal.NewFromFloat(marketCap)
	}

	return priceData, nil
}
