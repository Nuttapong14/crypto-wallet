package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
)

var (
	ErrExchangeSameWallets         = errors.New("exchange service: cannot exchange between the same wallet")
	ErrExchangeInsufficientBalance = errors.New("exchange service: insufficient balance in source wallet")
	ErrExchangeInvalidTradingPair  = errors.New("exchange service: invalid or inactive trading pair")
	ErrExchangeAmountTooSmall      = errors.New("exchange service: amount is below minimum swap requirement")
	ErrExchangeAmountTooLarge      = errors.New("exchange service: amount exceeds maximum swap limit")
	ErrExchangeNoLiquidity         = errors.New("exchange service: insufficient liquidity for this trading pair")
	ErrExchangeQuoteExpired        = errors.New("exchange service: quote has expired")
	ErrExchangeInvalidStatus       = errors.New("exchange service: invalid exchange operation status")
)

// ExchangeService provides domain-level business logic for cryptocurrency exchanges.
type ExchangeService struct {
	exchangeRepo    repositories.ExchangeOperationRepository
	tradingPairRepo repositories.TradingPairRepository
	walletRepo      repositories.WalletRepository
}

// NewExchangeService creates a new ExchangeService instance.
func NewExchangeService(
	exchangeRepo repositories.ExchangeOperationRepository,
	tradingPairRepo repositories.TradingPairRepository,
	walletRepo repositories.WalletRepository,
) *ExchangeService {
	return &ExchangeService{
		exchangeRepo:    exchangeRepo,
		tradingPairRepo: tradingPairRepo,
		walletRepo:      walletRepo,
	}
}

// GetExchangeRate retrieves the current exchange rate for a trading pair.
func (s *ExchangeService) GetExchangeRate(ctx context.Context, baseSymbol, quoteSymbol string) (entities.TradingPair, error) {
	pair, err := s.tradingPairRepo.GetBySymbols(ctx, baseSymbol, quoteSymbol)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrExchangeInvalidTradingPair
		}
		return nil, fmt.Errorf("exchange service: get trading pair: %w", err)
	}

	if !pair.IsActive() || !pair.HasLiquidity() {
		return nil, ErrExchangeInvalidTradingPair
	}

	return pair, nil
}

