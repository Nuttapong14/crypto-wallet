package transaction

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	domainservices "github.com/crypto-wallet/backend/internal/domain/services"
	"github.com/crypto-wallet/backend/internal/infrastructure/audit"
	"github.com/crypto-wallet/backend/internal/infrastructure/blockchain"
	appLogging "github.com/crypto-wallet/backend/internal/infrastructure/logging"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// SendTransactionInput encapsulates all necessary data to initiate a transaction.
type SendTransactionInput struct {
	UserID  string
	Payload dto.SendTransactionRequest
}

// SendTransactionUseCase coordinates the send flow between adapters and persistence.
type SendTransactionUseCase struct {
	service      Service
	transactions TransactionRepo
	wallets      WalletRepo
	ledgerWriter LedgerWriter
	resolver     BlockchainResolver
	auditLogger  AuditLogger
	logger       *slog.Logger
	retryCfg     blockchain.RetryConfig
}

// NewSendTransactionUseCase constructs the use case.
func NewSendTransactionUseCase(
	service Service,
	transactions TransactionRepo,
	wallets WalletRepo,
	ledger LedgerWriter,
	resolver BlockchainResolver,
	auditLogger AuditLogger,
	logger *slog.Logger,
) *SendTransactionUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &SendTransactionUseCase{
		service:      service,
		transactions: transactions,
		wallets:      wallets,
		ledgerWriter: ledger,
		resolver:     resolver,
		auditLogger:  auditLogger,
		logger:       logger,
		retryCfg:     blockchain.RetryConfig{Attempts: 3, Delay: 350 * time.Millisecond},
	}
}

