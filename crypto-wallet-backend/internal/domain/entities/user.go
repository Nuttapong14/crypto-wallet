package entities

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

// UserStatus represents the lifecycle state of a platform user.
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
)

// CurrencyCode enumerates supported display currencies for portfolio values.
type CurrencyCode string

const (
	CurrencyUSD CurrencyCode = "USD"
	CurrencyEUR CurrencyCode = "EUR"
	CurrencyTHB CurrencyCode = "THB"
	CurrencyGBP CurrencyCode = "GBP"
	CurrencyJPY CurrencyCode = "JPY"
)

var (
	errUserEmailRequired        = errors.New("user email is required")
	errUserEmailInvalid         = errors.New("user email is invalid")
	errUserPasswordHashRequired = errors.New("user password hash is required")
	errUserStatusInvalid        = errors.New("user status is invalid")
	errUserCurrencyInvalid      = errors.New("user preferred currency is invalid")
	errTwoFactorSecretMissing   = errors.New("two-factor secret must be provided when two-factor is enabled")
)

// Entity defines the base contract for all domain entities.
type Entity interface {
	Validate() error
}

// Identifiable is implemented by entities that expose an ID.
type Identifiable interface {
	GetID() uuid.UUID
}

// Timestamped is implemented by entities that track creation and update times.
type Timestamped interface {
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}

// User exposes the behavior required by the application layer when working with user entities.
type User interface {
	Entity
	Identifiable
	Timestamped

	GetEmail() string
	GetPasswordHash() string
	GetFirstName() string
	GetLastName() string
	GetPhoneNumber() string
	GetStatus() UserStatus
	GetPreferredCurrency() CurrencyCode
	IsTwoFactorEnabled() bool
	GetTwoFactorSecret() string
	IsEmailVerified() bool
	GetEmailVerifiedAt() *time.Time
	GetLastLoginAt() *time.Time
}

// UserEntity is the default implementation of the User interface.
type UserEntity struct {
	id                uuid.UUID
	email             string
	passwordHash      string
	firstName         string
	lastName          string
	phoneNumber       string
	status            UserStatus
	preferredCurrency CurrencyCode
	twoFactorEnabled  bool
	twoFactorSecret   string
	emailVerified     bool
	emailVerifiedAt   *time.Time
	lastLoginAt       *time.Time
	createdAt         time.Time
	updatedAt         time.Time
}

