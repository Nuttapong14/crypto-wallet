package entities

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	errTokenSymbolRequired       = errors.New("token symbol is required")
	errTokenSymbolFormat         = errors.New("token symbol must be 2-20 uppercase characters")
	errTokenNameRequired         = errors.New("token name is required")
	errTokenChainRequired        = errors.New("token chain identifier is required")
	errTokenSymbolUnsupported    = errors.New("unsupported chain symbol")
	errTokenDecimalsRange        = errors.New("token decimals must be between 0 and 18")
	errTokenContractRequired     = errors.New("token contract address is required for non-native tokens")
	errTokenContractForNative    = errors.New("token contract address must be empty for native tokens")
	errTokenLogoURLTooLong       = errors.New("token logo URL must be at most 500 characters")
	errTokenCoingeckoIDTooLong   = errors.New("token coingecko id must be at most 100 characters")
	errTokenNativeSymbolMismatch = errors.New("token native token symbol must match chain symbol when native")
)

// Token defines the behaviour required when interacting with token entities.
type Token interface {
	Entity
	GetID() int64
	GetSymbol() string
	GetName() string
	GetChainID() int64
	GetChainSymbol() Chain
	GetContractAddress() string
	GetDecimals() int
	IsNative() bool
	GetLogoURL() string
	GetCoingeckoID() string
	IsActive() bool
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetActive(active bool)
	Touch(at time.Time)
}

// TokenEntity is the concrete implementation of Token.
type TokenEntity struct {
	id              int64
	symbol          string
	name            string
	chainID         int64
	chainSymbol     Chain
	contractAddress string
	decimals        int
	isNative        bool
	logoURL         string
	coingeckoID     string
	isActive        bool
	createdAt       time.Time
	updatedAt       time.Time
}

