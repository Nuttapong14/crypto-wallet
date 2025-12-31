package kyc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
	"github.com/crypto-wallet/backend/internal/infrastructure/external"
	"github.com/crypto-wallet/backend/internal/infrastructure/security"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// UploadDocumentInput encapsulates the arguments required to upload a document.
type UploadDocumentInput struct {
	UserID       string
	DocumentType string
	FileName     string
	MimeType     string
	Content      []byte
}

// UploadDocumentUseCase handles verification document uploads.
type UploadDocumentUseCase struct {
	repository repositories.KYCRepository
	encryptor  *security.AESGCMEncryptor
	provider   external.KYCProviderClient
	logger     *slog.Logger
	now        func() time.Time
}

// NewUploadDocumentUseCase constructs an UploadDocumentUseCase.
func NewUploadDocumentUseCase(
	repo repositories.KYCRepository,
	encryptor *security.AESGCMEncryptor,
	provider external.KYCProviderClient,
	logger *slog.Logger,
) *UploadDocumentUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &UploadDocumentUseCase{
		repository: repo,
		encryptor:  encryptor,
		provider:   provider,
		logger:     logger,
		now:        time.Now,
	}
}

// Execute validates the request and stores a new document record.
func (uc *UploadDocumentUseCase) Execute(ctx context.Context, input UploadDocumentInput) (*dto.KYCDocumentUploadResponse, error) {
	if uc.repository == nil {
		return nil, errors.New("upload document: repository not configured")
	}
	if uc.encryptor == nil {
		return nil, errors.New("upload document: encryptor not configured")
	}

	if len(input.Content) == 0 {
		return nil, utils.NewAppError(
			"DOCUMENT_EMPTY",
			"no document content provided",
			fiber.StatusBadRequest,
			nil,
			nil,
		)
	}

	docType := entities.DocumentType(strings.TrimSpace(input.DocumentType))
	if !isSupportedDocumentType(docType) {
		return nil, utils.NewAppError(
			"DOCUMENT_TYPE_INVALID",
			"document type is not supported",
			fiber.StatusBadRequest,
			nil,
			map[string]any{"documentType": input.DocumentType},
		)
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
				"submit kyc information before uploading documents",
				fiber.StatusPreconditionFailed,
				err,
				nil,
			)
		}
		return nil, err
	}

	now := uc.now().UTC()
	hash := sha256.Sum256(input.Content)
	hashHex := hex.EncodeToString(hash[:])

	encryptedFileName, err := uc.encryptor.EncryptToString([]byte(strings.TrimSpace(input.FileName)), []byte(userID.String()))
	if err != nil {
		return nil, uc.wrapEncryptionError("file name", err)
	}

	remotePath := hashHex
	var providerResult *external.KYCDocumentUploadResult
	if uc.provider != nil {
		result, uploadErr := uc.provider.UploadDocument(ctx, external.KYCDocumentUploadPayload{
			ApplicationID: profile.GetID().String(),
			DocumentType:  string(docType),
			FileName:      strings.TrimSpace(input.FileName),
			MimeType:      strings.TrimSpace(input.MimeType),
			Content:       input.Content,
		})
		if uploadErr != nil {
			uc.logger.Warn("kyc provider document upload failed", slog.String("error", uploadErr.Error()))
		} else {
			providerResult = result
			if strings.TrimSpace(result.DocumentID) != "" {
				remotePath = result.DocumentID
			}
		}
	}

	encryptedPath, err := uc.encryptor.EncryptToString([]byte(remotePath), []byte(userID.String()))
	if err != nil {
		return nil, uc.wrapEncryptionError("storage path", err)
	}

	entity, err := entities.NewKYCDocumentEntity(entities.KYCDocumentParams{
		KYCProfileID:      profile.GetID(),
		DocumentType:      docType,
		FilePathEncrypted: encryptedPath,
		FileNameEncrypted: encryptedFileName,
		FileSizeBytes:     len(input.Content),
		FileHash:          hashHex,
		MimeType:          strings.TrimSpace(input.MimeType),
		Status:            entities.DocumentStatusPending,
		UploadedAt:        now,
		CreatedAt:         now,
		UpdatedAt:         now,
		Metadata: map[string]any{
			"originalFileName": strings.TrimSpace(input.FileName),
			"mimeType":         strings.TrimSpace(input.MimeType),
			"hash":             hashHex,
		},
	})
	if err != nil {
		return nil, utils.NewAppError(
			"DOCUMENT_INVALID",
			"failed to prepare document metadata",
			http.StatusInternalServerError,
			err,
			nil,
		)
	}

	if err := uc.repository.CreateDocument(ctx, entity); err != nil {
		return nil, err
	}

	response := &dto.KYCDocumentUploadResponse{
		Document: dto.MapKYCDocument(entity),
	}
	if providerResult != nil {
		response.Provider.DocumentID = providerResult.DocumentID
		response.Provider.Status = providerResult.Status
	}

	return response, nil
}

func (uc *UploadDocumentUseCase) wrapEncryptionError(field string, err error) error {
	return utils.NewAppError(
		"KYC_ENCRYPTION_ERROR",
		"failed to protect sensitive information",
		http.StatusInternalServerError,
		err,
		map[string]any{"field": field},
	)
}

func isSupportedDocumentType(docType entities.DocumentType) bool {
	switch docType {
	case entities.DocumentTypePassport,
		entities.DocumentTypeNationalID,
		entities.DocumentTypeDriversLicense,
		entities.DocumentTypeProofOfAddress,
		entities.DocumentTypeSelfie:
		return true
	default:
		return false
	}
}
