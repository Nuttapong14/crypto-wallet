package blockchain

import (
	"context"
	"log/slog"
	"time"
)

// RetryConfig controls blockchain retry behaviour.
type RetryConfig struct {
	Attempts int
	Delay    time.Duration
}

func (cfg RetryConfig) normalize() RetryConfig {
	normalized := cfg
	if normalized.Attempts <= 0 {
		normalized.Attempts = 3
	}
	if normalized.Delay <= 0 {
		normalized.Delay = 250 * time.Millisecond
	}
	return normalized
}

// Retry executes fn with simple linear back-off and returns the result.
func Retry[T any](ctx context.Context, logger *slog.Logger, cfg RetryConfig, operation string, fn func(context.Context) (T, error)) (T, error) {
	var zero T
	normalized := cfg.normalize()

	for attempt := 1; attempt <= normalized.Attempts; attempt++ {
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}

		result, err := fn(ctx)
		if err == nil {
			if logger != nil {
				logger.Debug("blockchain operation succeeded", slog.String("operation", operation), slog.Int("attempt", attempt))
			}
			return result, nil
		}

		if logger != nil {
			logger.Warn("blockchain operation failed",
				slog.String("operation", operation),
				slog.Int("attempt", attempt),
				slog.String("error", err.Error()),
			)
		}

		if attempt < normalized.Attempts {
			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(normalized.Delay):
			}
		}

		// If this was the final attempt, return the error.
		if attempt == normalized.Attempts {
			return zero, err
		}
	}

	return zero, context.Canceled
}

// retry executes fn with simple linear back-off and returns the result.
func retry[T any](ctx context.Context, logger *slog.Logger, cfg RetryConfig, operation string, fn func(context.Context) (T, error)) (T, error) {
	return Retry(ctx, logger, cfg, operation, fn)
}