// CalculateQuote calculates a quote for exchanging a specific amount.
func (s *ExchangeService) CalculateQuote(
	ctx context.Context,
	userID uuid.UUID,
	fromWalletID, toWalletID uuid.UUID,
	fromAmount decimal.Decimal,
) (*entities.ExchangeOperationEntity, error) {
	// Validate wallets are different
	if fromWalletID == toWalletID {
		return nil, ErrExchangeSameWallets
	}

	// Get wallets to determine symbols and check balance
	fromWallet, err := s.walletRepo.GetByID(ctx, fromWalletID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, fmt.Errorf("exchange service: source wallet not found")
		}
		return nil, fmt.Errorf("exchange service: get source wallet: %w", err)
	}

	toWallet, err := s.walletRepo.GetByID(ctx, toWalletID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, fmt.Errorf("exchange service: destination wallet not found")
		}
		return nil, fmt.Errorf("exchange service: get destination wallet: %w", err)
	}

	// Check if user owns both wallets
	if fromWallet.GetUserID() != userID || toWallet.GetUserID() != userID {
		return nil, fmt.Errorf("exchange service: wallet ownership mismatch")
	}

	// Check sufficient balance
	if fromWallet.GetBalance().LessThan(fromAmount) {
		return nil, ErrExchangeInsufficientBalance
	}

	// Get trading pair (determine base/quote from wallet chains)
	// For simplicity, we'll use chain as symbol - in real implementation, this would be more sophisticated
	baseSymbol := string(fromWallet.GetChain())
	quoteSymbol := string(toWallet.GetChain())

	pair, err := s.GetExchangeRate(ctx, baseSymbol, quoteSymbol)
	if err != nil {
		return nil, err
	}

	// Validate amount constraints
	if fromAmount.LessThan(pair.GetMinSwapAmount()) {
		return nil, ErrExchangeAmountTooSmall
	}
	if max := pair.GetMaxSwapAmount(); max != nil && fromAmount.GreaterThan(*max) {
		return nil, ErrExchangeAmountTooLarge
	}

	// Calculate exchange amounts
	feeAmount := pair.GetFeePercentage().Div(decimal.NewFromInt(100)).Mul(fromAmount)
	netAmount := fromAmount.Sub(feeAmount)
	toAmount := netAmount.Mul(pair.GetExchangeRate())

	// Create exchange operation with quote
	now := time.Now().UTC()
	quoteExpiresAt := now.Add(60 * time.Second) // 60 second quote expiration

	operation, err := entities.NewExchangeOperationEntity(entities.ExchangeOperationParams{
		UserID:         userID,
		FromWalletID:   fromWalletID,
		ToWalletID:     toWalletID,
		FromAmount:     fromAmount,
		ToAmount:       toAmount,
		ExchangeRate:   pair.GetExchangeRate(),
		FeePercentage:  pair.GetFeePercentage(),
		FeeAmount:      feeAmount,
		Status:         entities.ExchangeStatusPending,
		QuoteExpiresAt: quoteExpiresAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	if err != nil {
		return nil, fmt.Errorf("exchange service: create exchange operation: %w", err)
	}

	return operation, nil
}

// ExecuteExchange executes a pending exchange operation.
func (s *ExchangeService) ExecuteExchange(
	ctx context.Context,
	operationID uuid.UUID,
) (*entities.ExchangeOperationEntity, error) {
	// Get the exchange operation
	operation, err := s.exchangeRepo.GetByID(ctx, operationID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, fmt.Errorf("exchange service: exchange operation not found")
		}
		return nil, fmt.Errorf("exchange service: get exchange operation: %w", err)
	}

	// Validate status
	if operation.GetStatus() != entities.ExchangeStatusPending {
		return nil, ErrExchangeInvalidStatus
	}

	// Check quote expiration
	if operation.(*entities.ExchangeOperationEntity).IsQuoteExpired() {
		return nil, ErrExchangeQuoteExpired
	}

	// Mark as processing
	if err := operation.(*entities.ExchangeOperationEntity).MarkProcessing(); err != nil {
		return nil, fmt.Errorf("exchange service: mark processing: %w", err)
	}

	// Update in repository
	if err := s.exchangeRepo.Update(ctx, operation); err != nil {
		return nil, fmt.Errorf("exchange service: update processing status: %w", err)
	}

	// Get wallets to perform the exchange
	fromWallet, err := s.walletRepo.GetByID(ctx, operation.GetFromWalletID())
	if err != nil {
		return s.markExchangeFailed(ctx, operation, fmt.Sprintf("failed to get source wallet: %v", err))
	}

	toWallet, err := s.walletRepo.GetByID(ctx, operation.GetToWalletID())
	if err != nil {
		return s.markExchangeFailed(ctx, operation, fmt.Sprintf("failed to get destination wallet: %v", err))
	}

	// Check final balance (in case it changed since quote)
	if fromWallet.GetBalance().LessThan(operation.GetFromAmount()) {
		return s.markExchangeFailed(ctx, operation, "insufficient balance at execution time")
	}

	// Perform the exchange (this would typically involve transactions)
	// For now, we'll simulate the exchange
	now := time.Now().UTC()

	// Update from wallet (subtract amount)
	fromWalletEntity := fromWallet.(*entities.WalletEntity)
	if err := fromWalletEntity.UpdateBalance(fromWallet.GetBalance().Sub(operation.GetFromAmount()), now); err != nil {
		return s.markExchangeFailed(ctx, operation, fmt.Sprintf("failed to update source wallet balance: %v", err))
	}
	fromWalletEntity.Touch(now)

	if err := s.walletRepo.Update(ctx, fromWallet); err != nil {
		return s.markExchangeFailed(ctx, operation, fmt.Sprintf("failed to update source wallet: %v", err))
	}

	// Update to wallet (add amount)
	toWalletEntity := toWallet.(*entities.WalletEntity)
	if err := toWalletEntity.UpdateBalance(toWallet.GetBalance().Add(operation.GetToAmount()), now); err != nil {
		return s.markExchangeFailed(ctx, operation, fmt.Sprintf("failed to update destination wallet balance: %v", err))
	}
	toWalletEntity.Touch(now)

	if err := s.walletRepo.Update(ctx, toWallet); err != nil {
		return s.markExchangeFailed(ctx, operation, fmt.Sprintf("failed to update destination wallet: %v", err))
	}

	// Mark exchange as completed
	if err := operation.(*entities.ExchangeOperationEntity).MarkCompleted(now); err != nil {
		return nil, fmt.Errorf("exchange service: mark completed: %w", err)
	}

	// Update trading pair volume
	pair, err := s.tradingPairRepo.GetBySymbols(ctx,
		string(fromWallet.GetChain()),
		string(toWallet.GetChain()))
	if err == nil {
		pairEntity := pair.(*entities.TradingPairEntity)
		if err := pairEntity.AddVolume(operation.GetFromAmount()); err == nil {
			s.tradingPairRepo.Update(ctx, pair)
		}
	}

	// Final update
	if err := s.exchangeRepo.Update(ctx, operation); err != nil {
		return nil, fmt.Errorf("exchange service: update completed status: %w", err)
	}

	return operation.(*entities.ExchangeOperationEntity), nil
}