// UserParams captures the fields required to construct a UserEntity.
type UserParams struct {
	ID                uuid.UUID
	Email             string
	PasswordHash      string
	FirstName         string
	LastName          string
	PhoneNumber       string
	Status            UserStatus
	PreferredCurrency CurrencyCode
	TwoFactorEnabled  bool
	TwoFactorSecret   string
	EmailVerified     bool
	EmailVerifiedAt   *time.Time
	LastLoginAt       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewUserEntity validates the supplied parameters and returns a new UserEntity instance.
func NewUserEntity(params UserParams) (*UserEntity, error) {
	if params.ID == uuid.Nil {
		params.ID = uuid.New()
	}
	if params.CreatedAt.IsZero() {
		params.CreatedAt = time.Now().UTC()
	}
	if params.UpdatedAt.IsZero() {
		params.UpdatedAt = params.CreatedAt
	}
	if params.Status == "" {
		params.Status = UserStatusActive
	}
	if params.PreferredCurrency == "" {
		params.PreferredCurrency = CurrencyUSD
	}

	entity := &UserEntity{
		id:                params.ID,
		email:             strings.TrimSpace(params.Email),
		passwordHash:      strings.TrimSpace(params.PasswordHash),
		firstName:         strings.TrimSpace(params.FirstName),
		lastName:          strings.TrimSpace(params.LastName),
		phoneNumber:       strings.TrimSpace(params.PhoneNumber),
		status:            params.Status,
		preferredCurrency: params.PreferredCurrency,
		twoFactorEnabled:  params.TwoFactorEnabled,
		twoFactorSecret:   strings.TrimSpace(params.TwoFactorSecret),
		emailVerified:     params.EmailVerified,
		emailVerifiedAt:   params.EmailVerifiedAt,
		lastLoginAt:       params.LastLoginAt,
		createdAt:         params.CreatedAt,
		updatedAt:         params.UpdatedAt,
	}

	if err := entity.Validate(); err != nil {
		return nil, err
	}

	return entity, nil
}

// HydrateUserEntity creates a UserEntity without re-validating invariants (used for repository hydration).
func HydrateUserEntity(params UserParams) *UserEntity {
	return &UserEntity{
		id:                params.ID,
		email:             strings.TrimSpace(params.Email),
		passwordHash:      strings.TrimSpace(params.PasswordHash),
		firstName:         strings.TrimSpace(params.FirstName),
		lastName:          strings.TrimSpace(params.LastName),
		phoneNumber:       strings.TrimSpace(params.PhoneNumber),
		status:            params.Status,
		preferredCurrency: params.PreferredCurrency,
		twoFactorEnabled:  params.TwoFactorEnabled,
		twoFactorSecret:   strings.TrimSpace(params.TwoFactorSecret),
		emailVerified:     params.EmailVerified,
		emailVerifiedAt:   params.EmailVerifiedAt,
		lastLoginAt:       params.LastLoginAt,
		createdAt:         params.CreatedAt,
		updatedAt:         params.UpdatedAt,
	}
}

// Validate ensures the entity adheres to domain invariants.
func (u *UserEntity) Validate() error {
	var validationErr error

	if strings.TrimSpace(u.email) == "" {
		validationErr = errors.Join(validationErr, errUserEmailRequired)
	} else if _, err := mail.ParseAddress(u.email); err != nil {
		validationErr = errors.Join(validationErr, errUserEmailInvalid)
	}

	if u.passwordHash == "" {
		validationErr = errors.Join(validationErr, errUserPasswordHashRequired)
	}

	if !isValidUserStatus(u.status) {
		validationErr = errors.Join(validationErr, errUserStatusInvalid)
	}

	if !isValidCurrencyCode(u.preferredCurrency) {
		validationErr = errors.Join(validationErr, errUserCurrencyInvalid)
	}

	if u.twoFactorEnabled && u.twoFactorSecret == "" {
		validationErr = errors.Join(validationErr, errTwoFactorSecretMissing)
	}

	return validationErr
}

// Getter implementations satisfy the User interface.

func (u *UserEntity) GetID() uuid.UUID {
	return u.id
}

func (u *UserEntity) GetEmail() string {
	return u.email
}

func (u *UserEntity) GetPasswordHash() string {
	return u.passwordHash
}

func (u *UserEntity) GetFirstName() string {
	return u.firstName
}

func (u *UserEntity) GetLastName() string {
	return u.lastName
}

func (u *UserEntity) GetPhoneNumber() string {
	return u.phoneNumber
}

func (u *UserEntity) GetStatus() UserStatus {
	return u.status
}

func (u *UserEntity) GetPreferredCurrency() CurrencyCode {
	return u.preferredCurrency
}

func (u *UserEntity) IsTwoFactorEnabled() bool {
	return u.twoFactorEnabled
}

func (u *UserEntity) GetTwoFactorSecret() string {
	return u.twoFactorSecret
}

func (u *UserEntity) IsEmailVerified() bool {
	return u.emailVerified
}

func (u *UserEntity) GetEmailVerifiedAt() *time.Time {
	return u.emailVerifiedAt
}

func (u *UserEntity) GetLastLoginAt() *time.Time {
	return u.lastLoginAt
}

func (u *UserEntity) GetCreatedAt() time.Time {
	return u.createdAt
}

func (u *UserEntity) GetUpdatedAt() time.Time {
	return u.updatedAt
}

// Domain behavior helpers.

// SetStatus transitions the user to a new status when valid.
func (u *UserEntity) SetStatus(status UserStatus) error {
	if !isValidUserStatus(status) {
		return errUserStatusInvalid
	}
	u.status = status
	return nil
}

// SetPreferredCurrency updates the preferred currency when supported.
func (u *UserEntity) SetPreferredCurrency(code CurrencyCode) error {
	if !isValidCurrencyCode(code) {
		return errUserCurrencyInvalid
	}
	u.preferredCurrency = code
	return nil
}

// EnableTwoFactor toggles two-factor authentication with the provided secret.
func (u *UserEntity) EnableTwoFactor(secret string) error {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		if strings.TrimSpace(u.twoFactorSecret) == "" {
			return errTwoFactorSecretMissing
		}
		secret = u.twoFactorSecret
	}
	u.twoFactorEnabled = true
	u.twoFactorSecret = secret
	return nil
}

// DisableTwoFactor disables two-factor authentication and clears the secret.
func (u *UserEntity) DisableTwoFactor() {
	u.twoFactorEnabled = false
	u.twoFactorSecret = ""
}

// SetTwoFactorSecret stores the shared secret without enabling the feature.
func (u *UserEntity) SetTwoFactorSecret(secret string) error {
	if strings.TrimSpace(secret) == "" {
		return errTwoFactorSecretMissing
	}
	u.twoFactorSecret = strings.TrimSpace(secret)
	return nil
}

// MarkEmailVerified marks the user's email address as verified.
func (u *UserEntity) MarkEmailVerified(at time.Time) {
	u.emailVerified = true
	t := at
	if t.IsZero() {
		t = time.Now().UTC()
	}
	u.emailVerifiedAt = &t
}

// MarkEmailUnverified clears email verification flags.
func (u *UserEntity) MarkEmailUnverified() {
	u.emailVerified = false
	u.emailVerifiedAt = nil
}

// UpdateLastLogin records the timestamp of the user's last successful login.
func (u *UserEntity) UpdateLastLogin(at time.Time) {
	t := at
	if t.IsZero() {
		t = time.Now().UTC()
	}
	u.lastLoginAt = &t
}

// Touch refreshes the updatedAt timestamp.
func (u *UserEntity) Touch(at time.Time) {
	if at.IsZero() {
		u.updatedAt = time.Now().UTC()
		return
	}
	u.updatedAt = at
}

func isValidUserStatus(status UserStatus) bool {
	switch status {
	case UserStatusActive, UserStatusSuspended, UserStatusDeleted:
		return true
	default:
		return false
	}
}

func isValidCurrencyCode(code CurrencyCode) bool {
	switch code {
	case CurrencyUSD, CurrencyEUR, CurrencyTHB, CurrencyGBP, CurrencyJPY:
		return true
	default:
		return false
	}
}
