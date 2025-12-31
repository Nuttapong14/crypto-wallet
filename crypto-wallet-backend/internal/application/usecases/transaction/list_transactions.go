package transaction

import (
    "context"
    "log/slog"
    "strings"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"

    "github.com/crypto-wallet/backend/internal/application/dto"
    "github.com/crypto-wallet/backend/internal/domain/repositories"
    "github.com/crypto-wallet/backend/pkg/utils"
)

// ListTransactionsInput captures filter parameters for listing.
type ListTransactionsInput struct {
    WalletID  string
    Status    string
    Chain     string
    Limit     int
    Offset    int
    SortBy    string
    SortOrder string
}

// ListTransactionsUseCase handles paginated transaction retrieval.
type ListTransactionsUseCase struct {
    transactions TransactionRepo
    logger       *slog.Logger
}

// NewListTransactionsUseCase constructs the use case.
func NewListTransactionsUseCase(repo TransactionRepo, logger *slog.Logger) *ListTransactionsUseCase {
    if logger == nil {
        logger = slog.Default()
    }
    return &ListTransactionsUseCase{transactions: repo, logger: logger}
}

// Execute returns a paginated list of transactions for the given wallet.
func (uc *ListTransactionsUseCase) Execute(ctx context.Context, input ListTransactionsInput) (dto.TransactionListResponse, error) {
    walletID, err := uuid.Parse(strings.TrimSpace(input.WalletID))
    if err != nil {
        return dto.TransactionListResponse{}, utils.NewAppError(
            "VALIDATION_ERROR",
            "walletId must be a valid UUID",
            fiber.StatusBadRequest,
            nil,
            nil,
        )
    }

    opts := repositories.ListOptions{
        Limit:     input.Limit,
        Offset:    input.Offset,
        SortBy:    input.SortBy,
        SortOrder: repositories.SortOrder(strings.ToUpper(strings.TrimSpace(input.SortOrder))),
    }

    transactions, err := uc.transactions.ListByWallet(ctx, walletID, opts)
    if err != nil {
        return dto.TransactionListResponse{}, err
    }

    response := dto.TransactionListResponse{
        Items:  mapTransactions(transactions),
        Limit:  opts.WithDefaults().Limit,
        Offset: opts.Offset,
    }

    response.Total = int64(len(response.Items))

    return response, nil
}
