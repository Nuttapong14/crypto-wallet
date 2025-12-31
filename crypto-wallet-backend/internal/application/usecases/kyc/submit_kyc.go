package kyc

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
	"github.com/crypto-wallet/backend/internal/infrastructure/external"
	"github.com/crypto-wallet/backend/internal/infrastructure/security"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// SubmitKYCInput encapsulates the arguments required to run the use case.
type SubmitKYCInput struct {
	UserID  string
	Payload dto.KYCSubmitRequest
	Email   string
}

// SubmitKYCUseCase coordinates KYC submission orchestration.
type SubmitKYCUseCase struct {
	repository repositories.KYCRepository
	encryptor  *security.AESGCMEncryptor
	provider   external.KYCProviderClient
	logger     *slog.Logger
	now        func() time.Time
}

// NewSubmitKYCUseCase constructs a SubmitKYCUseCase.
func NewSubmitKYCUseCase(
	repo repositories.KYCRepository,
	encryptor *security.AESGCMEncryptor,
	provider external.KYCProviderClient,
	logger *slog.Logger,
) *SubmitKYCUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &SubmitKYCUseCase{
		repository: repo,
		encryptor:  encryptor,
		provider:   provider,
		logger:     logger,
		now:        time.Now,
	}
}

// Execute validates and submits KYC information.
func (uc *SubmitKYCUseCase) Execute(ctx context.Context, input SubmitKYCInput) (dto.KYCProfile, error) {
	if uc.repository == nil {
		return dto.KYCProfile{}, errors.New("submit kyc: repository not configured")
	}
	if uc.encryptor == nil {
		return dto.KYCProfile{}, errors.New("submit kyc: encryptor not configured")
	}

	if errs := input.Payload.Validate(); !errs.IsEmpty() {
		return dto.KYCProfile{}, utils.NewAppError(
			"VALIDATION_ERROR",
			"kyc submission payload invalid",
			fiber.StatusBadRequest,
			nil,
			errs.ToDetails(),
		)
	}

	userID, err := uuid.Parse(strings.TrimSpace(input.UserID))
	if err != nil {
		return dto.KYCProfile{}, utils.NewAppError(
			"INVALID_USER_ID",
			"user id must be a valid uuid",
			fiber.StatusBadRequest,
			err,
			nil,
		)
	}

	nationality := strings.ToUpper(strings.TrimSpace(input.Payload.Nationality))
	addressBytes, err := json.Marshal(input.Payload.Address)
	if err != nil {
		return dto.KYCProfile{}, utils.NewAppError(
			"KYC_ADDRESS_ERROR",
			"failed to process address information",
			http.StatusInternalServerError,
			err,
			nil,
		)
	}

	now := uc.now().UTC()
	encryptedFirstName, err := uc.encryptor.EncryptToString([]byte(strings.TrimSpace(input.Payload.FirstName)), []byte(userID.String()))
	if err != nil {
		return dto.KYCProfile{}, uc.wrapEncryptionError("first name", err)
	}
	encryptedLastName, err := uc.encryptor.EncryptToString([]byte(strings.TrimSpace(input.Payload.LastName)), []byte(userID.String()))
	if err != nil {
		return dto.KYCProfile{}, uc.wrapEncryptionError("last name", err)
	}
	encryptedDOB, err := uc.encryptor.EncryptToString([]byte(strings.TrimSpace(input.Payload.DateOfBirth)), []byte(userID.String()))
	if err != nil {
		return dto.KYCProfile{}, uc.wrapEncryptionError("date of birth", err)
	}
	encryptedNationality, err := uc.encryptor.EncryptToString([]byte(nationality), []byte(userID.String()))
	if err != nil {
		return dto.KYCProfile{}, uc.wrapEncryptionError("nationality", err)
	}
	encryptedAddress, err := uc.encryptor.EncryptToString(addressBytes, []byte(userID.String()))
	if err != nil {
		return dto.KYCProfile{}, uc.wrapEncryptionError("address", err)
	}

	var profile entities.KYCProfile
	profile, err = uc.repository.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			profile, err = uc.createProfile(ctx, userID, entities.KYCProfileParams{
				UserID:                 userID,
				VerificationLevel:      entities.VerificationLevelBasic,
				Status:                 entities.KYCStatusPending,
				FirstNameEncrypted:     encryptedFirstName,
				LastNameEncrypted:      encryptedLastName,
				DateOfBirthEncrypted:   encryptedDOB,
				NationalityEncrypted:   encryptedNationality,
				AddressEncrypted:       encryptedAddress,
				SubmittedAt:            &now,
				DailyLimitUSD:          decimal.NewFromInt(5000),
				MonthlyLimitUSD:        decimal.NewFromInt(50000),
				CreatedAt:              now,
				UpdatedAt:              now,
			})
			if err != nil {
				return dto.KYCProfile{}, err
			}
		} else {
			return dto.KYCProfile{}, err
		}
	} else {
		entity, ok := profile.(*entities.KYCProfileEntity)
		if !ok {
			return dto.KYCProfile{}, errors.New("submit kyc: unexpected profile implementation")
		}
		entity.UpdatePII(encryptedFirstName, encryptedLastName, encryptedDOB, encryptedNationality, "", encryptedAddress)
		entity.MarkSubmitted(now)
		_ = entity.SetVerificationLevel(entities.VerificationLevelBasic)
		_ = entity.UpdateLimits(decimal.NewFromInt(5000), decimal.NewFromInt(50000))
		if err := uc.repository.UpdateProfile(ctx, entity); err != nil {
			return dto.KYCProfile{}, err
		}
		profile = entity
	}

	if uc.provider != nil {
		uc.submitToProvider(ctx, input, profile, now)
	}

	return dto.MapKYCProfile(profile), nil
}

