package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// VerificationLevel represents the KYC verification tiers available to users.
type VerificationLevel string

const (
	VerificationLevelUnverified VerificationLevel = "unverified"
	VerificationLevelBasic      VerificationLevel = "basic"
	VerificationLevelFull       VerificationLevel = "full"
)

// KYCStatus models the current review stage for a KYC profile.
type KYCStatus string

const (
	KYCStatusNotStarted  KYCStatus = "not_started"
	KYCStatusPending     KYCStatus = "pending"
	KYCStatusUnderReview KYCStatus = "under_review"
	KYCStatusApproved    KYCStatus = "approved"
	KYCStatusRejected    KYCStatus = "rejected"
	KYCStatusExpired     KYCStatus = "expired"
)

var (
	errKYCUserIDRequired            = errors.New("kyc profile: user ID is required")
	errKYCLevelInvalid              = errors.New("kyc profile: verification level is invalid")
	errKYCStatusInvalid             = errors.New("kyc profile: status is invalid")
	errKYCLimitNegative             = errors.New("kyc profile: limits must be non-negative")
	errKYCTransitionInvalid         = errors.New("kyc profile: invalid status transition")
	errKYCSubmissionTimestampNeeded = errors.New("kyc profile: submitted_at must be set when status is pending or under_review")
)

// KYCProfile exposes the behaviours required by the application layer when working with KYC profiles.
type KYCProfile interface {
	Entity
	Identifiable
	Timestamped

	GetUserID() uuid.UUID
	GetVerificationLevel() VerificationLevel
	GetStatus() KYCStatus
	GetDailyLimitUSD() decimal.Decimal
	GetMonthlyLimitUSD() decimal.Decimal
	GetSubmittedAt() *time.Time
	GetReviewedAt() *time.Time
	GetApprovedAt() *time.Time
	GetExpiresAt() *time.Time
	GetRejectionReason() string
	GetReviewerNotes() string
	GetEncryptedFirstName() string
	GetEncryptedLastName() string
	GetEncryptedDateOfBirth() string
	GetEncryptedNationality() string
	GetEncryptedDocumentNumber() string
	GetEncryptedAddress() string

	SetVerificationLevel(level VerificationLevel) error
	SetStatus(status KYCStatus) error
	UpdateLimits(daily, monthly decimal.Decimal) error
	MarkSubmitted(at time.Time)
	MarkReviewed(at time.Time)
	MarkApproved(at time.Time)
	MarkExpired(at time.Time)
	Reject(reason string, notes string)
	UpdatePII(firstName, lastName, dob, nationality, docNumber, address string)
	Touch(at time.Time)
}

// KYCProfileEntity is the default implementation of the KYCProfile interface.
type KYCProfileEntity struct {
	id                    uuid.UUID
	userID                uuid.UUID
	verificationLevel     VerificationLevel
	status                KYCStatus
	firstNameEncrypted    string
	lastNameEncrypted     string
	dateOfBirthEncrypted  string
	nationalityEncrypted  string
	documentNumberEncrypted string
	addressEncrypted      string
	submittedAt           *time.Time
	reviewedAt            *time.Time
	approvedAt            *time.Time
	expiresAt             *time.Time
	rejectionReason       string
	reviewerNotes         string
	dailyLimitUSD         decimal.Decimal
	monthlyLimitUSD       decimal.Decimal
	createdAt             time.Time
	updatedAt             time.Time
}

