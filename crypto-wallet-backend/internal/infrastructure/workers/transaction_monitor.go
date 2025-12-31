package workers

import (
    "context"
    "log/slog"
    "time"

    "github.com/crypto-wallet/backend/internal/domain/entities"
    "github.com/crypto-wallet/backend/internal/domain/repositories"
    "github.com/crypto-wallet/backend/internal/infrastructure/blockchain"
)

// TransactionMonitor periodically queries blockchain adapters for transaction confirmations.
type TransactionMonitor struct {
    repository repositories.TransactionRepository
    adapters   map[entities.Chain]blockchain.BlockchainAdapter
    interval   time.Duration
    logger     *slog.Logger
}

// NewTransactionMonitor constructs a monitor with sane defaults.
func NewTransactionMonitor(repo repositories.TransactionRepository, adapters map[entities.Chain]blockchain.BlockchainAdapter, interval time.Duration, logger *slog.Logger) *TransactionMonitor {
    if interval <= 0 {
        interval = 10 * time.Second
    }
    if logger == nil {
        logger = slog.Default()
    }
    return &TransactionMonitor{
        repository: repo,
        adapters:   adapters,
        interval:   interval,
        logger:     logger.With(slog.String("component", "transaction_monitor")),
    }
}

// Run executes a single monitoring loop; callers are responsible for scheduling.
func (m *TransactionMonitor) Run(ctx context.Context) {
    if m.repository == nil || len(m.adapters) == 0 {
        m.logger.Warn("transaction monitor misconfigured; skipping execution")
        return
    }

    ticker := time.NewTicker(m.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            m.logger.Info("transaction monitor exiting", slog.String("reason", ctx.Err().Error()))
            return
        case <-ticker.C:
            m.logger.Debug("transaction monitor tick")
            // Stubs: in a full implementation we would pull pending transactions from the
            // repository, query the appropriate blockchain adapter, and persist status updates.
        }
    }
}
