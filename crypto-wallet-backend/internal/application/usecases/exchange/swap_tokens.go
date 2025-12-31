package exchange

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/services"
)

// SwapTokens handles the complete token swap process from quote to execution.
type SwapTokens struct {
	exchangeService *services.ExchangeService
}

// NewSwapTokens creates a new SwapTokens use case.
func NewSwapTokens(exchangeService *services.ExchangeService) *SwapTokens {
	return &SwapTokens{
		exchangeService: exchangeService,
	}
}

// GetQuote generates an exchange quote for the specified parameters.
func (uc *SwapTokens) GetQuote(ctx context.Context, userID uuid.UUID, req *dto.QuoteRequest) (*dto.QuoteResponse, error) {
	// Validate request
	if req.FromWalletID == uuid.Nil {
		return nil, errors.New("from wallet ID is required")
	}
	if req.ToWalletID == uuid.Nil {
		return nil, errors.New("to wallet ID is required")
	}
	if req.FromAmount == "" {
		return nil, errors.New("from amount is required")
	}

	// Parse from amount
	fromAmount, err := decimal.NewFromString(req.FromAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid from amount: %w", err)
	}

	if fromAmount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("from amount must be positive")
	}

	// Calculate quote using domain service
	operation, err := uc.exchangeService.CalculateQuote(ctx, userID, req.FromWalletID, req.ToWalletID, fromAmount)
	if err != nil {
		if errors.Is(err, services.ErrExchangeSameWallets) {
			return nil, errors.New("cannot exchange between the same wallet")
		}
		if errors.Is(err, services.ErrExchangeInsufficientBalance) {
			return nil, errors.New("insufficient balance in source wallet")
		}
		if errors.Is(err, services.ErrExchangeInvalidTradingPair) {
			return nil, errors.New("trading pair is not available or inactive")
		}
		if errors.Is(err, services.ErrExchangeAmountTooSmall) {
			return nil, errors.New("amount is below minimum swap requirement")
		}
		if errors.Is(err, services.ErrExchangeAmountTooLarge) {
			return nil, errors.New("amount exceeds maximum swap limit")
		}
		if errors.Is(err, services.ErrExchangeNoLiquidity) {
			return nil, errors.New("insufficient liquidity for this trading pair")
		}
		return nil, fmt.Errorf("failed to calculate quote: %w", err)
	}

	// Calculate expiration time in seconds
	expiresIn := int(operation.GetQuoteExpiresAt().Sub(time.Now().UTC()).Seconds())
	if expiresIn < 0 {
		expiresIn = 0
	}

	// Convert to response DTO
	response := &dto.QuoteResponse{
		OperationID:    operation.GetID(),
		FromWalletID:   operation.GetFromWalletID(),
		ToWalletID:     operation.GetToWalletID(),
		FromAmount:     operation.GetFromAmount(),
		ToAmount:       operation.GetToAmount(),
		ExchangeRate:   operation.GetExchangeRate(),
		FeePercentage:  operation.GetFeePercentage(),
		FeeAmount:      operation.GetFeeAmount(),
		QuoteExpiresAt: operation.GetQuoteExpiresAt(),
		ExpiresIn:      expiresIn,
	}

	return response, nil
}

// ExecuteSwap executes a previously quoted exchange operation.
func (uc *SwapTokens) ExecuteSwap(ctx context.Context, req *dto.ExecuteExchangeRequest) (*dto.ExecuteExchangeResponse, error) {
	// Validate request
	if req.OperationID == uuid.Nil {
		return nil, errors.New("operation ID is required")
	}

	// Execute the exchange using domain service
	operation, err := uc.exchangeService.ExecuteExchange(ctx, req.OperationID)
	if err != nil {
		if errors.Is(err, services.ErrExchangeQuoteExpired) {
			return nil, errors.New("quote has expired, please get a new quote")
		}
		if errors.Is(err, services.ErrExchangeInvalidStatus) {
			return nil, errors.New("exchange operation is not in a valid state for execution")
		}
		return nil, fmt.Errorf("failed to execute exchange: %w", err)
	}

	// Convert to response DTO
	response := &dto.ExecuteExchangeResponse{
		OperationID:       operation.GetID(),
		Status:            string(operation.GetStatus()),
		FromWalletID:      operation.GetFromWalletID(),
		ToWalletID:        operation.GetToWalletID(),
		FromAmount:        operation.GetFromAmount(),
		ToAmount:          operation.GetToAmount(),
		ExchangeRate:      operation.GetExchangeRate(),
		FeeAmount:         operation.GetFeeAmount(),
		ExecutedAt:        operation.GetExecutedAt(),
		FromTransactionID: operation.GetFromTransactionID(),
		ToTransactionID:   operation.GetToTransactionID(),
		ErrorMessage:      operation.GetErrorMessage(),
	}

	return response, nil
}

// CancelSwap cancels a pending exchange operation.
func (uc *SwapTokens) CancelSwap(ctx context.Context, req *dto.CancelExchangeRequest) (*dto.CancelExchangeResponse, error) {
	// Validate request
	if req.OperationID == uuid.Nil {
		return nil, errors.New("operation ID is required")
	}

	// Cancel the exchange using domain service
	err := uc.exchangeService.CancelExchange(ctx, req.OperationID, req.Reason)
	if err != nil {
		if errors.Is(err, services.ErrExchangeInvalidStatus) {
			return nil, errors.New("exchange operation is not in a valid state for cancellation")
		}
		return nil, fmt.Errorf("failed to cancel exchange: %w", err)
	}

	// Convert to response DTO
	response := &dto.CancelExchangeResponse{
		OperationID: req.OperationID,
		Status:      string(entities.ExchangeStatusCancelled),
		CancelledAt: time.Now().UTC(),
		Reason:      req.Reason,
	}

	return response, nil
}
