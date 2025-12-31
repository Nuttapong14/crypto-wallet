package dto

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// KYCAddress represents the address structure submitted during verification.
type KYCAddress struct {
	Street     string `json:"street,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postalCode,omitempty"`
	Country    string `json:"country,omitempty"`
}

// KYCSubmitRequest matches the OpenAPI contract for initiating verification.
type KYCSubmitRequest struct {
	FirstName   string     `json:"firstName"`
	LastName    string     `json:"lastName"`
	DateOfBirth string     `json:"dateOfBirth"` // ISO8601 date (YYYY-MM-DD)
	Nationality string     `json:"nationality"`
	Address     KYCAddress `json:"address"`
}

// Validate enforces request invariants.
func (r KYCSubmitRequest) Validate() utils.ValidationErrors {
	errs := utils.ValidationErrors{}
	utils.Require(&errs, "firstName", r.FirstName)
	utils.Require(&errs, "lastName", r.LastName)
	utils.RequirePattern(&errs, "dateOfBirth", r.DateOfBirth, `^\d{4}-\d{2}-\d{2}$`, "must be YYYY-MM-DD")
	if len(strings.TrimSpace(r.Nationality)) != 2 {
		errs.Add("nationality", "must be a 2-letter country code")
	}
	return errs
}

// KYCProfile represents the API response describing verification progress.
type KYCProfile struct {
	ID               uuid.UUID `json:"id"`
	VerificationLevel string    `json:"verificationLevel"`
	Status           string    `json:"status"`
	SubmittedAt      *time.Time `json:"submittedAt,omitempty"`
	ApprovedAt       *time.Time `json:"approvedAt,omitempty"`
	ReviewedAt       *time.Time `json:"reviewedAt,omitempty"`
	ExpiresAt        *time.Time `json:"expiresAt,omitempty"`
	DailyLimitUSD    string    `json:"dailyLimitUsd"`
	MonthlyLimitUSD  string    `json:"monthlyLimitUsd"`
	RejectionReason  string    `json:"rejectionReason,omitempty"`
	ReviewerNotes    string    `json:"reviewerNotes,omitempty"`
}

// KYCDocument represents an uploaded verification document.
type KYCDocument struct {
	ID          uuid.UUID  `json:"id"`
	DocumentType string     `json:"documentType"`
	Status      string     `json:"status"`
	UploadedAt  time.Time  `json:"uploadedAt"`
	ReviewedAt  *time.Time `json:"reviewedAt,omitempty"`
	RejectionReason string `json:"rejectionReason,omitempty"`
	FileName    string     `json:"fileName"`
	MimeType    string     `json:"mimeType"`
	FileSize    int        `json:"fileSize"`
}

// KYCDocumentUploadResponse summarises a document upload operation.
type KYCDocumentUploadResponse struct {
	Document KYCDocument `json:"document"`
	Provider struct {
		DocumentID string `json:"documentId,omitempty"`
		Status     string `json:"status,omitempty"`
	} `json:"provider"`
}

// KYCStatusResponse aggregates profile, documents and risk score.
type KYCStatusResponse struct {
	Profile   KYCProfile    `json:"profile"`
	Documents []KYCDocument `json:"documents"`
	RiskScore *RiskScore    `json:"riskScore,omitempty"`
}

// RiskScore represents AML risk scoring for a user.
type RiskScore struct {
	Score           int        `json:"score"`
	Level           string     `json:"level"`
	RiskFactors     []string   `json:"riskFactors"`
	AmlHits         []string   `json:"amlHits"`
	LastScreeningAt *time.Time `json:"lastScreeningAt,omitempty"`
	NextReviewAt    time.Time  `json:"nextReviewAt"`
}

// MapKYCProfile converts a domain entity into transport representation.
func MapKYCProfile(profile entities.KYCProfile) KYCProfile {
	if profile == nil {
		return KYCProfile{}
	}
	return KYCProfile{
		ID:               profile.GetID(),
		VerificationLevel: string(profile.GetVerificationLevel()),
		Status:           string(profile.GetStatus()),
		SubmittedAt:      profile.GetSubmittedAt(),
		ApprovedAt:       profile.GetApprovedAt(),
		ReviewedAt:       profile.GetReviewedAt(),
		ExpiresAt:        profile.GetExpiresAt(),
		DailyLimitUSD:    formatDecimal(profile.GetDailyLimitUSD()),
		MonthlyLimitUSD:  formatDecimal(profile.GetMonthlyLimitUSD()),
		RejectionReason:  profile.GetRejectionReason(),
		ReviewerNotes:    profile.GetReviewerNotes(),
	}
}

// MapKYCDocument converts domain document entity into DTO.
func MapKYCDocument(document entities.KYCDocument) KYCDocument {
	if document == nil {
		return KYCDocument{}
	}
	return KYCDocument{
		ID:             document.GetID(),
		DocumentType:   string(document.GetDocumentType()),
		Status:         string(document.GetStatus()),
		UploadedAt:     document.GetUploadedAt(),
		ReviewedAt:     document.GetReviewedAt(),
		RejectionReason: document.GetRejectionReason(),
		FileName:       document.GetEncryptedFileName(),
		MimeType:       document.GetMimeType(),
		FileSize:       document.GetFileSize(),
	}
}

// MapRiskScore converts risk score entity.
func MapRiskScore(score entities.UserRiskScore) *RiskScore {
	if score == nil {
		return nil
	}
	return &RiskScore{
		Score:           score.GetScore(),
		Level:           string(score.GetLevel()),
		RiskFactors:     score.GetRiskFactors(),
		AmlHits:         score.GetAMLHits(),
		LastScreeningAt: score.GetLastScreeningAt(),
		NextReviewAt:    score.GetNextReviewAt(),
	}
}

func formatDecimal(value decimal.Decimal) string {
	if value.IsZero() {
		return "0"
	}
	return value.StringFixedBank(2)
}
