package entities

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Chain identifies the supported blockchain networks across the platform.
type Chain string

const (
	ChainBTC Chain = "BTC"
	ChainETH Chain = "ETH"
	ChainSOL Chain = "SOL"
	ChainXLM Chain = "XLM"
)

// ChainNetworkType distinguishes between different blockchain environments.
type ChainNetworkType string

const (
	ChainNetworkMainnet ChainNetworkType = "mainnet"
	ChainNetworkTestnet ChainNetworkType = "testnet"
)

var (
	errChainSymbolRequired          = errors.New("chain symbol is required")
	errChainSymbolUnsupported       = errors.New("chain symbol is not supported")
	errChainNameRequired            = errors.New("chain name is required")
	errChainRPCURLRequired          = errors.New("chain RPC URL is required")
	errChainExplorerURLRequired     = errors.New("chain explorer URL is required")
	errChainNativeTokenSymbol       = errors.New("chain native token symbol is required")
	errChainNativeTokenDecimals     = errors.New("chain native token decimals must be between 0 and 18")
	errChainConfirmationThreshold   = errors.New("chain confirmation threshold must be greater than zero")
	errChainAverageBlockTimeInvalid = errors.New("chain average block time must be greater than or equal to zero")
	errChainNetworkTypeInvalid      = errors.New("chain network type is invalid")
)

// ChainConfig exposes the behaviour required when working with blockchain configuration entities.
type ChainConfig interface {
	Entity
	GetID() int64
	GetSymbol() Chain
	GetName() string
	GetRPCURL() string
	GetRPCFallbackURL() string
	GetExplorerURL() string
	GetNativeTokenSymbol() string
	GetNativeTokenDecimals() int
	GetConfirmationThreshold() int
	GetAverageBlockTime() time.Duration
	IsActive() bool
	GetNetworkType() ChainNetworkType
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}

// ChainEntity is the default implementation of ChainConfig.
type ChainEntity struct {
	id                    int64
	symbol                Chain
	name                  string
	rpcURL                string
	rpcURLFallback        string
	explorerURL           string
	nativeTokenSymbol     string
	nativeTokenDecimals   int
	confirmationThreshold int
	averageBlockTime      time.Duration
	isActive              bool
	networkType           ChainNetworkType
	createdAt             time.Time
	updatedAt             time.Time
}

