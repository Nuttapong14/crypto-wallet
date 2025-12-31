package workers

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	analyticsusecase "github.com/crypto-wallet/backend/internal/application/usecases/analytics"
	"github.com/crypto-wallet/backend/internal/application/dto"
)

// PortfolioCalculatorConfig configures the portfolio calculator worker.
type PortfolioCalculatorConfig struct {
	SummaryUseCase     *analyticsusecase.PortfolioSummaryUseCase
	PerformanceUseCase *analyticsusecase.PortfolioPerformanceUseCase
	Logger             *slog.Logger
}

// PortfolioCalculator recalculates portfolio metrics for background workflows.
type PortfolioCalculator struct {
	summaryUC     *analyticsusecase.PortfolioSummaryUseCase
	performanceUC *analyticsusecase.PortfolioPerformanceUseCase
	logger        *slog.Logger
}

// NewPortfolioCalculator constructs a portfolio calculator worker.
func NewPortfolioCalculator(cfg PortfolioCalculatorConfig) *PortfolioCalculator {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &PortfolioCalculator{
		summaryUC:     cfg.SummaryUseCase,
		performanceUC: cfg.PerformanceUseCase,
		logger:        logger,
	}
}

// Recalculate recomputes portfolio summary and performance for the supplied user.
func (w *PortfolioCalculator) Recalculate(ctx context.Context, userID uuid.UUID, period string) (dto.PortfolioSummary, dto.PortfolioPerformance, error) {
	var summary dto.PortfolioSummary
	var performance dto.PortfolioPerformance
	var err error

	if w.summaryUC != nil {
		summary, err = w.summaryUC.Execute(ctx, userID)
		if err != nil {
			w.logger.Error("failed to recompute portfolio summary", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
			return dto.PortfolioSummary{}, dto.PortfolioPerformance{}, err
		}
	}

	if w.performanceUC != nil {
		performance, err = w.performanceUC.Execute(ctx, userID, period)
		if err != nil {
			w.logger.Error("failed to recompute portfolio performance", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
			return summary, dto.PortfolioPerformance{}, err
		}
	}

	return summary, performance, nil
}
