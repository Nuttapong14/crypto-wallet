package transaction

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// ExportTransactionsInput captures filter parameters for transaction export.
type ExportTransactionsInput struct {
	WalletID  *uuid.UUID
	Chain     *entities.Chain
	Type      *entities.TransactionType
	Status    *entities.TransactionStatus
	StartDate *time.Time
	EndDate   *time.Time
	Format    string // csv, json
}

// ExportTransactionsUseCase handles transaction export functionality.
type ExportTransactionsUseCase struct {
	transactions TransactionRepo
	logger       *slog.Logger
}

// NewExportTransactionsUseCase constructs the use case.
func NewExportTransactionsUseCase(repo TransactionRepo, logger *slog.Logger) *ExportTransactionsUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &ExportTransactionsUseCase{transactions: repo, logger: logger}
}

// Execute generates an export of transactions based on filters.
func (uc *ExportTransactionsUseCase) Execute(ctx context.Context, input ExportTransactionsInput) (dto.ExportResponse, error) {
	// Build filter options
	filter := repositories.TransactionFilter{
		WalletID:  input.WalletID,
		Chain:     input.Chain,
		Type:      input.Type,
		Status:    input.Status,
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
	}

	// Get all transactions matching filters (no pagination for export)
	opts := repositories.ListOptions{
		Limit:     10000, // Large limit for export
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: repositories.SortDescending,
	}

	transactions, _, err := uc.transactions.ListWithFilters(ctx, filter, opts)
	if err != nil {
		uc.logger.Error("failed to list transactions for export", "error", err)
		return dto.ExportResponse{}, utils.NewAppError(
			"DATABASE_ERROR",
			"Failed to retrieve transactions for export",
			fiber.StatusInternalServerError,
			nil,
			map[string]any{"error": err.Error()},
		)
	}

	// Generate export based on format
	var filename string
	var content []byte

	switch strings.ToLower(input.Format) {
	case "csv":
		filename, content, err = uc.generateCSV(transactions)
	case "json":
		filename, content, err = uc.generateJSON(transactions)
	default:
		return dto.ExportResponse{}, utils.NewAppError(
			"VALIDATION_ERROR",
			"Unsupported export format",
			fiber.StatusBadRequest,
			nil,
			map[string]any{"format": input.Format},
		)
	}

	if err != nil {
		uc.logger.Error("failed to generate export", "format", input.Format, "error", err)
		return dto.ExportResponse{}, utils.NewAppError(
			"EXPORT_ERROR",
			"Failed to generate export file",
			fiber.StatusInternalServerError,
			nil,
			map[string]any{"error": err.Error()},
		)
	}

	// In a real implementation, you would store the file and return a download URL
	// For now, we'll return a mock response
	downloadURL := fmt.Sprintf("/api/v1/exports/download/%s", filename)

	response := dto.ExportResponse{
		DownloadURL: downloadURL,
		Filename:    filename,
		Size:        int64(len(content)),
		Format:      input.Format,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339Nano),
	}

	uc.logger.Info("successfully generated transaction export",
		"format", input.Format,
		"filename", filename,
		"size", len(content),
		"count", len(transactions),
	)

	return response, nil
}