// TokenParams captures the fields required to construct a TokenEntity.
type TokenParams struct {
	ID              int64
	Symbol          string
	Name            string
	ChainID         int64
	ChainSymbol     Chain
	ContractAddress string
	Decimals        int
	IsNative        bool
	LogoURL         string
	CoingeckoID     string
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewTokenEntity validates the supplied parameters and returns a new TokenEntity instance.
func NewTokenEntity(params TokenParams) (*TokenEntity, error) {
	if params.CreatedAt.IsZero() {
		params.CreatedAt = time.Now().UTC()
	}
	if params.UpdatedAt.IsZero() {
		params.UpdatedAt = params.CreatedAt
	}

	entity := &TokenEntity{
		id:              params.ID,
		symbol:          strings.ToUpper(strings.TrimSpace(params.Symbol)),
		name:            strings.TrimSpace(params.Name),
		chainID:         params.ChainID,
		chainSymbol:     params.ChainSymbol,
		contractAddress: strings.TrimSpace(params.ContractAddress),
		decimals:        params.Decimals,
		isNative:        params.IsNative,
		logoURL:         strings.TrimSpace(params.LogoURL),
		coingeckoID:     strings.TrimSpace(params.CoingeckoID),
		isActive:        params.IsActive,
		createdAt:       params.CreatedAt,
		updatedAt:       params.UpdatedAt,
	}

	if err := entity.Validate(); err != nil {
		return nil, err
	}

	return entity, nil
}

// HydrateTokenEntity builds a TokenEntity without executing validation (used for persistence hydration).
func HydrateTokenEntity(params TokenParams) *TokenEntity {
	return &TokenEntity{
		id:              params.ID,
		symbol:          strings.ToUpper(strings.TrimSpace(params.Symbol)),
		name:            strings.TrimSpace(params.Name),
		chainID:         params.ChainID,
		chainSymbol:     params.ChainSymbol,
		contractAddress: strings.TrimSpace(params.ContractAddress),
		decimals:        params.Decimals,
		isNative:        params.IsNative,
		logoURL:         strings.TrimSpace(params.LogoURL),
		coingeckoID:     strings.TrimSpace(params.CoingeckoID),
		isActive:        params.IsActive,
		createdAt:       params.CreatedAt,
		updatedAt:       params.UpdatedAt,
	}
}

// Validate ensures the token entity adheres to domain invariants.
func (t *TokenEntity) Validate() error {
	var validationErr error

	if len(t.symbol) == 0 {
		validationErr = errors.Join(validationErr, errTokenSymbolRequired)
	} else if len(t.symbol) < 2 || len(t.symbol) > 20 || strings.ToUpper(t.symbol) != t.symbol {
		validationErr = errors.Join(validationErr, errTokenSymbolFormat)
	}

	if strings.TrimSpace(t.name) == "" {
		validationErr = errors.Join(validationErr, errTokenNameRequired)
	}

	if t.chainID <= 0 {
		validationErr = errors.Join(validationErr, errTokenChainRequired)
	}

	if t.chainSymbol != "" && !IsSupportedChain(t.chainSymbol) {
		validationErr = errors.Join(validationErr, fmt.Errorf("%w: %s", errTokenSymbolUnsupported, t.chainSymbol))
	}

	if t.decimals < 0 || t.decimals > 18 {
		validationErr = errors.Join(validationErr, errTokenDecimalsRange)
	}

	if t.isNative {
		if t.contractAddress != "" {
			validationErr = errors.Join(validationErr, errTokenContractForNative)
		}
		if t.chainSymbol != "" && t.symbol != string(t.chainSymbol) {
			validationErr = errors.Join(validationErr, errTokenNativeSymbolMismatch)
		}
	} else {
		if strings.TrimSpace(t.contractAddress) == "" {
			validationErr = errors.Join(validationErr, errTokenContractRequired)
		}
	}

	if len(t.logoURL) > 500 {
		validationErr = errors.Join(validationErr, errTokenLogoURLTooLong)
	}
	if len(t.coingeckoID) > 0 && len(t.coingeckoID) > 100 {
		validationErr = errors.Join(validationErr, errTokenCoingeckoIDTooLong)
	}

	return validationErr
}

// GetID returns the token identifier.
func (t *TokenEntity) GetID() int64 {
	return t.id
}

// GetSymbol returns the token symbol.
func (t *TokenEntity) GetSymbol() string {
	return t.symbol
}

// GetName returns the token name.
func (t *TokenEntity) GetName() string {
	return t.name
}

// GetChainID returns the associated chain identifier.
func (t *TokenEntity) GetChainID() int64 {
	return t.chainID
}

// GetChainSymbol returns the associated chain symbol.
func (t *TokenEntity) GetChainSymbol() Chain {
	return t.chainSymbol
}

// GetContractAddress returns the smart contract address (empty for native tokens).
func (t *TokenEntity) GetContractAddress() string {
	return t.contractAddress
}

// GetDecimals returns the decimal precision.
func (t *TokenEntity) GetDecimals() int {
	return t.decimals
}

// IsNative indicates whether the token is the chain native asset.
func (t *TokenEntity) IsNative() bool {
	return t.isNative
}

// GetLogoURL returns an optional logo URL.
func (t *TokenEntity) GetLogoURL() string {
	return t.logoURL
}

// GetCoingeckoID returns the CoinGecko identifier.
func (t *TokenEntity) GetCoingeckoID() string {
	return t.coingeckoID
}

// IsActive indicates whether the token is enabled.
func (t *TokenEntity) IsActive() bool {
	return t.isActive
}

// GetCreatedAt returns the creation timestamp.
func (t *TokenEntity) GetCreatedAt() time.Time {
	return t.createdAt
}

// GetUpdatedAt returns the last update timestamp.
func (t *TokenEntity) GetUpdatedAt() time.Time {
	return t.updatedAt
}

// SetActive toggles the active status.
func (t *TokenEntity) SetActive(active bool) {
	t.isActive = active
}

// Touch updates the modification timestamp.
func (t *TokenEntity) Touch(at time.Time) {
	if at.IsZero() {
		t.updatedAt = time.Now().UTC()
		return
	}
	t.updatedAt = at.UTC()
}
