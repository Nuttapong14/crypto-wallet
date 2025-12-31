package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
)

const (
	selectKYCProfile = `
SELECT
	id,
	user_id,
	verification_level,
	status,
	first_name_encrypted,
	last_name_encrypted,
	date_of_birth_encrypted,
	nationality_encrypted,
	document_number_encrypted,
	address_encrypted,
	submitted_at,
	reviewed_at,
	approved_at,
	expires_at,
	rejection_reason,
	reviewer_notes,
	daily_limit_usd,
	monthly_limit_usd,
	created_at,
	updated_at
FROM kyc_profiles`

	selectKYCDocument = `
SELECT
	id,
	kyc_profile_id,
	document_type,
	file_path_encrypted,
	file_name_encrypted,
	file_size_bytes,
	file_hash,
	mime_type,
	status,
	uploaded_at,
	reviewed_at,
	rejection_reason,
	metadata,
	created_at,
	updated_at
FROM kyc_documents`

	selectRiskScore = `
SELECT
	id,
	user_id,
	risk_score,
	risk_level,
	risk_factors,
	aml_hits,
	last_screening_at,
	next_review_at,
	created_at,
	updated_at
FROM user_risk_scores`
)

// KYCRepository persists compliance entities in PostgreSQL.
type KYCRepository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewKYCRepository constructs a KYCRepository backed by the provided pool.
func NewKYCRepository(pool *pgxpool.Pool, logger *slog.Logger) *KYCRepository {
	if logger == nil {
		logger = slog.Default()
	}
	return &KYCRepository{
		pool:   pool,
		logger: logger,
	}
}

// GetProfileByUserID returns the KYC profile for the supplied user ID.
func (r *KYCRepository) GetProfileByUserID(ctx context.Context, userID uuid.UUID) (entities.KYCProfile, error) {
	if r.pool == nil {
		return nil, errors.New("kyc repository: pool not configured")
	}

	row := r.pool.QueryRow(ctx, selectKYCProfile+" WHERE user_id = $1", userID)
	return r.scanKYCProfile(row)
}

// CreateProfile inserts a new KYC profile record.
func (r *KYCRepository) CreateProfile(ctx context.Context, profile *entities.KYCProfileEntity) error {
	if r.pool == nil {
		return errors.New("kyc repository: pool not configured")
	}
	if profile == nil {
		return errors.New("kyc repository: profile entity is nil")
	}

	query := `
INSERT INTO kyc_profiles (
	id,
	user_id,
	verification_level,
	status,
	first_name_encrypted,
	last_name_encrypted,
	date_of_birth_encrypted,
	nationality_encrypted,
	document_number_encrypted,
	address_encrypted,
	submitted_at,
	reviewed_at,
	approved_at,
	expires_at,
	rejection_reason,
	reviewer_notes,
	daily_limit_usd,
	monthly_limit_usd,
	created_at,
	updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20
)`

	_, err := r.pool.Exec(
		ctx,
		query,
		profile.GetID(),
		profile.GetUserID(),
		profile.GetVerificationLevel(),
		profile.GetStatus(),
		profile.GetEncryptedFirstName(),
		profile.GetEncryptedLastName(),
		profile.GetEncryptedDateOfBirth(),
		profile.GetEncryptedNationality(),
		profile.GetEncryptedDocumentNumber(),
		profile.GetEncryptedAddress(),
		profile.GetSubmittedAt(),
		profile.GetReviewedAt(),
		profile.GetApprovedAt(),
		profile.GetExpiresAt(),
		nullIfEmpty(profile.GetRejectionReason()),
		nullIfEmpty(profile.GetReviewerNotes()),
		profile.GetDailyLimitUSD().String(),
		profile.GetMonthlyLimitUSD().String(),
		profile.GetCreatedAt(),
		profile.GetUpdatedAt(),
	)
	return mapPGError(err)
}