// ExecuteFromRequest executes the use case from a DTO request.
func (uc *ExportTransactionsUseCase) ExecuteFromRequest(ctx context.Context, req dto.ExportTransactionsRequest) (dto.ExportResponse, error) {
	// Validate request
	if errs := req.Validate(); len(errs) > 0 {
		return dto.ExportResponse{}, utils.NewAppError(
			"VALIDATION_ERROR",
			"Invalid request parameters",
			fiber.StatusBadRequest,
			errs,
			map[string]any{"errors": errs},
		)
	}

	// Parse optional wallet ID
	var walletID *uuid.UUID
	if req.WalletID != "" {
		parsed, err := uuid.Parse(req.WalletID)
		if err != nil {
			return dto.ExportResponse{}, utils.NewAppError(
				"VALIDATION_ERROR",
				"Invalid wallet ID format",
				fiber.StatusBadRequest,
				nil,
				map[string]any{"error": err.Error()},
			)
		}
		walletID = &parsed
	}

	// Parse optional chain
	var chain *entities.Chain
	if req.Chain != "" {
		c := entities.Chain(req.Chain)
		chain = &c
	}

	// Parse optional transaction type
	var txType *entities.TransactionType
	if req.Type != "" {
		t := entities.TransactionType(req.Type)
		txType = &t
	}

	// Parse optional status
	var status *entities.TransactionStatus
	if req.Status != "" {
		s := entities.TransactionStatus(req.Status)
		status = &s
	}

	// Parse optional dates
	var startDate, endDate *time.Time
	if req.StartDate != "" {
		t, err := time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			return dto.ExportResponse{}, utils.NewAppError(
				"VALIDATION_ERROR",
				"Invalid start date format",
				fiber.StatusBadRequest,
				nil,
				map[string]any{"error": err.Error()},
			)
		}
		startDate = &t
	}

	if req.EndDate != "" {
		t, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			return dto.ExportResponse{}, utils.NewAppError(
				"VALIDATION_ERROR",
				"Invalid end date format",
				fiber.StatusBadRequest,
				nil,
				map[string]any{"error": err.Error()},
			)
		}
		endDate = &t
	}

	input := ExportTransactionsInput{
		WalletID:  walletID,
		Chain:     chain,
		Type:      txType,
		Status:    status,
		StartDate: startDate,
		EndDate:   endDate,
		Format:    req.Format,
	}

	return uc.Execute(ctx, input)
}

// generateCSV creates a CSV export of transactions
func (uc *ExportTransactionsUseCase) generateCSV(transactions []entities.Transaction) (string, []byte, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"ID", "Wallet ID", "Chain", "Hash", "Type", "Amount", "Fee",
		"Status", "Confirmations", "From Address", "To Address",
		"Block Number", "Error Message", "Created At", "Confirmed At", "Updated At",
	}
	if err := writer.Write(header); err != nil {
		return "", nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write transaction rows
	for _, tx := range transactions {
		record := []string{
			tx.GetID().String(),
			tx.GetWalletID().String(),
			string(tx.GetChain()),
			tx.GetHash(),
			string(tx.GetType()),
			tx.GetAmount().String(),
			tx.GetFee().String(),
			string(tx.GetStatus()),
			fmt.Sprintf("%d", tx.GetConfirmations()),
			tx.GetFromAddress(),
			tx.GetToAddress(),
			"", // Block number - would need to be added to entity
			tx.GetErrorMessage(),
			tx.GetCreatedAt().UTC().Format(time.RFC3339Nano),
			"", // Confirmed at - would need to be added to entity
			tx.GetUpdatedAt().UTC().Format(time.RFC3339Nano),
		}

		if err := writer.Write(record); err != nil {
			return "", nil, fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", nil, fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	filename := fmt.Sprintf("transactions_%s.csv", time.Now().Format("20060102_150405"))
	return filename, []byte(buf.String()), nil
}

// generateJSON creates a JSON export of transactions
func (uc *ExportTransactionsUseCase) generateJSON(transactions []entities.Transaction) (string, []byte, error) {
	// Map transactions to DTO format
	dtoTransactions := make([]dto.TransactionStatusResponse, len(transactions))
	for i, tx := range transactions {
		dtoTransactions[i] = mapTransaction(tx)
	}

	// Create export structure
	exportData := map[string]interface{}{
		"exported_at":  time.Now().UTC().Format(time.RFC3339Nano),
		"total_count":  len(transactions),
		"transactions": dtoTransactions,
	}

	content, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	filename := fmt.Sprintf("transactions_%s.json", time.Now().Format("20060102_150405"))
	return filename, content, nil
}
