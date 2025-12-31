package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/services"
)

// QuoteExpirationWorker handles the expiration of pending quotes.
type QuoteExpirationWorker struct {
	exchangeService *services.ExchangeService
	logger          *slog.Logger
	interval        time.Duration
	ticker          *time.Ticker
	stopChan        chan struct{}
}

// NewQuoteExpirationWorker creates a new QuoteExpirationWorker.
func NewQuoteExpirationWorker(
	exchangeService *services.ExchangeService,
	logger *slog.Logger,
	interval time.Duration,
) *QuoteExpirationWorker {
	return &QuoteExpirationWorker{
		exchangeService: exchangeService,
		logger:          logger,
		interval:        interval,
		stopChan:        make(chan struct{}),
	}
}

// Start begins the quote expiration worker.
func (w *QuoteExpirationWorker) Start(ctx context.Context) {
	w.logger.Info("Starting quote expiration worker", "interval", w.interval)

	w.ticker = time.NewTicker(w.interval)
	defer w.ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Quote expiration worker stopped by context")
			return
		case <-w.stopChan:
			w.logger.Info("Quote expiration worker stopped by signal")
			return
		case <-w.ticker.C:
			w.expireQuotes(ctx)
		}
	}
}

// Stop stops the quote expiration worker.
func (w *QuoteExpirationWorker) Stop() {
	w.logger.Info("Stopping quote expiration worker")
	close(w.stopChan)
	if w.ticker != nil {
		w.ticker.Stop()
	}
}

// expireQuotes processes the expiration of pending quotes.
func (w *QuoteExpirationWorker) expireQuotes(ctx context.Context) {
	w.logger.Debug("Checking for expired quotes")

	expiredOperations, err := w.exchangeService.ExpirePendingQuotes(ctx)
	if err != nil {
		w.logger.Error("Failed to expire pending quotes", "error", err)
		return
	}

	if len(expiredOperations) > 0 {
		w.logger.Info("Expired pending quotes",
			"count", len(expiredOperations),
			"operations", w.getOperationIDs(expiredOperations))
	}
}

// getOperationIDs extracts operation IDs for logging.
func (w *QuoteExpirationWorker) getOperationIDs(operations []entities.ExchangeOperation) []string {
	ids := make([]string, len(operations))
	for i, op := range operations {
		ids[i] = op.GetID()
	}
	return ids
}
