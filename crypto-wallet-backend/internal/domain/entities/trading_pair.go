package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	errTradingPairBaseSymbolRequired  = errors.New("trading pair base symbol is required")
	errTradingPairQuoteSymbolRequired = errors.New("trading pair quote symbol is required")
	errTradingPairSameSymbols         = errors.New("trading pair base and quote symbols cannot be the same")
	errTradingPairRateInvalid         = errors.New("trading pair exchange rate must be positive")
	errTradingPairInverseRateInvalid  = errors.New("trading pair inverse rate must be positive")
	errTradingPairFeeInvalid          = errors.New("trading pair fee percentage cannot be negative")
	errTradingPairMinAmountInvalid    = errors.New("trading pair minimum swap amount cannot be negative")
	errTradingPairMaxAmountInvalid    = errors.New("trading pair maximum swap amount must be greater than minimum")
	errTradingPairDailyVolumeInvalid  = errors.New("trading pair daily volume cannot be negative")
)

// TradingPair exposes the behavior required by the application layer when working with trading pair entities.
type TradingPair interface {
	Entity
	Identifiable
	Timestamped

	GetBaseSymbol() string
	GetQuoteSymbol() string
	GetExchangeRate() decimal.Decimal
	GetInverseRate() decimal.Decimal
	GetFeePercentage() decimal.Decimal
	GetMinSwapAmount() decimal.Decimal
	GetMaxSwapAmount() *decimal.Decimal
	GetDailyVolume() decimal.Decimal
	IsActive() bool
	HasLiquidity() bool
	GetLastUpdated() time.Time
}

// TradingPairEntity is the default implementation of the TradingPair interface.
type TradingPairEntity struct {
	id            uuid.UUID
	baseSymbol    string
	quoteSymbol   string
	exchangeRate  decimal.Decimal
	inverseRate   decimal.Decimal
	feePercentage decimal.Decimal
	minSwapAmount decimal.Decimal
	maxSwapAmount *decimal.Decimal
	dailyVolume   decimal.Decimal
	isActive      bool
	hasLiquidity  bool
	lastUpdated   time.Time
	createdAt     time.Time
	updatedAt     time.Time
}

// TradingPairParams captures the fields required to construct a TradingPairEntity.
type TradingPairParams struct {
	ID            uuid.UUID
	BaseSymbol    string
	QuoteSymbol   string
	ExchangeRate  decimal.Decimal
	InverseRate   decimal.Decimal
	FeePercentage decimal.Decimal
	MinSwapAmount decimal.Decimal
	MaxSwapAmount *decimal.Decimal
	DailyVolume   decimal.Decimal
	IsActive      bool
	HasLiquidity  bool
	LastUpdated   time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewTradingPairEntity validates the supplied parameters and returns a new TradingPairEntity instance.
func NewTradingPairEntity(params TradingPairParams) (*TradingPairEntity, error) {
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
		params.LastUpdated = params.CreatedAt
	}

	entity := &TradingPairEntity{
		id:            params.ID,
		baseSymbol:    strings.ToUpper(strings.TrimSpace(params.BaseSymbol)),
		quoteSymbol:   strings.ToUpper(strings.TrimSpace(params.QuoteSymbol)),
		exchangeRate:  params.ExchangeRate,
		inverseRate:   params.InverseRate,
		feePercentage: params.FeePercentage,
		minSwapAmount: params.MinSwapAmount,
		maxSwapAmount: params.MaxSwapAmount,
		dailyVolume:   params.DailyVolume,
		isActive:      params.IsActive,
		hasLiquidity:  params.HasLiquidity,
		lastUpdated:   params.LastUpdated,
		createdAt:     params.CreatedAt,
		updatedAt:     params.UpdatedAt,
	}

	// Set default values
	if entity.feePercentage.IsZero() {
		entity.feePercentage = decimal.NewFromFloat(0.5) // Default 0.5% fee
	}

	if entity.minSwapAmount.IsZero() {
		entity.minSwapAmount = decimal.Zero
	}

	if entity.dailyVolume.IsZero() {
		entity.dailyVolume = decimal.Zero
	}

	entity.isActive = true
	entity.hasLiquidity = true

	if err := entity.Validate(); err != nil {
		return nil, err
	}

	return entity, nil
}