// UpdateProfile persists changes to an existing KYC profile.
func (r *KYCRepository) UpdateProfile(ctx context.Context, profile entities.KYCProfile) error {
	if r.pool == nil {
		return errors.New("kyc repository: pool not configured")
	}
	if profile == nil {
		return errors.New("kyc repository: profile entity is nil")
	}

	query := `
UPDATE kyc_profiles SET
	verification_level = $1,
	status = $2,
	first_name_encrypted = $3,
	last_name_encrypted = $4,
	date_of_birth_encrypted = $5,
	nationality_encrypted = $6,
	document_number_encrypted = $7,
	address_encrypted = $8,
	submitted_at = $9,
	reviewed_at = $10,
	approved_at = $11,
	expires_at = $12,
	rejection_reason = $13,
	reviewer_notes = $14,
	daily_limit_usd = $15,
	monthly_limit_usd = $16,
	updated_at = $17
WHERE id = $18`

	cmd, err := r.pool.Exec(
		ctx,
		query,
		profile.GetVerificationLevel(),
		profile.GetStatus(),
		profile.GetEncryptedFirstName(),
		profile.GetEncryptedLastName(),
		profile.GetEncryptedDateOfBirth(),
		profile.GetEncryptedNationality(),
		profile.GetEncryptedDocumentNumber(),
		profile.GetEncryptedAddress(),
		profile.GetSubmittedAt(),
		profile.GetReviewedAt(),
		profile.GetApprovedAt(),
		profile.GetExpiresAt(),
		nullIfEmpty(profile.GetRejectionReason()),
		nullIfEmpty(profile.GetReviewerNotes()),
		profile.GetDailyLimitUSD().String(),
		profile.GetMonthlyLimitUSD().String(),
		time.Now().UTC(),
		profile.GetID(),
	)
	if err != nil {
		return mapPGError(err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

// CreateDocument stores a new KYC document.
func (r *KYCRepository) CreateDocument(ctx context.Context, document *entities.KYCDocumentEntity) error {
	if r.pool == nil {
		return errors.New("kyc repository: pool not configured")
	}
	if document == nil {
		return errors.New("kyc repository: document entity is nil")
	}

	query := `
INSERT INTO kyc_documents (
	id,
	kyc_profile_id,
	document_type,
	file_path_encrypted,
	file_name_encrypted,
	file_size_bytes,
	file_hash,
	mime_type,
	status,
	uploaded_at,
	reviewed_at,
	rejection_reason,
	metadata,
	created_at,
	updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15
)`

	metadataJSON, err := marshalMetadata(document.GetMetadata())
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(
		ctx,
		query,
		document.GetID(),
		document.GetKYCProfileID(),
		document.GetDocumentType(),
		document.GetEncryptedFilePath(),
		document.GetEncryptedFileName(),
		document.GetFileSize(),
		document.GetFileHash(),
		document.GetMimeType(),
		document.GetStatus(),
		document.GetUploadedAt(),
		document.GetReviewedAt(),
		nullIfEmpty(document.GetRejectionReason()),
		metadataJSON,
		document.GetCreatedAt(),
		document.GetUpdatedAt(),
	)
	return mapPGError(err)
}

// GetDocumentByID returns a document by primary key.
func (r *KYCRepository) GetDocumentByID(ctx context.Context, id uuid.UUID) (entities.KYCDocument, error) {
	if r.pool == nil {
		return nil, errors.New("kyc repository: pool not configured")
	}

	row := r.pool.QueryRow(ctx, selectKYCDocument+" WHERE id = $1", id)
	return r.scanKYCDocument(row)
}

// ListDocumentsByProfile returns documents for the supplied profile.
func (r *KYCRepository) ListDocumentsByProfile(ctx context.Context, profileID uuid.UUID) ([]entities.KYCDocument, error) {
	if r.pool == nil {
		return nil, errors.New("kyc repository: pool not configured")
	}

	rows, err := r.pool.Query(ctx, selectKYCDocument+" WHERE kyc_profile_id = $1 ORDER BY uploaded_at DESC", profileID)
	if err != nil {
		return nil, mapPGError(err)
	}
	defer rows.Close()

	var documents []entities.KYCDocument
	for rows.Next() {
		document, scanErr := r.scanKYCDocument(rows)
		if scanErr != nil {
			return nil, mapPGError(scanErr)
		}
		documents = append(documents, document)
	}
	if rows.Err() != nil {
		return nil, mapPGError(rows.Err())
	}
	return documents, nil
}

// UpdateDocument persists changes to an existing document record.
func (r *KYCRepository) UpdateDocument(ctx context.Context, document entities.KYCDocument) error {
	if r.pool == nil {
		return errors.New("kyc repository: pool not configured")
	}
	if document == nil {
		return errors.New("kyc repository: document entity is nil")
	}

	metadataJSON, err := marshalMetadata(document.GetMetadata())
	if err != nil {
		return err
	}

	query := `
UPDATE kyc_documents SET
	status = $1,
	reviewed_at = $2,
	rejection_reason = $3,
	metadata = $4,
	updated_at = $5
WHERE id = $6`

	cmd, err := r.pool.Exec(
		ctx,
		query,
		document.GetStatus(),
		document.GetReviewedAt(),
		nullIfEmpty(document.GetRejectionReason()),
		metadataJSON,
		time.Now().UTC(),
		document.GetID(),
	)
	if err != nil {
		return mapPGError(err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrNotFound
	}
	return nil
}

// GetRiskScoreByUserID returns the risk score for a user.
func (r *KYCRepository) GetRiskScoreByUserID(ctx context.Context, userID uuid.UUID) (entities.UserRiskScore, error) {
	if r.pool == nil {
		return nil, errors.New("kyc repository: pool not configured")
	}

	row := r.pool.QueryRow(ctx, selectRiskScore+" WHERE user_id = $1", userID)
	return r.scanRiskScore(row)
}

// UpsertRiskScore creates or updates a user risk score record.
func (r *KYCRepository) UpsertRiskScore(ctx context.Context, score *entities.UserRiskScoreEntity) error {
	if r.pool == nil {
		return errors.New("kyc repository: pool not configured")
	}
	if score == nil {
		return errors.New("kyc repository: risk score entity is nil")
	}

	query := `
INSERT INTO user_risk_scores (
	id,
	user_id,
	risk_score,
	risk_level,
	risk_factors,
	aml_hits,
	last_screening_at,
	next_review_at,
	created_at,
	updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10
)
ON CONFLICT (user_id) DO UPDATE SET
	risk_score = EXCLUDED.risk_score,
	risk_level = EXCLUDED.risk_level,
	risk_factors = EXCLUDED.risk_factors,
	aml_hits = EXCLUDED.aml_hits,
	last_screening_at = EXCLUDED.last_screening_at,
	next_review_at = EXCLUDED.next_review_at,
	updated_at = EXCLUDED.updated_at`

	_, err := r.pool.Exec(
		ctx,
		query,
		score.GetID(),
		score.GetUserID(),
		score.GetScore(),
		score.GetLevel(),
		toJSONStringArray(score.GetRiskFactors()),
		toJSONStringArray(score.GetAMLHits()),
		score.GetLastScreeningAt(),
		score.GetNextReviewAt(),
		score.GetCreatedAt(),
		score.GetUpdatedAt(),
	)
	return mapPGError(err)
}

func (r *KYCRepository) scanKYCProfile(row pgx.Row) (entities.KYCProfile, error) {
	var (
		id                  uuid.UUID
		userID              uuid.UUID
		level               string
		status              string
		firstName           sql.NullString
		lastName            sql.NullString
		dateOfBirth         sql.NullString
		nationality         sql.NullString
		documentNumber      sql.NullString
		address             sql.NullString
		submittedAt         sql.NullTime
		reviewedAt          sql.NullTime
		approvedAt          sql.NullTime
		expiresAt           sql.NullTime
		rejectionReason     sql.NullString
		reviewerNotes       sql.NullString
		dailyLimitStr       string
		monthlyLimitStr     string
		createdAt           time.Time
		updatedAt           time.Time
	)

	err := row.Scan(
		&id,
		&userID,
		&level,
		&status,
		&firstName,
		&lastName,
		&dateOfBirth,
		&nationality,
		&documentNumber,
		&address,
		&submittedAt,
		&reviewedAt,
		&approvedAt,
		&expiresAt,
		&rejectionReason,
		&reviewerNotes,
		&dailyLimitStr,
		&monthlyLimitStr,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}

	dailyLimit, err := decimal.NewFromString(strings.TrimSpace(dailyLimitStr))
	if err != nil {
		return nil, fmt.Errorf("kyc repository: parse daily limit: %w", err)
	}
	monthlyLimit, err := decimal.NewFromString(strings.TrimSpace(monthlyLimitStr))
	if err != nil {
		return nil, fmt.Errorf("kyc repository: parse monthly limit: %w", err)
	}

	params := entities.KYCProfileParams{
		ID:                     id,
		UserID:                 userID,
		VerificationLevel:      entities.VerificationLevel(level),
		Status:                 entities.KYCStatus(status),
		FirstNameEncrypted:     firstName.String,
		LastNameEncrypted:      lastName.String,
		DateOfBirthEncrypted:   dateOfBirth.String,
		NationalityEncrypted:   nationality.String,
		DocumentNumberEncrypted: documentNumber.String,
		AddressEncrypted:       address.String,
		SubmittedAt:            nullTimePtr(submittedAt),
		ReviewedAt:             nullTimePtr(reviewedAt),
		ApprovedAt:             nullTimePtr(approvedAt),
		ExpiresAt:              nullTimePtr(expiresAt),
		RejectionReason:        rejectionReason.String,
		ReviewerNotes:          reviewerNotes.String,
		DailyLimitUSD:          dailyLimit,
		MonthlyLimitUSD:        monthlyLimit,
		CreatedAt:              createdAt,
		UpdatedAt:              updatedAt,
	}

	return entities.HydrateKYCProfileEntity(params), nil
}

func (r *KYCRepository) scanKYCDocument(row pgx.Row) (entities.KYCDocument, error) {
	var (
		id            uuid.UUID
		profileID     uuid.UUID
		docType       string
		filePath      string
		fileName      string
		fileSize      int
		fileHash      string
		mimeType      string
		status        string
		uploadedAt    time.Time
		reviewedAt    sql.NullTime
		rejection     sql.NullString
		metadataBytes []byte
		createdAt     time.Time
		updatedAt     time.Time
	)

	err := row.Scan(
		&id,
		&profileID,
		&docType,
		&filePath,
		&fileName,
		&fileSize,
		&fileHash,
		&mimeType,
		&status,
		&uploadedAt,
		&reviewedAt,
		&rejection,
		&metadataBytes,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}

	metadata := make(map[string]any)
	if len(metadataBytes) > 0 {
		if unmarshalErr := json.Unmarshal(metadataBytes, &metadata); unmarshalErr != nil {
			return nil, fmt.Errorf("kyc repository: decode metadata: %w", unmarshalErr)
		}
	}

	params := entities.KYCDocumentParams{
		ID:                id,
		KYCProfileID:      profileID,
		DocumentType:      entities.DocumentType(docType),
		FilePathEncrypted: filePath,
		FileNameEncrypted: fileName,
		FileSizeBytes:     fileSize,
		FileHash:          fileHash,
		MimeType:          mimeType,
		Status:            entities.DocumentStatus(status),
		UploadedAt:        uploadedAt,
		ReviewedAt:        nullTimePtr(reviewedAt),
		RejectionReason:   rejection.String,
		Metadata:          metadata,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}

	return entities.HydrateKYCDocumentEntity(params), nil
}

func (r *KYCRepository) scanRiskScore(row pgx.Row) (entities.UserRiskScore, error) {
	var (
		id              uuid.UUID
		userID          uuid.UUID
		score           int
		level           string
		riskFactorsJSON []byte
		amlHitsJSON     []byte
		lastScreening   sql.NullTime
		nextReview      time.Time
		createdAt       time.Time
		updatedAt       time.Time
	)

	err := row.Scan(
		&id,
		&userID,
		&score,
		&level,
		&riskFactorsJSON,
		&amlHitsJSON,
		&lastScreening,
		&nextReview,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}

	riskFactors, err := decodeStringArray(riskFactorsJSON)
	if err != nil {
		return nil, err
	}
	amlHits, err := decodeStringArray(amlHitsJSON)
	if err != nil {
		return nil, err
	}

	params := entities.UserRiskScoreParams{
		ID:              id,
		UserID:          userID,
		RiskScore:       score,
		RiskLevel:       entities.RiskLevel(level),
		RiskFactors:     riskFactors,
		AMLHits:         amlHits,
		LastScreeningAt: nullTimePtr(lastScreening),
		NextReviewAt:    nextReview,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}

	return entities.HydrateUserRiskScoreEntity(params), nil
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time.UTC()
	return &t
}

func marshalMetadata(metadata map[string]any) ([]byte, error) {
	if metadata == nil {
		return json.Marshal(map[string]any{})
	}
	return json.Marshal(metadata)
}

func decodeStringArray(payload []byte) ([]string, error) {
	if len(payload) == 0 {
		return []string{}, nil
	}
	var values []string
	if err := json.Unmarshal(payload, &values); err != nil {
		return nil, fmt.Errorf("kyc repository: decode string array: %w", err)
	}
	return values, nil
}

func toJSONStringArray(values []string) []byte {
	if values == nil {
		values = []string{}
	}
	data, err := json.Marshal(values)
	if err != nil {
		return []byte("[]")
	}
	return data
}
