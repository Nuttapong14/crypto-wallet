package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DocumentType enumerates supported KYC document categories.
type DocumentType string

const (
	DocumentTypePassport       DocumentType = "passport"
	DocumentTypeNationalID     DocumentType = "national_id"
	DocumentTypeDriversLicense DocumentType = "drivers_license"
	DocumentTypeProofOfAddress DocumentType = "proof_of_address"
	DocumentTypeSelfie         DocumentType = "selfie"
)

// DocumentStatus represents the lifecycle state of a KYC document.
type DocumentStatus string

const (
	DocumentStatusPending DocumentStatus = "pending"
	DocumentStatusApproved DocumentStatus = "approved"
	DocumentStatusRejected DocumentStatus = "rejected"
)

var (
	errKYCDocumentProfileIDRequired = errors.New("kyc document: profile ID is required")
	errKYCDocumentTypeInvalid       = errors.New("kyc document: type is invalid")
	errKYCDocumentPathRequired      = errors.New("kyc document: encrypted file path is required")
	errKYCDocumentNameRequired      = errors.New("kyc document: encrypted file name is required")
	errKYCDocumentHashRequired      = errors.New("kyc document: file hash is required")
	errKYCDocumentMimeRequired      = errors.New("kyc document: mime type is required")
	errKYCDocumentSizeInvalid       = errors.New("kyc document: file size must be greater than zero")
	errKYCDocumentStatusInvalid     = errors.New("kyc document: status is invalid")
)

// KYCDocument exposes behaviours required by the application layer when working with documents.
type KYCDocument interface {
	Entity
	Identifiable
	Timestamped

	GetKYCProfileID() uuid.UUID
	GetDocumentType() DocumentType
	GetEncryptedFilePath() string
	GetEncryptedFileName() string
	GetFileSize() int
	GetFileHash() string
	GetMimeType() string
	GetStatus() DocumentStatus
	GetUploadedAt() time.Time
	GetReviewedAt() *time.Time
	GetRejectionReason() string
	GetMetadata() map[string]any

	MarkApproved(at time.Time)
	MarkRejected(reason string, at time.Time)
	UpdateMetadata(metadata map[string]any)
	SetStatus(status DocumentStatus) error
	Touch(at time.Time)
}

// KYCDocumentEntity is the default implementation of the KYCDocument interface.
type KYCDocumentEntity struct {
	id                uuid.UUID
	kycProfileID      uuid.UUID
	documentType      DocumentType
	filePathEncrypted string
	fileNameEncrypted string
	fileSizeBytes     int
	fileHash          string
	mimeType          string
	status            DocumentStatus
	uploadedAt        time.Time
	reviewedAt        *time.Time
	rejectionReason   string
	metadata          map[string]any
	createdAt         time.Time
	updatedAt         time.Time
}

