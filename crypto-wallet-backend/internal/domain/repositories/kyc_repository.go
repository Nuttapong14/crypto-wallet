package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/domain/entities"
)

// KYCRepository defines persistence operations for compliance entities.
type KYCRepository interface {
	GetProfileByUserID(ctx context.Context, userID uuid.UUID) (entities.KYCProfile, error)
	CreateProfile(ctx context.Context, profile *entities.KYCProfileEntity) error
	UpdateProfile(ctx context.Context, profile entities.KYCProfile) error

	CreateDocument(ctx context.Context, document *entities.KYCDocumentEntity) error
	GetDocumentByID(ctx context.Context, id uuid.UUID) (entities.KYCDocument, error)
	ListDocumentsByProfile(ctx context.Context, profileID uuid.UUID) ([]entities.KYCDocument, error)
	UpdateDocument(ctx context.Context, document entities.KYCDocument) error

	GetRiskScoreByUserID(ctx context.Context, userID uuid.UUID) (entities.UserRiskScore, error)
	UpsertRiskScore(ctx context.Context, score *entities.UserRiskScoreEntity) error
}
