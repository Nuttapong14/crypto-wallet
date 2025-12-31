package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// RiskLevel represents AML risk categories.
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

var (
	errRiskScoreUserIDRequired = errors.New("risk score: user ID is required")
	errRiskScoreOutOfRange     = errors.New("risk score: value must be between 0 and 100")
	errRiskLevelInvalid        = errors.New("risk score: level is invalid")
	errNextReviewRequired      = errors.New("risk score: next review timestamp is required")
)

// UserRiskScore exposes behaviours required by the application layer when working with risk scores.
type UserRiskScore interface {
	Entity
	Identifiable
	Timestamped

	GetUserID() uuid.UUID
	GetScore() int
	GetLevel() RiskLevel
	GetRiskFactors() []string
	GetAMLHits() []string
	GetLastScreeningAt() *time.Time
	GetNextReviewAt() time.Time

	UpdateScore(score int, level RiskLevel)
	SetRiskFactors(factors []string)
	SetAMLHits(hits []string)
	MarkScreened(at time.Time, nextReview time.Time)
	Touch(at time.Time)
}

// UserRiskScoreEntity is the default implementation of UserRiskScore.
type UserRiskScoreEntity struct {
	id              uuid.UUID
	userID          uuid.UUID
	riskScore       int
	riskLevel       RiskLevel
	riskFactors     []string
	amlHits         []string
	lastScreeningAt *time.Time
	nextReviewAt    time.Time
	createdAt       time.Time
	updatedAt       time.Time
}

// UserRiskScoreParams captures the fields required to construct a UserRiskScoreEntity.
type UserRiskScoreParams struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	RiskScore       int
	RiskLevel       RiskLevel
	RiskFactors     []string
	AMLHits         []string
	LastScreeningAt *time.Time
	NextReviewAt    time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewUserRiskScoreEntity validates the supplied parameters and returns a UserRiskScoreEntity.
func NewUserRiskScoreEntity(params UserRiskScoreParams) (*UserRiskScoreEntity, error) {
	if params.ID == uuid.Nil {
		params.ID = uuid.New()
	}
	if params.RiskLevel == "" {
		params.RiskLevel = RiskLevelLow
	}
	if params.CreatedAt.IsZero() {
		params.CreatedAt = time.Now().UTC()
	}
	if params.UpdatedAt.IsZero() {
		params.UpdatedAt = params.CreatedAt
	}

	entity := &UserRiskScoreEntity{
		id:              params.ID,
		userID:          params.UserID,
		riskScore:       params.RiskScore,
		riskLevel:       params.RiskLevel,
		riskFactors:     cloneStringSlice(params.RiskFactors),
		amlHits:         cloneStringSlice(params.AMLHits),
		lastScreeningAt: params.LastScreeningAt,
		nextReviewAt:    params.NextReviewAt,
		createdAt:       params.CreatedAt,
		updatedAt:       params.UpdatedAt,
	}

	if err := entity.Validate(); err != nil {
		return nil, err
	}

	return entity, nil
}

// HydrateUserRiskScoreEntity constructs an entity without re-validating invariants.
func HydrateUserRiskScoreEntity(params UserRiskScoreParams) *UserRiskScoreEntity {
	return &UserRiskScoreEntity{
		id:              params.ID,
		userID:          params.UserID,
		riskScore:       params.RiskScore,
		riskLevel:       params.RiskLevel,
		riskFactors:     cloneStringSlice(params.RiskFactors),
		amlHits:         cloneStringSlice(params.AMLHits),
		lastScreeningAt: params.LastScreeningAt,
		nextReviewAt:    params.NextReviewAt,
		createdAt:       params.CreatedAt,
		updatedAt:       params.UpdatedAt,
	}
}

// Validate ensures domain invariants.
func (r *UserRiskScoreEntity) Validate() error {
	var validationErr error

	if r.userID == uuid.Nil {
		validationErr = errors.Join(validationErr, errRiskScoreUserIDRequired)
	}
	if r.riskScore < 0 || r.riskScore > 100 {
		validationErr = errors.Join(validationErr, errRiskScoreOutOfRange)
	}
	if !isValidRiskLevel(r.riskLevel) {
		validationErr = errors.Join(validationErr, errRiskLevelInvalid)
	}
	if r.nextReviewAt.IsZero() {
		validationErr = errors.Join(validationErr, errNextReviewRequired)
	}
	return validationErr
}

// Getter implementations.

func (r *UserRiskScoreEntity) GetID() uuid.UUID {
	return r.id
}

func (r *UserRiskScoreEntity) GetUserID() uuid.UUID {
	return r.userID
}

func (r *UserRiskScoreEntity) GetScore() int {
	return r.riskScore
}

func (r *UserRiskScoreEntity) GetLevel() RiskLevel {
	return r.riskLevel
}

func (r *UserRiskScoreEntity) GetRiskFactors() []string {
	return cloneStringSlice(r.riskFactors)
}

func (r *UserRiskScoreEntity) GetAMLHits() []string {
	return cloneStringSlice(r.amlHits)
}

func (r *UserRiskScoreEntity) GetLastScreeningAt() *time.Time {
	return r.lastScreeningAt
}

func (r *UserRiskScoreEntity) GetNextReviewAt() time.Time {
	return r.nextReviewAt
}

func (r *UserRiskScoreEntity) GetCreatedAt() time.Time {
	return r.createdAt
}

func (r *UserRiskScoreEntity) GetUpdatedAt() time.Time {
	return r.updatedAt
}

// Behaviour helpers.

func (r *UserRiskScoreEntity) UpdateScore(score int, level RiskLevel) {
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	if isValidRiskLevel(level) {
		r.riskLevel = level
	}
	r.riskScore = score
	r.Touch(time.Now().UTC())
}

func (r *UserRiskScoreEntity) SetRiskFactors(factors []string) {
	r.riskFactors = cloneStringSlice(factors)
	r.Touch(time.Now().UTC())
}

func (r *UserRiskScoreEntity) SetAMLHits(hits []string) {
	r.amlHits = cloneStringSlice(hits)
	r.Touch(time.Now().UTC())
}

func (r *UserRiskScoreEntity) MarkScreened(at time.Time, nextReview time.Time) {
	t := normaliseTimestamp(at)
	r.lastScreeningAt = &t
	if nextReview.IsZero() {
		nextReview = t.Add(30 * 24 * time.Hour)
	}
	r.nextReviewAt = nextReview.UTC()
	r.Touch(t)
}

func (r *UserRiskScoreEntity) Touch(at time.Time) {
	r.updatedAt = normaliseTimestamp(at)
}

func isValidRiskLevel(level RiskLevel) bool {
	switch level {
	case RiskLevelLow, RiskLevelMedium, RiskLevelHigh, RiskLevelCritical:
		return true
	default:
		return false
	}
}

func cloneStringSlice(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	copied := make([]string, len(values))
	copy(copied, values)
	for i, value := range copied {
		copied[i] = strings.TrimSpace(value)
	}
	return copied
}