// CancelExchange cancels a pending exchange operation.
func (s *ExchangeService) CancelExchange(
	ctx context.Context,
	operationID uuid.UUID,
	reason string,
) error {
	operation, err := s.exchangeRepo.GetByID(ctx, operationID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return fmt.Errorf("exchange service: exchange operation not found")
		}
		return fmt.Errorf("exchange service: get exchange operation: %w", err)
	}

	// Validate status
	if operation.GetStatus() != entities.ExchangeStatusPending {
		return ErrExchangeInvalidStatus
	}

	// Mark as cancelled
	if err := operation.(*entities.ExchangeOperationEntity).MarkCancelled(); err != nil {
		return fmt.Errorf("exchange service: mark cancelled: %w", err)
	}

	// Set error message if provided
	if reason != "" {
		operation.(*entities.ExchangeOperationEntity).SetErrorMessage(reason)
	}

	// Update in repository
	if err := s.exchangeRepo.Update(ctx, operation); err != nil {
		return fmt.Errorf("exchange service: update cancelled status: %w", err)
	}

	return nil
}

// GetUserExchangeHistory retrieves exchange history for a user.
func (s *ExchangeService) GetUserExchangeHistory(
	ctx context.Context,
	userID uuid.UUID,
	filter repositories.ExchangeOperationFilter,
	opts repositories.ListOptions,
) ([]entities.ExchangeOperation, error) {
	return s.exchangeRepo.GetByUser(ctx, userID, filter, opts)
}

// GetUserExchangeHistoryCount retrieves the total count of exchange operations for a user with filters.
func (s *ExchangeService) GetUserExchangeHistoryCount(
	ctx context.Context,
	userID uuid.UUID,
	filter repositories.ExchangeOperationFilter,
) (int64, error) {
	return s.exchangeRepo.GetCountByUser(ctx, userID, filter)
}

// GetActiveTradingPairs retrieves all active trading pairs.
func (s *ExchangeService) GetActiveTradingPairs(ctx context.Context) ([]entities.TradingPair, error) {
	return s.tradingPairRepo.GetActivePairs(ctx)
}

// ExpirePendingQuotes expires all pending quotes that have passed their expiration time.
func (s *ExchangeService) ExpirePendingQuotes(ctx context.Context) ([]entities.ExchangeOperation, error) {
	expiredOperations, err := s.exchangeRepo.GetExpiredPending(ctx)
	if err != nil {
		return nil, fmt.Errorf("exchange service: get expired pending: %w", err)
	}

	now := time.Now().UTC()
	for _, operation := range expiredOperations {
		operationEntity := operation.(*entities.ExchangeOperationEntity)
		if err := operationEntity.MarkCancelled(); err != nil {
			continue // Skip if we can't mark as cancelled
		}
		operationEntity.SetErrorMessage("Quote expired")
		operationEntity.Touch(now)

		// Update in repository (ignore errors for individual operations)
		s.exchangeRepo.Update(ctx, operation)
	}

	return expiredOperations, nil
}

// GetExchangeStats retrieves exchange statistics for a user.
func (s *ExchangeService) GetExchangeStats(ctx context.Context, userID uuid.UUID) (*dto.ExchangeStatsResponse, error) {
	// Get total operations count
	totalOps, err := s.exchangeRepo.GetCountByUser(ctx, userID, repositories.ExchangeOperationFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to get total operations: %w", err)
	}

	// Get total volume
	totalVolume, err := s.exchangeRepo.GetVolumeByUser(ctx, userID, repositories.ExchangeOperationFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to get total volume: %w", err)
	}

	// Get completed count
	completedStatus := entities.ExchangeStatusCompleted
	completedCount, err := s.exchangeRepo.GetCountByUser(ctx, userID, repositories.ExchangeOperationFilter{
		Status: &completedStatus,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get completed count: %w", err)
	}

	// Get failed count
	failedStatus := entities.ExchangeStatusFailed
	failedCount, err := s.exchangeRepo.GetCountByUser(ctx, userID, repositories.ExchangeOperationFilter{
		Status: &failedStatus,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get failed count: %w", err)
	}

	// Get pending count
	pendingStatus := entities.ExchangeStatusPending
	pendingCount, err := s.exchangeRepo.GetCountByUser(ctx, userID, repositories.ExchangeOperationFilter{
		Status: &pendingStatus,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pending count: %w", err)
	}

	response := &dto.ExchangeStatsResponse{
		TotalOperations: totalOps,
		TotalVolume:     totalVolume,
		CompletedCount:  completedCount,
		FailedCount:     failedCount,
		PendingCount:    pendingCount,
	}

	return response, nil
}

// Helper method to mark exchange as failed
func (s *ExchangeService) markExchangeFailed(
	ctx context.Context,
	operation entities.ExchangeOperation,
	reason string,
) (*entities.ExchangeOperationEntity, error) {
	operationEntity := operation.(*entities.ExchangeOperationEntity)
	if err := operationEntity.MarkFailed(reason); err != nil {
		return nil, fmt.Errorf("exchange service: mark failed: %w", err)
	}

	if err := s.exchangeRepo.Update(ctx, operation); err != nil {
		return nil, fmt.Errorf("exchange service: update failed status: %w", err)
	}

	return operationEntity, nil
}
