package analytics

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
	appLogging "github.com/crypto-wallet/backend/internal/infrastructure/logging"
	"github.com/crypto-wallet/backend/pkg/utils"
)

var (
	errPerformanceWalletRepo = errors.New("portfolio performance: wallet repository not configured")
	errPerformanceRateRepo   = errors.New("portfolio performance: rate repository not configured")
)

type periodConfig struct {
	label    string
	duration time.Duration
	interval entities.IntervalType
}

// PortfolioPerformanceUseCase calculates historical portfolio performance.
type PortfolioPerformanceUseCase struct {
	wallets repositories.WalletRepository
	rates   repositories.RateRepository
	logger  *slog.Logger
	now     func() time.Time
}

// NewPortfolioPerformanceUseCase constructs the use case.
func NewPortfolioPerformanceUseCase(wallets repositories.WalletRepository, rates repositories.RateRepository, logger *slog.Logger) *PortfolioPerformanceUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &PortfolioPerformanceUseCase{
		wallets: wallets,
		rates:   rates,
		logger:  logger,
		now:     func() time.Time { return time.Now().UTC() },
	}
}

// Execute returns the portfolio performance for the provided user and period identifier.
func (uc *PortfolioPerformanceUseCase) Execute(ctx context.Context, userID uuid.UUID, period string) (dto.PortfolioPerformance, error) {
	if uc.wallets == nil {
		return dto.PortfolioPerformance{}, errPerformanceWalletRepo
	}
	if uc.rates == nil {
		return dto.PortfolioPerformance{}, errPerformanceRateRepo
	}
	if userID == uuid.Nil {
		return dto.PortfolioPerformance{}, utils.NewAppError(
			"VALIDATION_ERROR",
			"user id is required",
			fiber.StatusBadRequest,
			nil,
			nil,
		)
	}

	config := resolvePeriod(period)
	if period == "" {
		period = config.label
	}

	ctxLogger := appLogging.LoggerFromContext(ctx, uc.logger).With(
		slog.String("user_id", userID.String()),
		slog.String("period", period),
	)

	wallets, err := uc.wallets.ListByUser(ctx, userID, repositories.WalletFilter{}, repositories.ListOptions{Limit: 1000, SortBy: "created_at", SortOrder: repositories.SortDescending})
	if err != nil {
		ctxLogger.Error("failed to list wallets for portfolio performance", slog.String("error", err.Error()))
        return dto.PortfolioPerformance{}, utils.NewAppError(
            "DATABASE_ERROR",
            "unable to load wallets",
            fiber.StatusInternalServerError,
            err,
			map[string]any{"userId": userID.String()},
		)
	}

	if len(wallets) == 0 {
		return dto.PortfolioPerformance{
			Period:             config.label,
			InitialValueUSD:    "0.00",
			FinalValueUSD:      "0.00",
			GainLossUSD:        "0.00",
			GainLossPercentage: "0.00",
			DataPoints: []dto.PortfolioPerformancePoint{
				{Timestamp: uc.now().Format(time.RFC3339Nano), ValueUSD: "0.00"},
			},
		}, nil
	}

	assetBalances := make(map[string]decimal.Decimal)
	for _, wallet := range wallets {
		chain := strings.ToUpper(string(wallet.GetChain()))
		if !entities.IsSupportedSymbol(chain) {
			continue
		}
		balance := wallet.GetBalance()
		if balance.IsZero() {
			continue
		}
		if current, ok := assetBalances[chain]; ok {
			assetBalances[chain] = current.Add(balance)
		} else {
			assetBalances[chain] = balance
		}
	}

	if len(assetBalances) == 0 {
		return dto.PortfolioPerformance{
			Period:             config.label,
			InitialValueUSD:    "0.00",
			FinalValueUSD:      "0.00",
			GainLossUSD:        "0.00",
			GainLossPercentage: "0.00",
			DataPoints: []dto.PortfolioPerformancePoint{
				{Timestamp: uc.now().Format(time.RFC3339Nano), ValueUSD: "0.00"},
			},
		}, nil
	}

	symbols := make([]string, 0, len(assetBalances))
	for symbol := range assetBalances {
		symbols = append(symbols, symbol)
	}

    rates, err := uc.rates.GetRatesBySymbols(ctx, symbols)
    if err != nil {
        ctxLogger.Error("failed to load exchange rates for portfolio performance", slog.String("error", err.Error()))
        return dto.PortfolioPerformance{}, utils.NewAppError(
            "RATE_LOOKUP_FAILED",
            "unable to load exchange rates",
            fiber.StatusInternalServerError,
            err,
			map[string]any{"symbols": symbols},
		)
	}

	rateMap := make(map[string]entities.ExchangeRate, len(rates))
	for _, rate := range rates {
		if rate == nil {
			continue
		}
		rateMap[strings.ToUpper(strings.TrimSpace(rate.GetSymbol()))] = rate
	}

	seriesByAsset := make(map[string][]seriesPoint)
	now := uc.now()
	fromTime := time.Time{}
	if config.duration > 0 {
		fromTime = now.Add(-config.duration)
	}

	for _, symbol := range symbols {
		balance := assetBalances[symbol]
        priceHistory, histErr := uc.loadPriceHistory(ctx, symbol, config.interval, fromTime, now)
        if histErr != nil {
            ctxLogger.Warn("failed to load price history", slog.String("symbol", symbol), slog.String("error", histErr.Error()))
        }

		points := make([]seriesPoint, 0, len(priceHistory)+1)
		for _, entry := range priceHistory {
			value := balance.Mul(entry.price)
			points = append(points, seriesPoint{timestamp: entry.timestamp, value: value})
		}

		rate, ok := rateMap[symbol]
		if ok && rate != nil {
			currentValue := balance.Mul(rate.GetPriceUSD())
			points = append(points, seriesPoint{timestamp: now, value: currentValue})
		}

		if len(points) == 0 {
			// fallback to zero-value point to participate in aggregation
			points = append(points, seriesPoint{timestamp: now, value: decimal.Zero})
		}

		sort.Slice(points, func(i, j int) bool {
			return points[i].timestamp.Before(points[j].timestamp)
		})

		seriesByAsset[symbol] = points
	}

	dataPoints := aggregateSeries(seriesByAsset)
	if len(dataPoints) == 0 {
		dataPoints = append(dataPoints, dto.PortfolioPerformancePoint{Timestamp: now.Format(time.RFC3339Nano), ValueUSD: "0.00"})
	}

	initialValue, _ := decimal.NewFromString(dataPoints[0].ValueUSD)
	finalValue, _ := decimal.NewFromString(dataPoints[len(dataPoints)-1].ValueUSD)
	gainLoss := finalValue.Sub(initialValue)
	gainPercentage := decimal.Zero
	if !initialValue.IsZero() {
		gainPercentage = gainLoss.Div(initialValue).Mul(decimal.NewFromInt(100))
	}

	ctxLogger.Info("portfolio performance calculated",
		slog.String("initial_value_usd", initialValue.StringFixedBank(2)),
		slog.String("gain_loss_usd", gainLoss.StringFixedBank(2)),
	)

	return dto.PortfolioPerformance{
		Period:             config.label,
		InitialValueUSD:    initialValue.StringFixedBank(2),
		FinalValueUSD:      finalValue.StringFixedBank(2),
		GainLossUSD:        gainLoss.StringFixedBank(2),
		GainLossPercentage: gainPercentage.StringFixedBank(2),
		DataPoints:         dataPoints,
	}, nil
}