// KYCDocumentParams captures the fields required to construct a KYCDocumentEntity.
type KYCDocumentParams struct {
	ID                uuid.UUID
	KYCProfileID      uuid.UUID
	DocumentType      DocumentType
	FilePathEncrypted string
	FileNameEncrypted string
	FileSizeBytes     int
	FileHash          string
	MimeType          string
	Status            DocumentStatus
	UploadedAt        time.Time
	ReviewedAt        *time.Time
	RejectionReason   string
	Metadata          map[string]any
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewKYCDocumentEntity validates the supplied parameters and returns a KYCDocumentEntity instance.
func NewKYCDocumentEntity(params KYCDocumentParams) (*KYCDocumentEntity, error) {
	if params.ID == uuid.Nil {
		params.ID = uuid.New()
	}
	if params.UploadedAt.IsZero() {
		params.UploadedAt = time.Now().UTC()
	}
	if params.CreatedAt.IsZero() {
		params.CreatedAt = params.UploadedAt
	}
	if params.UpdatedAt.IsZero() {
		params.UpdatedAt = params.CreatedAt
	}
	if params.Status == "" {
		params.Status = DocumentStatusPending
	}

	entity := &KYCDocumentEntity{
		id:                params.ID,
		kycProfileID:      params.KYCProfileID,
		documentType:      params.DocumentType,
		filePathEncrypted: strings.TrimSpace(params.FilePathEncrypted),
		fileNameEncrypted: strings.TrimSpace(params.FileNameEncrypted),
		fileSizeBytes:     params.FileSizeBytes,
		fileHash:          strings.TrimSpace(params.FileHash),
		mimeType:          strings.TrimSpace(params.MimeType),
		status:            params.Status,
		uploadedAt:        params.UploadedAt,
		reviewedAt:        params.ReviewedAt,
		rejectionReason:   strings.TrimSpace(params.RejectionReason),
		metadata:          cloneMetadata(params.Metadata),
		createdAt:         params.CreatedAt,
		updatedAt:         params.UpdatedAt,
	}

	if err := entity.Validate(); err != nil {
		return nil, err
	}

	return entity, nil
}

// HydrateKYCDocumentEntity creates a KYCDocumentEntity without re-validating invariants (used for repository hydration).
func HydrateKYCDocumentEntity(params KYCDocumentParams) *KYCDocumentEntity {
	return &KYCDocumentEntity{
		id:                params.ID,
		kycProfileID:      params.KYCProfileID,
		documentType:      params.DocumentType,
		filePathEncrypted: strings.TrimSpace(params.FilePathEncrypted),
		fileNameEncrypted: strings.TrimSpace(params.FileNameEncrypted),
		fileSizeBytes:     params.FileSizeBytes,
		fileHash:          strings.TrimSpace(params.FileHash),
		mimeType:          strings.TrimSpace(params.MimeType),
		status:            params.Status,
		uploadedAt:        params.UploadedAt,
		reviewedAt:        params.ReviewedAt,
		rejectionReason:   strings.TrimSpace(params.RejectionReason),
		metadata:          cloneMetadata(params.Metadata),
		createdAt:         params.CreatedAt,
		updatedAt:         params.UpdatedAt,
	}
}

// Validate ensures the entity adheres to domain invariants.
func (d *KYCDocumentEntity) Validate() error {
	var validationErr error

	if d.kycProfileID == uuid.Nil {
		validationErr = errors.Join(validationErr, errKYCDocumentProfileIDRequired)
	}
	if !isValidDocumentType(d.documentType) {
		validationErr = errors.Join(validationErr, errKYCDocumentTypeInvalid)
	}
	if strings.TrimSpace(d.filePathEncrypted) == "" {
		validationErr = errors.Join(validationErr, errKYCDocumentPathRequired)
	}
	if strings.TrimSpace(d.fileNameEncrypted) == "" {
		validationErr = errors.Join(validationErr, errKYCDocumentNameRequired)
	}
	if strings.TrimSpace(d.fileHash) == "" {
		validationErr = errors.Join(validationErr, errKYCDocumentHashRequired)
	}
	if strings.TrimSpace(d.mimeType) == "" {
		validationErr = errors.Join(validationErr, errKYCDocumentMimeRequired)
	}
	if d.fileSizeBytes <= 0 {
		validationErr = errors.Join(validationErr, errKYCDocumentSizeInvalid)
	}
	if !isValidDocumentStatus(d.status) {
		validationErr = errors.Join(validationErr, errKYCDocumentStatusInvalid)
	}
	return validationErr
}

// Getters satisfy the KYCDocument interface.

func (d *KYCDocumentEntity) GetID() uuid.UUID {
	return d.id
}

func (d *KYCDocumentEntity) GetKYCProfileID() uuid.UUID {
	return d.kycProfileID
}

func (d *KYCDocumentEntity) GetDocumentType() DocumentType {
	return d.documentType
}

func (d *KYCDocumentEntity) GetEncryptedFilePath() string {
	return d.filePathEncrypted
}

func (d *KYCDocumentEntity) GetEncryptedFileName() string {
	return d.fileNameEncrypted
}

func (d *KYCDocumentEntity) GetFileSize() int {
	return d.fileSizeBytes
}

func (d *KYCDocumentEntity) GetFileHash() string {
	return d.fileHash
}

func (d *KYCDocumentEntity) GetMimeType() string {
	return d.mimeType
}

func (d *KYCDocumentEntity) GetStatus() DocumentStatus {
	return d.status
}

func (d *KYCDocumentEntity) GetUploadedAt() time.Time {
	return d.uploadedAt
}

func (d *KYCDocumentEntity) GetReviewedAt() *time.Time {
	return d.reviewedAt
}

func (d *KYCDocumentEntity) GetRejectionReason() string {
	return d.rejectionReason
}

func (d *KYCDocumentEntity) GetMetadata() map[string]any {
	return cloneMetadata(d.metadata)
}

func (d *KYCDocumentEntity) GetCreatedAt() time.Time {
	return d.createdAt
}

func (d *KYCDocumentEntity) GetUpdatedAt() time.Time {
	return d.updatedAt
}

// Behaviour helpers.

func (d *KYCDocumentEntity) MarkApproved(at time.Time) {
	d.status = DocumentStatusApproved
	t := normaliseTimestamp(at)
	d.reviewedAt = &t
	d.rejectionReason = ""
	d.Touch(t)
}

func (d *KYCDocumentEntity) MarkRejected(reason string, at time.Time) {
	d.status = DocumentStatusRejected
	d.rejectionReason = strings.TrimSpace(reason)
	t := normaliseTimestamp(at)
	d.reviewedAt = &t
	d.Touch(t)
}

func (d *KYCDocumentEntity) UpdateMetadata(metadata map[string]any) {
	d.metadata = cloneMetadata(metadata)
	d.Touch(time.Now().UTC())
}

func (d *KYCDocumentEntity) SetStatus(status DocumentStatus) error {
	if !isValidDocumentStatus(status) {
		return errKYCDocumentStatusInvalid
	}
	d.status = status
	d.Touch(time.Now().UTC())
	return nil
}

func (d *KYCDocumentEntity) Touch(at time.Time) {
	d.updatedAt = normaliseTimestamp(at)
}

func isValidDocumentType(docType DocumentType) bool {
	switch docType {
	case DocumentTypePassport,
		DocumentTypeNationalID,
		DocumentTypeDriversLicense,
		DocumentTypeProofOfAddress,
		DocumentTypeSelfie:
		return true
	default:
		return false
	}
}

func isValidDocumentStatus(status DocumentStatus) bool {
	switch status {
	case DocumentStatusPending, DocumentStatusApproved, DocumentStatusRejected:
		return true
	default:
		return false
	}
}

func cloneMetadata(metadata map[string]any) map[string]any {
	if metadata == nil {
		return make(map[string]any)
	}
	cloned := make(map[string]any, len(metadata))
	for key, value := range metadata {
		cloned[key] = value
	}
	return cloned
}
