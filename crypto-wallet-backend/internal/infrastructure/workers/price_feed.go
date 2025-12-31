package workers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
	"github.com/crypto-wallet/backend/internal/infrastructure/external"
	"github.com/crypto-wallet/backend/internal/infrastructure/messaging"
)

const (
	// Default configuration
	defaultFetchInterval  = 5 * time.Second  // Fetch prices every 5 seconds (meets <5s latency requirement)
	defaultRetryDelay     = 2 * time.Second
	defaultMaxRetries     = 3
	defaultStaleThreshold = 30 * time.Second // Consider data stale if >30s old
)

// PriceFeedWorker periodically fetches cryptocurrency prices and broadcasts them.
type PriceFeedWorker struct {
	coinGeckoClient external.CoinGeckoClient
	pubSubManager   messaging.RedisPubSubManager
	rateRepository  repositories.RateRepository
	logger          *slog.Logger
	symbols         []string
	fetchInterval   time.Duration
	retryDelay      time.Duration
	maxRetries      int
	stopCh          chan struct{}
	doneCh          chan struct{}
}

// PriceFeedWorkerConfig holds configuration for the price feed worker.
type PriceFeedWorkerConfig struct {
	CoinGeckoClient external.CoinGeckoClient
	PubSubManager   messaging.RedisPubSubManager
	RateRepository  repositories.RateRepository
	Logger          *slog.Logger
	Symbols         []string
	FetchInterval   time.Duration
	RetryDelay      time.Duration
	MaxRetries      int
}

// NewPriceFeedWorker creates a new price feed worker.
func NewPriceFeedWorker(config PriceFeedWorkerConfig) *PriceFeedWorker {
	if config.FetchInterval == 0 {
		config.FetchInterval = defaultFetchInterval
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = defaultRetryDelay
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = defaultMaxRetries
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	if len(config.Symbols) == 0 {
		config.Symbols = []string{"BTC", "ETH", "SOL", "XLM"}
	}

	return &PriceFeedWorker{
		coinGeckoClient: config.CoinGeckoClient,
		pubSubManager:   config.PubSubManager,
		rateRepository:  config.RateRepository,
		logger:          config.Logger,
		symbols:         config.Symbols,
		fetchInterval:   config.FetchInterval,
		retryDelay:      config.RetryDelay,
		maxRetries:      config.MaxRetries,
		stopCh:          make(chan struct{}),
		doneCh:          make(chan struct{}),
	}
}

// Start begins the price feed worker loop.
func (w *PriceFeedWorker) Start(ctx context.Context) error {
	w.logger.Info("Starting price feed worker",
		"symbols", w.symbols,
		"fetch_interval", w.fetchInterval)

	// Fetch initial prices immediately
	if err := w.fetchAndBroadcastPrices(ctx); err != nil {
		w.logger.Error("Initial price fetch failed", "error", err)
	}

	// Start periodic price fetching
	ticker := time.NewTicker(w.fetchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Price feed worker stopped due to context cancellation")
			close(w.doneCh)
			return ctx.Err()
		case <-w.stopCh:
			w.logger.Info("Price feed worker stopped")
			close(w.doneCh)
			return nil
		case <-ticker.C:
			if err := w.fetchAndBroadcastPrices(ctx); err != nil {
				w.logger.Error("Failed to fetch and broadcast prices", "error", err)
			}
		}
	}
}

// Stop signals the worker to stop gracefully.
func (w *PriceFeedWorker) Stop() {
	w.logger.Info("Stopping price feed worker")
	close(w.stopCh)
	<-w.doneCh
}

// fetchAndBroadcastPrices fetches prices from CoinGecko and broadcasts them via Redis Pub/Sub.
func (w *PriceFeedWorker) fetchAndBroadcastPrices(ctx context.Context) error {
	startTime := time.Now()

	// Fetch prices from CoinGecko with retry logic
	prices, err := w.fetchPricesWithRetry(ctx)
	if err != nil {
		return fmt.Errorf("fetch prices: %w", err)
	}

	if len(prices) == 0 {
		w.logger.Warn("No prices fetched from CoinGecko")
		return nil
	}

	// Store prices in database
	if err := w.storePricesInDatabase(ctx, prices); err != nil {
		w.logger.Error("Failed to store prices in database", "error", err)
		// Continue with broadcast even if database storage fails
	}

	// Broadcast prices via Redis Pub/Sub
	if err := w.broadcastPrices(ctx, prices); err != nil {
		w.logger.Error("Failed to broadcast prices", "error", err)
		// Don't return error, prices are still stored in database
	}

	duration := time.Since(startTime)
	w.logger.Info("Price fetch and broadcast completed",
		"price_count", len(prices),
		"duration_ms", duration.Milliseconds())

	return nil
}