// ChainParams captures the fields required to construct a ChainEntity.
type ChainParams struct {
	ID                    int64
	Symbol                Chain
	Name                  string
	RPCURL                string
	RPCURLFallback        string
	ExplorerURL           string
	NativeTokenSymbol     string
	NativeTokenDecimals   int
	ConfirmationThreshold int
	AverageBlockTime      time.Duration
	IsActive              bool
	NetworkType           ChainNetworkType
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// NewChainEntity validates the supplied parameters and returns a new ChainEntity instance.
func NewChainEntity(params ChainParams) (*ChainEntity, error) {
	if params.Symbol == "" {
		return nil, errChainSymbolRequired
	}
	if !IsSupportedChain(params.Symbol) {
		return nil, fmt.Errorf("%w: %s", errChainSymbolUnsupported, params.Symbol)
	}

	if params.CreatedAt.IsZero() {
		params.CreatedAt = time.Now().UTC()
	}
	if params.UpdatedAt.IsZero() {
		params.UpdatedAt = params.CreatedAt
	}
	if params.NativeTokenDecimals == 0 {
		params.NativeTokenDecimals = 18
	}
	if params.ConfirmationThreshold == 0 {
		params.ConfirmationThreshold = 1
	}
	if params.NetworkType == "" {
		params.NetworkType = ChainNetworkMainnet
	}

	entity := &ChainEntity{
		id:                    params.ID,
		symbol:                params.Symbol,
		name:                  strings.TrimSpace(params.Name),
		rpcURL:                strings.TrimSpace(params.RPCURL),
		rpcURLFallback:        strings.TrimSpace(params.RPCURLFallback),
		explorerURL:           strings.TrimSpace(params.ExplorerURL),
		nativeTokenSymbol:     strings.ToUpper(strings.TrimSpace(params.NativeTokenSymbol)),
		nativeTokenDecimals:   params.NativeTokenDecimals,
		confirmationThreshold: params.ConfirmationThreshold,
		averageBlockTime:      params.AverageBlockTime,
		isActive:              params.IsActive,
		networkType:           params.NetworkType,
		createdAt:             params.CreatedAt,
		updatedAt:             params.UpdatedAt,
	}

	if err := entity.Validate(); err != nil {
		return nil, err
	}

	return entity, nil
}

// HydrateChainEntity builds a ChainEntity without executing validation (used when hydrating from persistence).
func HydrateChainEntity(params ChainParams) *ChainEntity {
	return &ChainEntity{
		id:                    params.ID,
		symbol:                params.Symbol,
		name:                  strings.TrimSpace(params.Name),
		rpcURL:                strings.TrimSpace(params.RPCURL),
		rpcURLFallback:        strings.TrimSpace(params.RPCURLFallback),
		explorerURL:           strings.TrimSpace(params.ExplorerURL),
		nativeTokenSymbol:     strings.ToUpper(strings.TrimSpace(params.NativeTokenSymbol)),
		nativeTokenDecimals:   params.NativeTokenDecimals,
		confirmationThreshold: params.ConfirmationThreshold,
		averageBlockTime:      params.AverageBlockTime,
		isActive:              params.IsActive,
		networkType:           params.NetworkType,
		createdAt:             params.CreatedAt,
		updatedAt:             params.UpdatedAt,
	}
}

// Validate ensures the chain entity adheres to domain invariants.
func (c *ChainEntity) Validate() error {
	var validationErr error

	if c.symbol == "" {
		validationErr = errors.Join(validationErr, errChainSymbolRequired)
	} else if !IsSupportedChain(c.symbol) {
		validationErr = errors.Join(validationErr, fmt.Errorf("%w: %s", errChainSymbolUnsupported, c.symbol))
	}

	if strings.TrimSpace(c.name) == "" {
		validationErr = errors.Join(validationErr, errChainNameRequired)
	}
	if strings.TrimSpace(c.rpcURL) == "" {
		validationErr = errors.Join(validationErr, errChainRPCURLRequired)
	}
	if strings.TrimSpace(c.explorerURL) == "" {
		validationErr = errors.Join(validationErr, errChainExplorerURLRequired)
	}
	if strings.TrimSpace(c.nativeTokenSymbol) == "" {
		validationErr = errors.Join(validationErr, errChainNativeTokenSymbol)
	}
	if c.nativeTokenDecimals < 0 || c.nativeTokenDecimals > 18 {
		validationErr = errors.Join(validationErr, errChainNativeTokenDecimals)
	}
	if c.confirmationThreshold <= 0 {
		validationErr = errors.Join(validationErr, errChainConfirmationThreshold)
	}
	if c.averageBlockTime < 0 {
		validationErr = errors.Join(validationErr, errChainAverageBlockTimeInvalid)
	}
	if c.networkType != ChainNetworkMainnet && c.networkType != ChainNetworkTestnet {
		validationErr = errors.Join(validationErr, errChainNetworkTypeInvalid)
	}

	return validationErr
}

// GetID returns the chain identifier.
func (c *ChainEntity) GetID() int64 {
	return c.id
}

// GetSymbol returns the chain symbol (BTC/ETH/...).
func (c *ChainEntity) GetSymbol() Chain {
	return c.symbol
}

// GetName returns the human readable name for the chain.
func (c *ChainEntity) GetName() string {
	return c.name
}

// GetRPCURL returns the primary RPC endpoint for the chain.
func (c *ChainEntity) GetRPCURL() string {
	return c.rpcURL
}

// GetRPCFallbackURL returns the fallback RPC endpoint for the chain.
func (c *ChainEntity) GetRPCFallbackURL() string {
	return c.rpcURLFallback
}

// GetExplorerURL returns the block explorer base URL.
func (c *ChainEntity) GetExplorerURL() string {
	return c.explorerURL
}

// GetNativeTokenSymbol returns the chain native token symbol (e.g., BTC).
func (c *ChainEntity) GetNativeTokenSymbol() string {
	return c.nativeTokenSymbol
}

// GetNativeTokenDecimals returns the decimal precision for the native token.
func (c *ChainEntity) GetNativeTokenDecimals() int {
	return c.nativeTokenDecimals
}

// GetConfirmationThreshold returns the required confirmations for finality.
func (c *ChainEntity) GetConfirmationThreshold() int {
	return c.confirmationThreshold
}

// GetAverageBlockTime returns the average block time for the chain.
func (c *ChainEntity) GetAverageBlockTime() time.Duration {
	return c.averageBlockTime
}

// IsActive indicates whether the chain is enabled.
func (c *ChainEntity) IsActive() bool {
	return c.isActive
}

// GetNetworkType returns the configured network type (mainnet/testnet).
func (c *ChainEntity) GetNetworkType() ChainNetworkType {
	return c.networkType
}

// GetCreatedAt returns the creation timestamp.
func (c *ChainEntity) GetCreatedAt() time.Time {
	return c.createdAt
}

// GetUpdatedAt returns the last update timestamp.
func (c *ChainEntity) GetUpdatedAt() time.Time {
	return c.updatedAt
}

// SetActive toggles the active status of the chain.
func (c *ChainEntity) SetActive(active bool) {
	c.isActive = active
}

// Touch updates the modification timestamp.
func (c *ChainEntity) Touch(at time.Time) {
	if at.IsZero() {
		c.updatedAt = time.Now().UTC()
		return
	}
	c.updatedAt = at.UTC()
}

// IsSupportedChain reports whether the supplied chain symbol is one of the supported networks.
func IsSupportedChain(chain Chain) bool {
	switch chain {
	case ChainBTC, ChainETH, ChainSOL, ChainXLM:
		return true
	default:
		return false
	}
}

// NormalizeChain converts a user supplied chain string into the canonical Chain type.
func NormalizeChain(value string) Chain {
	trimmed := strings.ToUpper(strings.TrimSpace(value))
	switch Chain(trimmed) {
	case ChainBTC, ChainETH, ChainSOL, ChainXLM:
		return Chain(trimmed)
	default:
		return Chain("")
	}
}

// SupportedChains returns a slice of all supported chain codes.
func SupportedChains() []Chain {
	return []Chain{ChainBTC, ChainETH, ChainSOL, ChainXLM}
}
