package analytics

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"strings"

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
	errWalletRepositoryRequired = errors.New("portfolio summary: wallet repository not configured")
	errRateRepositoryRequired   = errors.New("portfolio summary: rate repository not configured")
)

// PortfolioSummaryUseCase calculates a user's portfolio allocation and totals.
type PortfolioSummaryUseCase struct {
	wallets repositories.WalletRepository
	rates   repositories.RateRepository
	logger  *slog.Logger
}

// NewPortfolioSummaryUseCase constructs the use case.
func NewPortfolioSummaryUseCase(wallets repositories.WalletRepository, rates repositories.RateRepository, logger *slog.Logger) *PortfolioSummaryUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &PortfolioSummaryUseCase{
		wallets: wallets,
		rates:   rates,
		logger:  logger,
	}
}

// Execute returns the aggregated portfolio summary for the supplied user.
func (uc *PortfolioSummaryUseCase) Execute(ctx context.Context, userID uuid.UUID) (dto.PortfolioSummary, error) {
	if uc.wallets == nil {
		return dto.PortfolioSummary{}, errWalletRepositoryRequired
	}
	if uc.rates == nil {
		return dto.PortfolioSummary{}, errRateRepositoryRequired
	}
	if userID == uuid.Nil {
		return dto.PortfolioSummary{}, utils.NewAppError(
			"VALIDATION_ERROR",
			"user id is required",
			fiber.StatusBadRequest,
			nil,
			nil,
		)
	}

	ctxLogger := appLogging.LoggerFromContext(ctx, uc.logger).With(slog.String("user_id", userID.String()))
	ctxLogger.Debug("compiling portfolio summary")

	wallets, err := uc.wallets.ListByUser(ctx, userID, repositories.WalletFilter{}, repositories.ListOptions{Limit: 1000, SortBy: "created_at", SortOrder: repositories.SortDescending})
	if err != nil {
		ctxLogger.Error("failed to list wallets for portfolio summary", slog.String("error", err.Error()))
        return dto.PortfolioSummary{}, utils.NewAppError(
            "DATABASE_ERROR",
            "unable to load wallets",
            fiber.StatusInternalServerError,
            err,
			map[string]any{"userId": userID.String()},
		)
	}

	if len(wallets) == 0 {
		return dto.PortfolioSummary{
			TotalBalanceUSD:          "0.00",
			TotalChange24h:           "0.00",
			TotalChangePercentage24h: "0.00",
			Assets:                   []dto.PortfolioAsset{},
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
		return dto.PortfolioSummary{
			TotalBalanceUSD:          "0.00",
			TotalChange24h:           "0.00",
			TotalChangePercentage24h: "0.00",
			Assets:                   []dto.PortfolioAsset{},
		}, nil
	}

	symbols := make([]string, 0, len(assetBalances))
	for symbol := range assetBalances {
		symbols = append(symbols, symbol)
	}

	rates, err := uc.rates.GetRatesBySymbols(ctx, symbols)
	if err != nil {
		ctxLogger.Error("failed to load exchange rates for portfolio summary", slog.String("error", err.Error()))
        return dto.PortfolioSummary{}, utils.NewAppError(
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

	totalBalanceUSD := decimal.Zero
	totalChangeUSD := decimal.Zero
	previousTotalUSD := decimal.Zero
	assets := make([]dto.PortfolioAsset, 0, len(assetBalances))

	for _, symbol := range symbols {
		balance := assetBalances[symbol]
		rate, ok := rateMap[symbol]
		price := decimal.Zero
		change24h := decimal.Zero
		if ok && rate != nil {
			price = rate.GetPriceUSD()
			change24h = rate.GetPriceChange24h()
		}

		valueUSD := balance.Mul(price)
		changeUSD := balance.Mul(change24h)

		totalBalanceUSD = totalBalanceUSD.Add(valueUSD)
		totalChangeUSD = totalChangeUSD.Add(changeUSD)

		previousPrice := price.Sub(change24h)
		if previousPrice.IsNegative() {
			previousPrice = decimal.Zero
		}
		previousTotalUSD = previousTotalUSD.Add(balance.Mul(previousPrice))

		assets = append(assets, dto.PortfolioAsset{
			Symbol:     symbol,
			Balance:    balance.StringFixedBank(8),
			BalanceUSD: valueUSD.StringFixedBank(2),
			Percentage: "0.00",
		})
	}

	if totalBalanceUSD.IsZero() {
		// If balances cancel out, ensure totals are consistent
		return dto.PortfolioSummary{
			TotalBalanceUSD:          "0.00",
			TotalChange24h:           totalChangeUSD.StringFixedBank(2),
			TotalChangePercentage24h: "0.00",
			Assets:                   assets,
		}, nil
	}

	// Sort assets by USD balance desc
	sort.Slice(assets, func(i, j int) bool {
		left, _ := decimal.NewFromString(assets[i].BalanceUSD)
		right, _ := decimal.NewFromString(assets[j].BalanceUSD)
		return left.GreaterThan(right)
	})

	for idx, asset := range assets {
		value, err := decimal.NewFromString(asset.BalanceUSD)
		if err != nil {
			continue
		}
		percentage := decimal.Zero
		if !totalBalanceUSD.IsZero() {
			percentage = value.Div(totalBalanceUSD).Mul(decimal.NewFromInt(100))
		}
		assets[idx].Percentage = percentage.StringFixedBank(2)
	}

	changePercentage := decimal.Zero
    if !previousTotalUSD.IsZero() {
        changePercentage = totalChangeUSD.Div(previousTotalUSD).Mul(decimal.NewFromInt(100))
    }

	ctxLogger.Info("portfolio summary calculated",
		slog.String("total_balance_usd", totalBalanceUSD.StringFixedBank(2)),
		slog.Int("asset_count", len(assets)),
	)

    return dto.PortfolioSummary{
        TotalBalanceUSD:          totalBalanceUSD.StringFixedBank(2),
        TotalChange24h:           totalChangeUSD.StringFixedBank(2),
        TotalChangePercentage24h: changePercentage.StringFixedBank(2),
        Assets:                   assets,
	}, nil
}
