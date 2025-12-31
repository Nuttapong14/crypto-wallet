package security

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultBCryptCost defines the default cost factor used for password hashing.
	DefaultBCryptCost = 12
	// MaxBCryptCost guards against pathological inputs that would hang the process.
	MaxBCryptCost = 18
)

var (
	// ErrPasswordTooShort enforces strong password requirements.
	ErrPasswordTooShort = errors.New("security: password must be at least 12 characters")
	// ErrPasswordHashEmpty ensures stored hashes are not empty.
	ErrPasswordHashEmpty = errors.New("security: password hash is required")
)

// PasswordHasher provides password hashing and verification utilities.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash string, password string) error
	NeedsRehash(hash string) (bool, error)
}

// BcryptHasher implements PasswordHasher using bcrypt.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher constructs a BcryptHasher with the desired cost (defaults applied when <= 0).
func NewBcryptHasher(cost int) (*BcryptHasher, error) {
	if cost <= 0 {
		cost = DefaultBCryptCost
	}
	if cost < bcrypt.MinCost || cost > MaxBCryptCost {
		return nil, fmt.Errorf("security: bcrypt cost must be between %d and %d", bcrypt.MinCost, MaxBCryptCost)
	}
	return &BcryptHasher{cost: cost}, nil
}

// Hash generates a bcrypt hash for the supplied password.
func (h *BcryptHasher) Hash(password string) (string, error) {
	if h == nil {
		return "", errors.New("security: hasher not initialised")
	}
	password = strings.TrimSpace(password)
	if len(password) < 12 {
		return "", ErrPasswordTooShort
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("security: generate hash: %w", err)
	}
	return string(hash), nil
}

// Compare verifies a bcrypt hash against the supplied password.
func (h *BcryptHasher) Compare(hash string, password string) error {
	if h == nil {
		return errors.New("security: hasher not initialised")
	}
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return ErrPasswordHashEmpty
	}
	password = strings.TrimSpace(password)
	if password == "" {
		return ErrPasswordTooShort
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("security: password mismatch: %w", err)
	}
	return nil
}

// NeedsRehash indicates whether the stored hash should be regenerated with the current cost.
func (h *BcryptHasher) NeedsRehash(hash string) (bool, error) {
	if h == nil {
		return false, errors.New("security: hasher not initialised")
	}
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return false, ErrPasswordHashEmpty
	}

	currentCost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return false, fmt.Errorf("security: parse hash cost: %w", err)
	}
	return currentCost != h.cost, nil
}