// HydrateTradingPairEntity creates a TradingPairEntity without re-validating invariants (used for repository hydration).
func HydrateTradingPairEntity(params TradingPairParams) *TradingPairEntity {
	return &TradingPairEntity{
		id:            params.ID,
		baseSymbol:    strings.ToUpper(strings.TrimSpace(params.BaseSymbol)),
		quoteSymbol:   strings.ToUpper(strings.TrimSpace(params.QuoteSymbol)),
		exchangeRate:  params.ExchangeRate,
		inverseRate:   params.InverseRate,
		feePercentage: params.FeePercentage,
		minSwapAmount: params.MinSwapAmount,
		maxSwapAmount: params.MaxSwapAmount,
		dailyVolume:   params.DailyVolume,
		isActive:      params.IsActive,
		hasLiquidity:  params.HasLiquidity,
		lastUpdated:   params.LastUpdated,
		createdAt:     params.CreatedAt,
		updatedAt:     params.UpdatedAt,
	}
}

// Validate ensures the entity adheres to domain invariants.
func (t *TradingPairEntity) Validate() error {
	var validationErr error

	if t.baseSymbol == "" {
		validationErr = errors.Join(validationErr, errTradingPairBaseSymbolRequired)
	}

	if t.quoteSymbol == "" {
		validationErr = errors.Join(validationErr, errTradingPairQuoteSymbolRequired)
	}

	if t.baseSymbol == t.quoteSymbol {
		validationErr = errors.Join(validationErr, errTradingPairSameSymbols)
	}

	if t.exchangeRate.LessThanOrEqual(decimal.Zero) {
		validationErr = errors.Join(validationErr, errTradingPairRateInvalid)
	}

	if t.inverseRate.LessThanOrEqual(decimal.Zero) {
		validationErr = errors.Join(validationErr, errTradingPairInverseRateInvalid)
	}

	if t.feePercentage.IsNegative() {
		validationErr = errors.Join(validationErr, errTradingPairFeeInvalid)
	}

	if t.minSwapAmount.IsNegative() {
		validationErr = errors.Join(validationErr, errTradingPairMinAmountInvalid)
	}

	if t.maxSwapAmount != nil && t.maxSwapAmount.LessThan(t.minSwapAmount) {
		validationErr = errors.Join(validationErr, errTradingPairMaxAmountInvalid)
	}

	if t.dailyVolume.IsNegative() {
		validationErr = errors.Join(validationErr, errTradingPairDailyVolumeInvalid)
	}

	return validationErr
}

// Getter implementations satisfy the TradingPair interface.

func (t *TradingPairEntity) GetID() uuid.UUID {
	return t.id
}

func (t *TradingPairEntity) GetBaseSymbol() string {
	return t.baseSymbol
}

func (t *TradingPairEntity) GetQuoteSymbol() string {
	return t.quoteSymbol
}

func (t *TradingPairEntity) GetExchangeRate() decimal.Decimal {
	return t.exchangeRate
}

func (t *TradingPairEntity) GetInverseRate() decimal.Decimal {
	return t.inverseRate
}

func (t *TradingPairEntity) GetFeePercentage() decimal.Decimal {
	return t.feePercentage
}

func (t *TradingPairEntity) GetMinSwapAmount() decimal.Decimal {
	return t.minSwapAmount
}

func (t *TradingPairEntity) GetMaxSwapAmount() *decimal.Decimal {
	return t.maxSwapAmount
}

func (t *TradingPairEntity) GetDailyVolume() decimal.Decimal {
	return t.dailyVolume
}

func (t *TradingPairEntity) IsActive() bool {
	return t.isActive
}

func (t *TradingPairEntity) HasLiquidity() bool {
	return t.hasLiquidity
}

func (t *TradingPairEntity) GetLastUpdated() time.Time {
	return t.lastUpdated
}

func (t *TradingPairEntity) GetCreatedAt() time.Time {
	return t.createdAt
}

func (t *TradingPairEntity) GetUpdatedAt() time.Time {
	return t.updatedAt
}

// Domain behavior helpers.

// UpdateRates updates the exchange rate and inverse rate.
func (t *TradingPairEntity) UpdateRates(exchangeRate decimal.Decimal) error {
	if exchangeRate.LessThanOrEqual(decimal.Zero) {
		return errTradingPairRateInvalid
	}

	t.exchangeRate = exchangeRate
	t.inverseRate = decimal.NewFromInt(1).Div(exchangeRate)
	t.lastUpdated = time.Now().UTC()
	t.Touch(t.lastUpdated)

	return nil
}

