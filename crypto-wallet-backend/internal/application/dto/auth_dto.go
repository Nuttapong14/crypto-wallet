package dto

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/pkg/utils"
)

type RegisterRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	PhoneNumber     string `json:"phoneNumber"`
}

type LoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
}

type LogoutRequest struct {
	UserID uuid.UUID `json:"userId"`
}

type AuthTokens struct {
	AccessToken      string    `json:"accessToken"`
	RefreshToken     string    `json:"refreshToken,omitempty"`
	ExpiresAt        time.Time `json:"expiresAt"`
	RefreshExpiresAt time.Time `json:"refreshExpiresAt,omitempty"`
}

type AuthUser struct {
	ID                uuid.UUID `json:"id"`
	Email             string    `json:"email"`
	FirstName         string    `json:"firstName,omitempty"`
	LastName          string    `json:"lastName,omitempty"`
	PhoneNumber       string    `json:"phoneNumber,omitempty"`
	Status            string    `json:"status"`
	PreferredCurrency string    `json:"preferredCurrency"`
	TwoFactorEnabled  bool      `json:"twoFactorEnabled"`
	EmailVerified     bool      `json:"emailVerified"`
	LastLoginAt       *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type AuthResponse struct {
	User   AuthUser   `json:"user"`
	Tokens AuthTokens `json:"tokens"`
}

func (r RegisterRequest) Validate() utils.ValidationErrors {
	errs := utils.ValidationErrors{}
	utils.RequireEmail(&errs, "email", r.Email)
	utils.RequireMinLength(&errs, "password", r.Password, 12)
	if strings.TrimSpace(r.ConfirmPassword) == "" {
		errs.Add("confirmPassword", "is required")
	} else if r.Password != r.ConfirmPassword {
		errs.Add("confirmPassword", "does not match password")
	}

	if !errs.IsEmpty() {
		return errs
	}

	return errs
}

func (r LoginRequest) Validate() utils.ValidationErrors {
	errs := utils.ValidationErrors{}
	utils.RequireEmail(&errs, "email", r.Email)
	utils.Require(&errs, "password", r.Password)
	return errs
}

func NewAuthUser(user entities.User) AuthUser {
	return AuthUser{
		ID:                user.GetID(),
		Email:             user.GetEmail(),
		FirstName:         user.GetFirstName(),
		LastName:          user.GetLastName(),
		PhoneNumber:       user.GetPhoneNumber(),
		Status:            string(user.GetStatus()),
		PreferredCurrency: string(user.GetPreferredCurrency()),
		TwoFactorEnabled:  user.IsTwoFactorEnabled(),
		EmailVerified:     user.IsEmailVerified(),
		LastLoginAt:       user.GetLastLoginAt(),
		CreatedAt:         user.GetCreatedAt(),
		UpdatedAt:         user.GetUpdatedAt(),
	}
}

// TwoFactorSetupResponse describes the result of initiating 2FA setup.
type TwoFactorSetupResponse struct {
	Secret     string `json:"secret"`
	OtpauthURL string `json:"otpauthUrl"`
}

// EnableTwoFactorRequest carries the verification code provided by the user.
type EnableTwoFactorRequest struct {
    Code string `json:"code"`
}

func (r EnableTwoFactorRequest) Validate() utils.ValidationErrors {
    errs := utils.ValidationErrors{}
    if len(strings.TrimSpace(r.Code)) != 6 {
        errs.Add("code", "must be a 6-digit verification code")
    }
    return errs
}

// DisableTwoFactorRequest optionally includes a verification code for confirmation.
type DisableTwoFactorRequest struct {
    Code string `json:"code,omitempty"`
}

func (r DisableTwoFactorRequest) Validate() utils.ValidationErrors {
    errs := utils.ValidationErrors{}
    if strings.TrimSpace(r.Code) != "" && len(strings.TrimSpace(r.Code)) != 6 {
        errs.Add("code", "must be a 6-digit verification code")
    }
    return errs
}

// TwoFactorStatusResponse reports the updated 2FA state for the account.
type TwoFactorStatusResponse struct {
	Enabled bool `json:"enabled"`
}