func (uc *SubmitKYCUseCase) createProfile(
	ctx context.Context,
	userID uuid.UUID,
	params entities.KYCProfileParams,
) (entities.KYCProfile, error) {
	entity, err := entities.NewKYCProfileEntity(params)
	if err != nil {
		return nil, utils.NewAppError(
			"KYC_PROFILE_INVALID",
			"failed to create kyc profile",
			http.StatusInternalServerError,
			err,
			nil,
		)
	}

	if err := uc.repository.CreateProfile(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (uc *SubmitKYCUseCase) submitToProvider(
	ctx context.Context,
	input SubmitKYCInput,
	profile entities.KYCProfile,
	now time.Time,
) {
	payload := external.KYCSubmissionPayload{
		ExternalUserID: profile.GetUserID().String(),
		Email:          strings.TrimSpace(input.Email),
		FirstName:      strings.TrimSpace(input.Payload.FirstName),
		LastName:       strings.TrimSpace(input.Payload.LastName),
		DateOfBirth:    strings.TrimSpace(input.Payload.DateOfBirth),
		Nationality:    strings.ToUpper(strings.TrimSpace(input.Payload.Nationality)),
		Metadata: map[string]string{
			"submitted_at": now.Format(time.RFC3339),
		},
	}

	result, err := uc.provider.SubmitApplication(ctx, payload)
	if err != nil {
		uc.logger.Warn("kyc provider submission failed", slog.String("error", err.Error()))
		return
	}

	entity, ok := profile.(*entities.KYCProfileEntity)
	if !ok {
		return
	}

	if result.ApplicationID != "" {
		notes := strings.TrimSpace(entity.GetReviewerNotes())
		externalRef := "External application ID: " + result.ApplicationID
		if !strings.Contains(notes, result.ApplicationID) {
			if notes != "" {
				notes += " | "
			}
			notes += externalRef
		}
		entity.SetReviewerNotes(notes)
		entity.Touch(now)
		_ = uc.repository.UpdateProfile(ctx, entity)
	}
}

func (uc *SubmitKYCUseCase) wrapEncryptionError(field string, err error) error {
	return utils.NewAppError(
		"KYC_ENCRYPTION_ERROR",
		"failed to protect sensitive information",
		http.StatusInternalServerError,
		err,
		map[string]any{"field": field},
	)
}