// fetchPricesWithRetry fetches prices with retry logic.
func (w *PriceFeedWorker) fetchPricesWithRetry(ctx context.Context) (map[string]*external.CoinGeckoPriceData, error) {
	var prices map[string]*external.CoinGeckoPriceData
	var lastErr error

	for attempt := 0; attempt < w.maxRetries; attempt++ {
		if attempt > 0 {
			w.logger.Warn("Retrying price fetch",
				"attempt", attempt+1,
				"max_attempts", w.maxRetries)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(w.retryDelay):
			}
		}

		var err error
		prices, err = w.coinGeckoClient.GetPrices(ctx, w.symbols)
		if err == nil {
			return prices, nil
		}

		lastErr = err
		w.logger.Error("Price fetch attempt failed",
			"attempt", attempt+1,
			"error", err)
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", w.maxRetries, lastErr)
}

// storePricesInDatabase persists prices to the rates database.
func (w *PriceFeedWorker) storePricesInDatabase(ctx context.Context, prices map[string]*external.CoinGeckoPriceData) error {
	for symbol, priceData := range prices {
		// Create ExchangeRate entity
		rateEntity, err := entities.NewExchangeRateEntity(entities.ExchangeRateParams{
			ID:             uuid.New(),
			Symbol:         symbol,
			PriceUSD:       priceData.PriceUSD,
			PriceChange24h: priceData.PriceChange24h,
			Volume24h:      priceData.Volume24h,
			MarketCap:      priceData.MarketCap,
			LastUpdated:    priceData.LastUpdated,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		})
		if err != nil {
			w.logger.Error("Failed to create exchange rate entity",
				"symbol", symbol,
				"error", err)
			continue
		}

		// Upsert to database (insert or update)
		if err := w.rateRepository.UpsertRate(ctx, rateEntity); err != nil {
			w.logger.Error("Failed to upsert exchange rate",
				"symbol", symbol,
				"error", err)
			continue
		}

		w.logger.Debug("Stored exchange rate in database", "symbol", symbol)
	}

	return nil
}

// broadcastPrices publishes prices to Redis Pub/Sub channels.
func (w *PriceFeedWorker) broadcastPrices(ctx context.Context, prices map[string]*external.CoinGeckoPriceData) error {
	// Prepare batch message
	batchPrices := make([]*messaging.PriceUpdateMessage, 0, len(prices))

	// Publish individual price updates and collect for batch
	for symbol, priceData := range prices {
		priceMsg := &messaging.PriceUpdateMessage{
			Symbol:         symbol,
			PriceUSD:       priceData.PriceUSD.String(),
			PriceChange24h: priceData.PriceChange24h.String(),
			Volume24h:      priceData.Volume24h.String(),
			Timestamp:      priceData.LastUpdated.Format(time.RFC3339),
		}

		// Publish to symbol-specific channel (prices:BTC, prices:ETH, etc.)
		if err := w.pubSubManager.PublishPrice(ctx, symbol, priceMsg); err != nil {
			w.logger.Error("Failed to publish price update",
				"symbol", symbol,
				"error", err)
			continue
		}

		batchPrices = append(batchPrices, priceMsg)
	}

	// Publish batch update to batch channel
	if len(batchPrices) > 0 {
		if err := w.pubSubManager.PublishBatchPrices(ctx, batchPrices); err != nil {
			w.logger.Error("Failed to publish batch price update", "error", err)
			return err
		}
	}

	return nil
}

// GetCurrentPrices returns the most recent prices from the database.
func (w *PriceFeedWorker) GetCurrentPrices(ctx context.Context) (map[string]*external.CoinGeckoPriceData, error) {
	rates, err := w.rateRepository.GetRatesBySymbols(ctx, w.symbols)
	if err != nil {
		return nil, fmt.Errorf("get rates from database: %w", err)
	}

	prices := make(map[string]*external.CoinGeckoPriceData)
	for _, rate := range rates {
		prices[rate.GetSymbol()] = &external.CoinGeckoPriceData{
			Symbol:         rate.GetSymbol(),
			PriceUSD:       rate.GetPriceUSD(),
			PriceChange24h: rate.GetPriceChange24h(),
			Volume24h:      rate.GetVolume24h(),
			MarketCap:      rate.GetMarketCap(),
			LastUpdated:    rate.GetLastUpdated(),
		}
	}

	return prices, nil
}

// IsDataStale checks if the price data is stale (older than threshold).
func (w *PriceFeedWorker) IsDataStale(lastUpdated time.Time) bool {
	return time.Since(lastUpdated) > defaultStaleThreshold
}