// SetFeePercentage updates the fee percentage.
func (t *TradingPairEntity) SetFeePercentage(feePercentage decimal.Decimal) error {
	if feePercentage.IsNegative() {
		return errTradingPairFeeInvalid
	}

	t.feePercentage = feePercentage
	t.lastUpdated = time.Now().UTC()
	t.Touch(t.lastUpdated)

	return nil
}

// SetMinMaxAmounts updates the minimum and maximum swap amounts.
func (t *TradingPairEntity) SetMinMaxAmounts(min, max *decimal.Decimal) error {
	if min != nil && min.IsNegative() {
		return errTradingPairMinAmountInvalid
	}

	if max != nil && min != nil && max.LessThan(*min) {
		return errTradingPairMaxAmountInvalid
	}

	if min != nil {
		t.minSwapAmount = *min
	}

	if max != nil {
		t.maxSwapAmount = max
	}

	t.lastUpdated = time.Now().UTC()
	t.Touch(t.lastUpdated)

	return nil
}

// SetActive sets the active status of the trading pair.
func (t *TradingPairEntity) SetActive(active bool) {
	t.isActive = active
	t.lastUpdated = time.Now().UTC()
	t.Touch(t.lastUpdated)
}

// SetLiquidity sets the liquidity availability status.
func (t *TradingPairEntity) SetLiquidity(hasLiquidity bool) {
	t.hasLiquidity = hasLiquidity
	t.lastUpdated = time.Now().UTC()
	t.Touch(t.lastUpdated)
}

// AddVolume adds to the daily volume (used for tracking swap activity).
func (t *TradingPairEntity) AddVolume(volume decimal.Decimal) error {
	if volume.IsNegative() {
		return errTradingPairDailyVolumeInvalid
	}

	t.dailyVolume = t.dailyVolume.Add(volume)
	t.lastUpdated = time.Now().UTC()
	t.Touch(t.lastUpdated)

	return nil
}

// ResetDailyVolume resets the daily volume to zero (typically called at midnight).
func (t *TradingPairEntity) ResetDailyVolume() {
	t.dailyVolume = decimal.Zero
	t.lastUpdated = time.Now().UTC()
	t.Touch(t.lastUpdated)
}

// CalculateFeeAmount calculates the fee amount for a given swap amount.
func (t *TradingPairEntity) CalculateFeeAmount(amount decimal.Decimal) decimal.Decimal {
	return amount.Mul(t.feePercentage).Div(decimal.NewFromInt(100))
}

// CalculateReceivedAmount calculates the amount received after fees for a given input amount.
func (t *TradingPairEntity) CalculateReceivedAmount(fromAmount decimal.Decimal) decimal.Decimal {
	feeAmount := t.CalculateFeeAmount(fromAmount)
	netAmount := fromAmount.Sub(feeAmount)
	return netAmount.Mul(t.exchangeRate)
}

// CalculateRequiredAmount calculates the required input amount for a desired output amount.
func (t *TradingPairEntity) CalculateRequiredAmount(toAmount decimal.Decimal) decimal.Decimal {
	// Account for fees: required = (desired / rate) / (1 - fee_percentage/100)
	feeMultiplier := decimal.NewFromInt(1).Sub(t.feePercentage.Div(decimal.NewFromInt(100)))
	grossAmount := toAmount.Div(t.exchangeRate)
	requiredAmount := grossAmount.Div(feeMultiplier)
	return requiredAmount
}

// IsValidAmount checks if an amount is within the allowed swap range.
func (t *TradingPairEntity) IsValidAmount(amount decimal.Decimal) bool {
	if amount.LessThan(t.minSwapAmount) {
		return false
	}

	if t.maxSwapAmount != nil && amount.GreaterThan(*t.maxSwapAmount) {
		return false
	}

	return true
}

// GetPairIdentifier returns a unique identifier for the trading pair.
func (t *TradingPairEntity) GetPairIdentifier() string {
	return t.baseSymbol + "/" + t.quoteSymbol
}

// IsInversePair checks if this pair is the inverse of another pair.
func (t *TradingPairEntity) IsInversePair(other TradingPair) bool {
	return t.baseSymbol == other.GetQuoteSymbol() && t.quoteSymbol == other.GetBaseSymbol()
}

// Touch refreshes the updatedAt timestamp.
func (t *TradingPairEntity) Touch(at time.Time) {
	if at.IsZero() {
		t.updatedAt = time.Now().UTC()
		return
	}
	t.updatedAt = at
}
