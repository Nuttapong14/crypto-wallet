package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	errExchangeRateSymbolRequired       = errors.New("exchange rate symbol is required")
	errExchangeRatePriceInvalid         = errors.New("exchange rate price must be positive")
	errExchangeRateChange24hInvalid     = errors.New("exchange rate 24h change is invalid")
	errExchangeRateVolume24hInvalid     = errors.New("exchange rate 24h volume must be non-negative")
	errExchangeRateMarketCapInvalid     = errors.New("exchange rate market cap must be non-negative")
	errExchangeRateLastUpdatedInvalid   = errors.New("exchange rate last updated is required")
)

// ExchangeRate exposes the behavior required by the application layer when working with exchange rate entities.
type ExchangeRate interface {
	Entity
	Identifiable
	Timestamped

	GetSymbol() string
	GetPriceUSD() decimal.Decimal
	GetPriceChange24h() decimal.Decimal
	GetVolume24h() decimal.Decimal
	GetMarketCap() decimal.Decimal
	GetLastUpdated() time.Time
	UpdatePrice(priceUSD, priceChange24h, volume24h, marketCap decimal.Decimal, lastUpdated time.Time) error
	Touch(at time.Time)
}

// ExchangeRateEntity is the default implementation of the ExchangeRate interface.
type ExchangeRateEntity struct {
	id              uuid.UUID
	symbol          string
	priceUSD        decimal.Decimal
	priceChange24h  decimal.Decimal
	volume24h       decimal.Decimal
	marketCap       decimal.Decimal
	lastUpdated     time.Time
	createdAt       time.Time
	updatedAt       time.Time
}

// ExchangeRateParams captures the fields required to construct an ExchangeRateEntity.
type ExchangeRateParams struct {
	ID              uuid.UUID
	Symbol          string
	PriceUSD        decimal.Decimal
	PriceChange24h  decimal.Decimal
	Volume24h       decimal.Decimal
	MarketCap       decimal.Decimal
	LastUpdated     time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewExchangeRateEntity validates the supplied parameters and returns a new ExchangeRateEntity instance.
func NewExchangeRateEntity(params ExchangeRateParams) (*ExchangeRateEntity, error) {
	if params.ID == uuid.Nil {
		params.ID = uuid.New()
	}

	if params.CreatedAt.IsZero() {
		params.CreatedAt = time.Now().UTC()
	}

	if params.UpdatedAt.IsZero() {
		params.UpdatedAt = params.CreatedAt
	}

	if params.LastUpdated.IsZero() {
		params.LastUpdated = time.Now().UTC()
	}

	entity := &ExchangeRateEntity{
		id:             params.ID,
		symbol:         strings.ToUpper(strings.TrimSpace(params.Symbol)),
		priceUSD:       params.PriceUSD,
		priceChange24h: params.PriceChange24h,
		volume24h:      params.Volume24h,
		marketCap:      params.MarketCap,
		lastUpdated:    params.LastUpdated,
		createdAt:      params.CreatedAt,
		updatedAt:      params.UpdatedAt,
	}

	if err := entity.Validate(); err != nil {
		return nil, err
	}

	return entity, nil
}

// HydrateExchangeRateEntity creates an ExchangeRateEntity without re-validating invariants (used for repository hydration).
func HydrateExchangeRateEntity(params ExchangeRateParams) *ExchangeRateEntity {
	return &ExchangeRateEntity{
		id:             params.ID,
		symbol:         strings.ToUpper(strings.TrimSpace(params.Symbol)),
		priceUSD:       params.PriceUSD,
		priceChange24h: params.PriceChange24h,
		volume24h:      params.Volume24h,
		marketCap:      params.MarketCap,
		lastUpdated:    params.LastUpdated,
		createdAt:      params.CreatedAt,
		updatedAt:      params.UpdatedAt,
	}
}

// Validate ensures the entity adheres to domain invariants.
func (e *ExchangeRateEntity) Validate() error {
	var validationErr error

	if strings.TrimSpace(e.symbol) == "" {
		validationErr = errors.Join(validationErr, errExchangeRateSymbolRequired)
	}

	if e.priceUSD.IsNegative() || e.priceUSD.IsZero() {
		validationErr = errors.Join(validationErr, errExchangeRatePriceInvalid)
	}

	if e.volume24h.IsNegative() {
		validationErr = errors.Join(validationErr, errExchangeRateVolume24hInvalid)
	}

	if e.marketCap.IsNegative() {
		validationErr = errors.Join(validationErr, errExchangeRateMarketCapInvalid)
	}

	if e.lastUpdated.IsZero() {
		validationErr = errors.Join(validationErr, errExchangeRateLastUpdatedInvalid)
	}

	return validationErr
}

// Getter implementations satisfy the ExchangeRate interface.

func (e *ExchangeRateEntity) GetID() uuid.UUID {
	return e.id
}

func (e *ExchangeRateEntity) GetSymbol() string {
	return e.symbol
}

func (e *ExchangeRateEntity) GetPriceUSD() decimal.Decimal {
	return e.priceUSD
}

func (e *ExchangeRateEntity) GetPriceChange24h() decimal.Decimal {
	return e.priceChange24h
}

func (e *ExchangeRateEntity) GetVolume24h() decimal.Decimal {
	return e.volume24h
}

func (e *ExchangeRateEntity) GetMarketCap() decimal.Decimal {
	return e.marketCap
}

func (e *ExchangeRateEntity) GetLastUpdated() time.Time {
	return e.lastUpdated
}

func (e *ExchangeRateEntity) GetCreatedAt() time.Time {
	return e.createdAt
}

func (e *ExchangeRateEntity) GetUpdatedAt() time.Time {
	return e.updatedAt
}

// Domain behavior helpers.

// UpdatePrice sets the current price and related metrics.
func (e *ExchangeRateEntity) UpdatePrice(priceUSD, priceChange24h, volume24h, marketCap decimal.Decimal, lastUpdated time.Time) error {
	if priceUSD.IsNegative() || priceUSD.IsZero() {
		return errExchangeRatePriceInvalid
	}

	if volume24h.IsNegative() {
		return errExchangeRateVolume24hInvalid
	}

	if marketCap.IsNegative() {
		return errExchangeRateMarketCapInvalid
	}

	e.priceUSD = priceUSD
	e.priceChange24h = priceChange24h
	e.volume24h = volume24h
	e.marketCap = marketCap

	if lastUpdated.IsZero() {
		e.lastUpdated = time.Now().UTC()
	} else {
		e.lastUpdated = lastUpdated.UTC()
	}

	e.Touch(time.Time{})
	return nil
}

// Touch refreshes the updatedAt timestamp.
func (e *ExchangeRateEntity) Touch(at time.Time) {
	if at.IsZero() {
		e.updatedAt = time.Now().UTC()
		return
	}
	e.updatedAt = at
}

// IsSupportedSymbol reports whether the supplied symbol is one of the supported cryptocurrencies.
func IsSupportedSymbol(symbol string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(symbol))
	switch normalized {
	case "BTC", "ETH", "SOL", "XLM":
		return true
	default:
		return false
	}
}

// SupportedSymbols returns a slice of all supported cryptocurrency symbols.
func SupportedSymbols() []string {
	return []string{"BTC", "ETH", "SOL", "XLM"}
}