// Execute performs the send transaction workflow end-to-end.
func (uc *SendTransactionUseCase) Execute(ctx context.Context, input SendTransactionInput) (dto.TransactionStatusResponse, error) {
	logger := appLogging.LoggerFromContext(ctx, uc.logger)
	validation := input.Payload.Validate()

	userID, err := uuid.Parse(strings.TrimSpace(input.UserID))
	if err != nil {
		validation.Add("userId", "must be a valid UUID")
	}

	walletID, err := uuid.Parse(strings.TrimSpace(input.Payload.WalletID))
	if err != nil {
		validation.Add("walletId", "must be a valid UUID")
	}

	if err == nil {
		logger = logger.With(
			slog.String("user_id", userID.String()),
			slog.String("wallet_id", walletID.String()),
			slog.String("chain", strings.ToUpper(input.Payload.Chain)),
		)
	}

	logger.Debug("send transaction request received")

	chain := entities.NormalizeChain(input.Payload.Chain)
	if chain == "" {
		validation.Add("chain", "must be one of BTC, ETH, SOL, XLM")
	}

	amount := decimal.Zero
	fee := decimal.Zero
	if validation.IsEmpty() {
		amount = parseDecimal(input.Payload.Amount, "amount", &validation)
		if strings.TrimSpace(input.Payload.Fee) != "" {
			fee = parseDecimal(input.Payload.Fee, "fee", &validation)
		}
		if fee.IsZero() {
			fee = decimal.Zero
		}
	}

	if !validation.IsEmpty() {
		return dto.TransactionStatusResponse{}, wrapValidationError(validation)
	}

	wallet, err := uc.wallets.GetByID(ctx, walletID)
	if err != nil {
		logger.Error("failed to load wallet", slog.String("error", err.Error()))
		return dto.TransactionStatusResponse{}, err
	}

	if wallet.GetStatus() != entities.WalletStatusActive {
		return dto.TransactionStatusResponse{}, utils.NewAppError(
			"WALLET_INACTIVE",
			"wallet must be active to send transactions",
			fiber.StatusForbidden,
			nil,
			nil,
		)
	}

	if wallet.GetChain() != chain {
		return dto.TransactionStatusResponse{}, utils.NewAppError(
			"CHAIN_MISMATCH",
			"wallet chain mismatch",
			fiber.StatusBadRequest,
			nil,
			map[string]any{"expected": wallet.GetChain(), "received": chain},
		)
	}

	adapter, err := uc.resolver.Resolve(chain)
	if err != nil {
		logger.Error("blockchain adapter resolve failed", slog.String("error", err.Error()))
		return dto.TransactionStatusResponse{}, utils.NewAppError(
			"ADAPTER_NOT_FOUND",
			"blockchain adapter not configured",
			fiber.StatusBadGateway,
			err,
			nil,
		)
	}

	txnRequest := &blockchain.TransactionRequest{
		FromAddress: wallet.GetAddress(),
		ToAddress:   input.Payload.ToAddress,
		Amount:      amount.String(),
		Fee:         fee.String(),
		Memo:        input.Payload.Memo,
		Metadata:    input.Payload.Metadata,
	}

	logger.Debug("creating unsigned transaction")
	unsigned, err := blockchain.Retry(ctx, logger, uc.retryCfg, "create_transaction", func(inner context.Context) (*blockchain.UnsignedTransaction, error) {
		return adapter.CreateTransaction(inner, txnRequest)
	})
	if err != nil {
		logger.Error("create transaction failed", slog.String("error", err.Error()))
		return dto.TransactionStatusResponse{}, err
	}

	signed, err := adapter.SignTransaction(ctx, unsigned, wallet.GetEncryptedPrivateKey())
	if err != nil {
		logger.Error("sign transaction failed", slog.String("error", err.Error()))
		return dto.TransactionStatusResponse{}, err
	}

	broadcastHash, err := blockchain.Retry(ctx, logger, uc.retryCfg, "broadcast_transaction", func(inner context.Context) (string, error) {
		return adapter.BroadcastTransaction(inner, signed)
	})
	if err != nil {
		logger.Error("broadcast transaction failed", slog.String("error", err.Error()))
		return dto.TransactionStatusResponse{}, err
	}
	logger.Info("transaction broadcast", slog.String("tx_hash", broadcastHash))

	domainResult, err := uc.service.PrepareSend(domainservices.SendParams{
		WalletID:    wallet.GetID(),
		Chain:       chain,
		FromAddress: wallet.GetAddress(),
		ToAddress:   input.Payload.ToAddress,
		Amount:      amount,
		Fee:         fee,
		Metadata:    mergeMetadata(unsigned.Metadata, signed.Metadata, input.Payload.Metadata),
	})
	if err != nil {
		return dto.TransactionStatusResponse{}, err
	}

	transaction := domainResult.Transaction
	if setErr := transaction.SetHash(broadcastHash); setErr != nil {
		return dto.TransactionStatusResponse{}, setErr
	}
	if statusErr := transaction.SetStatus(entities.TransactionStatusConfirming); statusErr != nil {
		return dto.TransactionStatusResponse{}, statusErr
	}
	transaction.Touch(time.Now().UTC())

	if err := uc.transactions.Create(ctx, transaction); err != nil {
		logger.Error("persist transaction failed", slog.String("error", err.Error()))
		return dto.TransactionStatusResponse{}, err
	}

	if uc.ledgerWriter != nil {
		entries := []*entities.LedgerEntryEntity{}
		if domainResult.LedgerDebit != nil {
			entries = append(entries, domainResult.LedgerDebit)
		}
		if domainResult.LedgerCredit != nil {
			entries = append(entries, domainResult.LedgerCredit)
		}
		if len(entries) > 0 {
			if err := uc.ledgerWriter.CreateEntries(ctx, entries...); err != nil {
				uc.logger.Warn("failed to persist ledger entries", slog.String("error", err.Error()))
			}
		}
	}

	if uc.auditLogger != nil {
		_ = uc.auditLogger.Record(ctx, audit.Entry{
			ActorID:  userID,
			Action:   "transaction_send",
			TargetID: transaction.GetID().String(),
			Metadata: map[string]any{
				"wallet_id":    wallet.GetID().String(),
				"chain":        chain,
				"hash":         transaction.GetHash(),
				"amount":       transaction.GetAmount().String(),
				"to_address":   transaction.GetToAddress(),
				"from_address": transaction.GetFromAddress(),
			},
		})
	}

	return mapTransaction(transaction), nil
}

func mergeMetadata(values ...map[string]any) map[string]any {
	merged := map[string]any{}
	for _, value := range values {
		if value == nil {
			continue
		}
		for k, v := range value {
			merged[k] = v
		}
	}
	return merged
}