// KYCProfileParams captures the fields required to construct a KYCProfileEntity.
type KYCProfileParams struct {
	ID                     uuid.UUID
	UserID                 uuid.UUID
	VerificationLevel      VerificationLevel
	Status                 KYCStatus
	FirstNameEncrypted     string
	LastNameEncrypted      string
	DateOfBirthEncrypted   string
	NationalityEncrypted   string
	DocumentNumberEncrypted string
	AddressEncrypted       string
	SubmittedAt            *time.Time
	ReviewedAt             *time.Time
	ApprovedAt             *time.Time
	ExpiresAt              *time.Time
	RejectionReason        string
	ReviewerNotes          string
	DailyLimitUSD          decimal.Decimal
	MonthlyLimitUSD        decimal.Decimal
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// NewKYCProfileEntity validates the supplied parameters and returns a new KYCProfileEntity instance.
func NewKYCProfileEntity(params KYCProfileParams) (*KYCProfileEntity, error) {
	if params.ID == uuid.Nil {
		params.ID = uuid.New()
	}
	if params.CreatedAt.IsZero() {
		params.CreatedAt = time.Now().UTC()
	}
	if params.UpdatedAt.IsZero() {
		params.UpdatedAt = params.CreatedAt
	}
	if params.VerificationLevel == "" {
		params.VerificationLevel = VerificationLevelUnverified
	}
	if params.Status == "" {
		params.Status = KYCStatusNotStarted
	}
	if params.DailyLimitUSD.IsZero() && params.MonthlyLimitUSD.IsZero() {
		params.DailyLimitUSD = decimal.NewFromInt(500)
		params.MonthlyLimitUSD = decimal.NewFromInt(5000)
	}

	entity := &KYCProfileEntity{
		id:                     params.ID,
		userID:                 params.UserID,
		verificationLevel:      params.VerificationLevel,
		status:                 params.Status,
		firstNameEncrypted:     strings.TrimSpace(params.FirstNameEncrypted),
		lastNameEncrypted:      strings.TrimSpace(params.LastNameEncrypted),
		dateOfBirthEncrypted:   strings.TrimSpace(params.DateOfBirthEncrypted),
		nationalityEncrypted:   strings.TrimSpace(params.NationalityEncrypted),
		documentNumberEncrypted: strings.TrimSpace(params.DocumentNumberEncrypted),
		addressEncrypted:       strings.TrimSpace(params.AddressEncrypted),
		submittedAt:            params.SubmittedAt,
		reviewedAt:             params.ReviewedAt,
		approvedAt:             params.ApprovedAt,
		expiresAt:              params.ExpiresAt,
		rejectionReason:        strings.TrimSpace(params.RejectionReason),
		reviewerNotes:          strings.TrimSpace(params.ReviewerNotes),
		dailyLimitUSD:          params.DailyLimitUSD,
		monthlyLimitUSD:        params.MonthlyLimitUSD,
		createdAt:              params.CreatedAt,
		updatedAt:              params.UpdatedAt,
	}

	if err := entity.Validate(); err != nil {
		return nil, err
	}

	return entity, nil
}

// HydrateKYCProfileEntity creates a KYCProfileEntity without re-validating invariants (used for repository hydration).
func HydrateKYCProfileEntity(params KYCProfileParams) *KYCProfileEntity {
	return &KYCProfileEntity{
		id:                     params.ID,
		userID:                 params.UserID,
		verificationLevel:      params.VerificationLevel,
		status:                 params.Status,
		firstNameEncrypted:     strings.TrimSpace(params.FirstNameEncrypted),
		lastNameEncrypted:      strings.TrimSpace(params.LastNameEncrypted),
		dateOfBirthEncrypted:   strings.TrimSpace(params.DateOfBirthEncrypted),
		nationalityEncrypted:   strings.TrimSpace(params.NationalityEncrypted),
		documentNumberEncrypted: strings.TrimSpace(params.DocumentNumberEncrypted),
		addressEncrypted:       strings.TrimSpace(params.AddressEncrypted),
		submittedAt:            params.SubmittedAt,
		reviewedAt:             params.ReviewedAt,
		approvedAt:             params.ApprovedAt,
		expiresAt:              params.ExpiresAt,
		rejectionReason:        strings.TrimSpace(params.RejectionReason),
		reviewerNotes:          strings.TrimSpace(params.ReviewerNotes),
		dailyLimitUSD:          params.DailyLimitUSD,
		monthlyLimitUSD:        params.MonthlyLimitUSD,
		createdAt:              params.CreatedAt,
		updatedAt:              params.UpdatedAt,
	}
}

// Validate ensures the entity adheres to domain invariants.
func (k *KYCProfileEntity) Validate() error {
	var validationErr error

	if k.userID == uuid.Nil {
		validationErr = errors.Join(validationErr, errKYCUserIDRequired)
	}

	if !isValidVerificationLevel(k.verificationLevel) {
		validationErr = errors.Join(validationErr, errKYCLevelInvalid)
	}

	if !isValidKYCStatus(k.status) {
		validationErr = errors.Join(validationErr, errKYCStatusInvalid)
	}

	if k.dailyLimitUSD.IsNegative() || k.monthlyLimitUSD.IsNegative() {
		validationErr = errors.Join(validationErr, errKYCLimitNegative)
	}

	if (k.status == KYCStatusPending || k.status == KYCStatusUnderReview) && k.submittedAt == nil {
		validationErr = errors.Join(validationErr, errKYCSubmissionTimestampNeeded)
	}

	return validationErr
}

// Getter implementations satisfy the KYCProfile interface.

func (k *KYCProfileEntity) GetID() uuid.UUID {
	return k.id
}

func (k *KYCProfileEntity) GetUserID() uuid.UUID {
	return k.userID
}

func (k *KYCProfileEntity) GetVerificationLevel() VerificationLevel {
	return k.verificationLevel
}

func (k *KYCProfileEntity) GetStatus() KYCStatus {
	return k.status
}

func (k *KYCProfileEntity) GetDailyLimitUSD() decimal.Decimal {
	return k.dailyLimitUSD
}

func (k *KYCProfileEntity) GetMonthlyLimitUSD() decimal.Decimal {
	return k.monthlyLimitUSD
}

func (k *KYCProfileEntity) GetSubmittedAt() *time.Time {
	return k.submittedAt
}

func (k *KYCProfileEntity) GetReviewedAt() *time.Time {
	return k.reviewedAt
}

func (k *KYCProfileEntity) GetApprovedAt() *time.Time {
	return k.approvedAt
}

func (k *KYCProfileEntity) GetExpiresAt() *time.Time {
	return k.expiresAt
}

func (k *KYCProfileEntity) GetRejectionReason() string {
	return k.rejectionReason
}

func (k *KYCProfileEntity) GetReviewerNotes() string {
	return k.reviewerNotes
}

func (k *KYCProfileEntity) GetEncryptedFirstName() string {
	return k.firstNameEncrypted
}

func (k *KYCProfileEntity) GetEncryptedLastName() string {
	return k.lastNameEncrypted
}

func (k *KYCProfileEntity) GetEncryptedDateOfBirth() string {
	return k.dateOfBirthEncrypted
}

func (k *KYCProfileEntity) GetEncryptedNationality() string {
	return k.nationalityEncrypted
}

func (k *KYCProfileEntity) GetEncryptedDocumentNumber() string {
	return k.documentNumberEncrypted
}

func (k *KYCProfileEntity) GetEncryptedAddress() string {
	return k.addressEncrypted
}

func (k *KYCProfileEntity) GetCreatedAt() time.Time {
	return k.createdAt
}

func (k *KYCProfileEntity) GetUpdatedAt() time.Time {
	return k.updatedAt
}

// Behaviour helpers.

func (k *KYCProfileEntity) SetVerificationLevel(level VerificationLevel) error {
	if !isValidVerificationLevel(level) {
		return errKYCLevelInvalid
	}
	k.verificationLevel = level
	return nil
}

func (k *KYCProfileEntity) SetStatus(status KYCStatus) error {
	if !isValidKYCStatus(status) {
		return errKYCStatusInvalid
	}

	if !canTransitionKYCStatus(k.status, status) {
		return errKYCTransitionInvalid
	}

	k.status = status
	return nil
}

func (k *KYCProfileEntity) UpdateLimits(daily, monthly decimal.Decimal) error {
	if daily.IsNegative() || monthly.IsNegative() {
		return errKYCLimitNegative
	}
	k.dailyLimitUSD = daily
	k.monthlyLimitUSD = monthly
	return nil
}

func (k *KYCProfileEntity) MarkSubmitted(at time.Time) {
	t := normaliseTimestamp(at)
	k.submittedAt = &t
	_ = k.SetStatus(KYCStatusPending)
	k.Touch(t)
}

func (k *KYCProfileEntity) MarkReviewed(at time.Time) {
	t := normaliseTimestamp(at)
	k.reviewedAt = &t
	if k.status == KYCStatusPending {
		_ = k.SetStatus(KYCStatusUnderReview)
	}
	k.Touch(t)
}

func (k *KYCProfileEntity) MarkApproved(at time.Time) {
	t := normaliseTimestamp(at)
	k.approvedAt = &t
	_ = k.SetStatus(KYCStatusApproved)
	k.Touch(t)
}

func (k *KYCProfileEntity) MarkExpired(at time.Time) {
	t := normaliseTimestamp(at)
	k.expiresAt = &t
	_ = k.SetStatus(KYCStatusExpired)
	k.Touch(t)
}

func (k *KYCProfileEntity) Reject(reason string, notes string) {
	k.rejectionReason = strings.TrimSpace(reason)
	k.reviewerNotes = strings.TrimSpace(notes)
	_ = k.SetStatus(KYCStatusRejected)
	k.Touch(time.Now().UTC())
}

// SetReviewerNotes updates internal reviewer notes without altering state.
func (k *KYCProfileEntity) SetReviewerNotes(notes string) {
	k.reviewerNotes = strings.TrimSpace(notes)
	k.Touch(time.Now().UTC())
}

func (k *KYCProfileEntity) UpdatePII(firstName, lastName, dob, nationality, docNumber, address string) {
	k.firstNameEncrypted = strings.TrimSpace(firstName)
	k.lastNameEncrypted = strings.TrimSpace(lastName)
	k.dateOfBirthEncrypted = strings.TrimSpace(dob)
	k.nationalityEncrypted = strings.TrimSpace(nationality)
	k.documentNumberEncrypted = strings.TrimSpace(docNumber)
	k.addressEncrypted = strings.TrimSpace(address)
	k.Touch(time.Now().UTC())
}

func (k *KYCProfileEntity) Touch(at time.Time) {
	k.updatedAt = normaliseTimestamp(at)
}

func isValidVerificationLevel(level VerificationLevel) bool {
	switch level {
	case VerificationLevelUnverified, VerificationLevelBasic, VerificationLevelFull:
		return true
	default:
		return false
	}
}

func isValidKYCStatus(status KYCStatus) bool {
	switch status {
	case KYCStatusNotStarted,
		KYCStatusPending,
		KYCStatusUnderReview,
		KYCStatusApproved,
		KYCStatusRejected,
		KYCStatusExpired:
		return true
	default:
		return false
	}
}

func canTransitionKYCStatus(current, next KYCStatus) bool {
	switch current {
	case KYCStatusNotStarted:
		return next == KYCStatusPending
	case KYCStatusPending:
		return next == KYCStatusUnderReview || next == KYCStatusRejected
	case KYCStatusUnderReview:
		return next == KYCStatusApproved || next == KYCStatusRejected
	case KYCStatusApproved:
		return next == KYCStatusExpired
	case KYCStatusRejected:
		return next == KYCStatusPending
	case KYCStatusExpired:
		return next == KYCStatusPending
	default:
		return false
	}
}

func normaliseTimestamp(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}