type pricePoint struct {
	timestamp time.Time
	price     decimal.Decimal
}

type seriesPoint struct {
	timestamp time.Time
	value     decimal.Decimal
}

func (uc *PortfolioPerformanceUseCase) loadPriceHistory(ctx context.Context, symbol string, interval entities.IntervalType, from time.Time, to time.Time) ([]pricePoint, error) {
	if uc.rates == nil {
		return nil, errPerformanceRateRepo
	}

	filter := repositories.PriceHistoryFilter{Symbol: symbol, Interval: interval}
	if !from.IsZero() {
		filter.From = &from
	}
	if !to.IsZero() {
		filter.To = &to
	}

	entries, err := uc.rates.ListPriceHistory(ctx, filter, repositories.ListOptions{Limit: 1000, SortBy: "timestamp", SortOrder: repositories.SortAscending})
	if err != nil {
		return nil, err
	}

	results := make([]pricePoint, 0, len(entries))
	for _, entry := range entries {
		if entry == nil {
			continue
		}
		results = append(results, pricePoint{
			timestamp: entry.GetTimestamp().UTC(),
			price:     entry.GetClose(),
		})
	}

	return results, nil
}

func aggregateSeries(series map[string][]seriesPoint) []dto.PortfolioPerformancePoint {
	if len(series) == 0 {
		return nil
	}

	timestampSet := make(map[int64]time.Time)
	for _, points := range series {
		for _, point := range points {
			ts := point.timestamp.UTC()
			key := ts.UnixNano()
			timestampSet[key] = ts
		}
	}

	keys := make([]int64, 0, len(timestampSet))
	for key := range timestampSet {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	assetIndices := make(map[string]int, len(series))
	assetLastValues := make(map[string]decimal.Decimal, len(series))
	for symbol := range series {
		assetIndices[symbol] = 0
		assetLastValues[symbol] = decimal.Zero
	}

	results := make([]dto.PortfolioPerformancePoint, 0, len(keys))
	for _, key := range keys {
		timestamp := timestampSet[key]
		total := decimal.Zero

		for symbol, points := range series {
			idx := assetIndices[symbol]
			for idx < len(points) && !points[idx].timestamp.After(timestamp) {
				assetLastValues[symbol] = points[idx].value
				idx++
			}
			assetIndices[symbol] = idx
			total = total.Add(assetLastValues[symbol])
		}

		results = append(results, dto.PortfolioPerformancePoint{
			Timestamp: timestamp.Format(time.RFC3339Nano),
			ValueUSD:  total.StringFixedBank(2),
		})
	}

	return results
}

func resolvePeriod(period string) periodConfig {
	switch strings.ToLower(strings.TrimSpace(period)) {
	case "24h":
		return periodConfig{label: "24h", duration: 24 * time.Hour, interval: entities.Interval1h}
	case "7d":
		return periodConfig{label: "7d", duration: 7 * 24 * time.Hour, interval: entities.Interval4h}
	case "30d":
		return periodConfig{label: "30d", duration: 30 * 24 * time.Hour, interval: entities.Interval1d}
	case "90d":
		return periodConfig{label: "90d", duration: 90 * 24 * time.Hour, interval: entities.Interval1d}
	case "1y":
		return periodConfig{label: "1y", duration: 365 * 24 * time.Hour, interval: entities.Interval1w}
	case "all":
		return periodConfig{label: "all", duration: 0, interval: entities.Interval1w}
	default:
		return periodConfig{label: "30d", duration: 30 * 24 * time.Hour, interval: entities.Interval1d}
	}
}
