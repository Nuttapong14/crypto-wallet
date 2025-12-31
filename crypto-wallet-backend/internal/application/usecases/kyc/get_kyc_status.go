package kyc

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// GetKYCStatusInput encapsulates the arguments required to query status.
type GetKYCStatusInput struct {
	UserID string
}

// GetKYCStatusUseCase aggregates profile, documents, and risk score.
type GetKYCStatusUseCase struct {
	repository repositories.KYCRepository
	logger     *slog.Logger
}

// NewGetKYCStatusUseCase constructs the use case.
func NewGetKYCStatusUseCase(repo repositories.KYCRepository, logger *slog.Logger) *GetKYCStatusUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetKYCStatusUseCase{
		repository: repo,
		logger:     logger,
	}
}

// Execute loads the latest KYC state for the supplied user.
func (uc *GetKYCStatusUseCase) Execute(ctx context.Context, input GetKYCStatusInput) (*dto.KYCStatusResponse, error) {
	if uc.repository == nil {
		return nil, errors.New("get kyc status: repository not configured")
	}

	userID, err := uuid.Parse(strings.TrimSpace(input.UserID))
	if err != nil {
		return nil, utils.NewAppError(
			"INVALID_USER_ID",
			"user id must be a valid uuid",
			fiber.StatusBadRequest,
			err,
			nil,
		)
	}

	profile, err := uc.repository.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, utils.NewAppError(
				"KYC_PROFILE_MISSING",
				"kyc profile not found",
				http.StatusNotFound,
				err,
				nil,
			)
		}
		return nil, err
	}

	documents, err := uc.repository.ListDocumentsByProfile(ctx, profile.GetID())
	if err != nil {
		return nil, err
	}

	var riskScore entities.UserRiskScore
	if rs, rsErr := uc.repository.GetRiskScoreByUserID(ctx, userID); rsErr != nil {
		if !errors.Is(rsErr, repositories.ErrNotFound) {
			uc.logger.Warn("risk score fetch failed", slog.String("error", rsErr.Error()))
		}
	} else {
		riskScore = rs
	}

	response := &dto.KYCStatusResponse{
		Profile: dto.MapKYCProfile(profile),
	}

	response.Documents = make([]dto.KYCDocument, 0, len(documents))
	for _, document := range documents {
		response.Documents = append(response.Documents, dto.MapKYCDocument(document))
	}

	response.RiskScore = dto.MapRiskScore(riskScore)

	return response, nil
}
