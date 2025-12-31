package transaction

import (
    "context"
    "errors"
    "log/slog"
    "strings"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"

    "github.com/crypto-wallet/backend/internal/application/dto"
    "github.com/crypto-wallet/backend/internal/domain/entities"
    "github.com/crypto-wallet/backend/pkg/utils"
)

// GetTransactionStatusInput captures the lookup parameters.
type GetTransactionStatusInput struct {
    TransactionID string
    Chain         string
    Hash          string
}

// GetTransactionStatusUseCase resolves transaction status queries.
type GetTransactionStatusUseCase struct {
    transactions TransactionRepo
    logger       *slog.Logger
}

// NewGetTransactionStatusUseCase constructs the use case.
func NewGetTransactionStatusUseCase(repo TransactionRepo, logger *slog.Logger) *GetTransactionStatusUseCase {
    if logger == nil {
        logger = slog.Default()
    }
    return &GetTransactionStatusUseCase{transactions: repo, logger: logger}
}

// Execute retrieves the transaction status using the provided criteria.
func (uc *GetTransactionStatusUseCase) Execute(ctx context.Context, input GetTransactionStatusInput) (dto.TransactionStatusResponse, error) {
    if uc.transactions == nil {
        return dto.TransactionStatusResponse{}, errors.New("transaction repository not configured")
    }

    if id := strings.TrimSpace(input.TransactionID); id != "" {
        txID, err := uuid.Parse(id)
        if err != nil {
            return dto.TransactionStatusResponse{}, utils.NewAppError(
                "VALIDATION_ERROR",
                "transaction id must be a UUID",
                fiber.StatusBadRequest,
                nil,
                nil,
            )
        }
        tx, err := uc.transactions.GetByID(ctx, txID)
        if err != nil {
            return dto.TransactionStatusResponse{}, err
        }
        return mapTransaction(tx), nil
    }

    if strings.TrimSpace(input.Hash) == "" || strings.TrimSpace(input.Chain) == "" {
        return dto.TransactionStatusResponse{}, utils.NewAppError(
            "VALIDATION_ERROR",
            "transaction id or (chain + hash) is required",
            fiber.StatusBadRequest,
            nil,
            nil,
        )
    }

    chain := entities.NormalizeChain(input.Chain)
    if chain == "" {
        return dto.TransactionStatusResponse{}, utils.NewAppError(
            "VALIDATION_ERROR",
            "invalid chain supplied",
            fiber.StatusBadRequest,
            nil,
            nil,
        )
    }

    tx, err := uc.transactions.GetByHash(ctx, chain, strings.TrimSpace(input.Hash))
    if err != nil {
        return dto.TransactionStatusResponse{}, err
    }
    return mapTransaction(tx), nil
}
