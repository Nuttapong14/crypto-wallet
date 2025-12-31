package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// IntervalType represents the time interval for price history data points.
type IntervalType string

const (
	Interval1m  IntervalType = "1m"
	Interval5m  IntervalType = "5m"
	Interval15m IntervalType = "15m"
	Interval1h  IntervalType = "1h"
	Interval4h  IntervalType = "4h"
	Interval1d  IntervalType = "1d"
	Interval1w  IntervalType = "1w"
)

var (
	errPriceHistorySymbolRequired     = errors.New("price history symbol is required")
	errPriceHistoryIntervalInvalid    = errors.New("price history interval is invalid")
	errPriceHistoryTimestampInvalid   = errors.New("price history timestamp is required")
	errPriceHistoryOpenInvalid        = errors.New("price history open must be positive")
	errPriceHistoryHighInvalid        = errors.New("price history high must be positive")
	errPriceHistoryLowInvalid         = errors.New("price history low must be positive")
	errPriceHistoryCloseInvalid       = errors.New("price history close must be positive")
	errPriceHistoryVolumeInvalid      = errors.New("price history volume must be non-negative")
	errPriceHistoryHighLowInvalid     = errors.New("price history high must be >= low")
)

// PriceHistory exposes the behavior required by the application layer when working with price history entities.
type PriceHistory interface {
	Entity
	Identifiable
	Timestamped

	GetSymbol() string
	GetInterval() IntervalType
	GetTimestamp() time.Time
	GetOpen() decimal.Decimal
	GetHigh() decimal.Decimal
	GetLow() decimal.Decimal
	GetClose() decimal.Decimal
	GetVolume() decimal.Decimal
	Touch(at time.Time)
}

// PriceHistoryEntity is the default implementation of the PriceHistory interface.
type PriceHistoryEntity struct {
	id        uuid.UUID
	symbol    string
	interval  IntervalType
	timestamp time.Time
	open      decimal.Decimal
	high      decimal.Decimal
	low       decimal.Decimal
	close     decimal.Decimal
	volume    decimal.Decimal
	createdAt time.Time
	updatedAt time.Time
}

// PriceHistoryParams captures the fields required to construct a PriceHistoryEntity.
type PriceHistoryParams struct {
	ID        uuid.UUID
	Symbol    string
	Interval  IntervalType
	Timestamp time.Time
	Open      decimal.Decimal
	High      decimal.Decimal
	Low       decimal.Decimal
	Close     decimal.Decimal
	Volume    decimal.Decimal
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewPriceHistoryEntity validates the supplied parameters and returns a new PriceHistoryEntity instance.
func NewPriceHistoryEntity(params PriceHistoryParams) (*PriceHistoryEntity, error) {
	if params.ID == uuid.Nil {
		params.ID = uuid.New()
	}

	if params.CreatedAt.IsZero() {
		params.CreatedAt = time.Now().UTC()
	}

	if params.UpdatedAt.IsZero() {
		params.UpdatedAt = params.CreatedAt
	}

	entity := &PriceHistoryEntity{
		id:        params.ID,
		symbol:    strings.ToUpper(strings.TrimSpace(params.Symbol)),
		interval:  params.Interval,
		timestamp: params.Timestamp.UTC(),
		open:      params.Open,
		high:      params.High,
		low:       params.Low,
		close:     params.Close,
		volume:    params.Volume,
		createdAt: params.CreatedAt,
		updatedAt: params.UpdatedAt,
	}

	if err := entity.Validate(); err != nil {
		return nil, err
	}

	return entity, nil
}

// HydratePriceHistoryEntity creates a PriceHistoryEntity without re-validating invariants (used for repository hydration).
func HydratePriceHistoryEntity(params PriceHistoryParams) *PriceHistoryEntity {
	return &PriceHistoryEntity{
		id:        params.ID,
		symbol:    strings.ToUpper(strings.TrimSpace(params.Symbol)),
		interval:  params.Interval,
		timestamp: params.Timestamp.UTC(),
		open:      params.Open,
		high:      params.High,
		low:       params.Low,
		close:     params.Close,
		volume:    params.Volume,
		createdAt: params.CreatedAt,
		updatedAt: params.UpdatedAt,
	}
}

// Validate ensures the entity adheres to domain invariants.
func (p *PriceHistoryEntity) Validate() error {
	var validationErr error

	if strings.TrimSpace(p.symbol) == "" {
		validationErr = errors.Join(validationErr, errPriceHistorySymbolRequired)
	}

	if !isValidInterval(p.interval) {
		validationErr = errors.Join(validationErr, errPriceHistoryIntervalInvalid)
	}

	if p.timestamp.IsZero() {
		validationErr = errors.Join(validationErr, errPriceHistoryTimestampInvalid)
	}

	if p.open.IsNegative() || p.open.IsZero() {
		validationErr = errors.Join(validationErr, errPriceHistoryOpenInvalid)
	}

	if p.high.IsNegative() || p.high.IsZero() {
		validationErr = errors.Join(validationErr, errPriceHistoryHighInvalid)
	}

	if p.low.IsNegative() || p.low.IsZero() {
		validationErr = errors.Join(validationErr, errPriceHistoryLowInvalid)
	}

	if p.close.IsNegative() || p.close.IsZero() {
		validationErr = errors.Join(validationErr, errPriceHistoryCloseInvalid)
	}

	if p.volume.IsNegative() {
		validationErr = errors.Join(validationErr, errPriceHistoryVolumeInvalid)
	}

	if p.high.LessThan(p.low) {
		validationErr = errors.Join(validationErr, errPriceHistoryHighLowInvalid)
	}

	return validationErr
}

// Getter implementations satisfy the PriceHistory interface.

func (p *PriceHistoryEntity) GetID() uuid.UUID {
	return p.id
}

func (p *PriceHistoryEntity) GetSymbol() string {
	return p.symbol
}

func (p *PriceHistoryEntity) GetInterval() IntervalType {
	return p.interval
}

func (p *PriceHistoryEntity) GetTimestamp() time.Time {
	return p.timestamp
}

func (p *PriceHistoryEntity) GetOpen() decimal.Decimal {
	return p.open
}

func (p *PriceHistoryEntity) GetHigh() decimal.Decimal {
	return p.high
}

func (p *PriceHistoryEntity) GetLow() decimal.Decimal {
	return p.low
}

func (p *PriceHistoryEntity) GetClose() decimal.Decimal {
	return p.close
}

func (p *PriceHistoryEntity) GetVolume() decimal.Decimal {
	return p.volume
}

func (p *PriceHistoryEntity) GetCreatedAt() time.Time {
	return p.createdAt
}

func (p *PriceHistoryEntity) GetUpdatedAt() time.Time {
	return p.updatedAt
}

// Domain behavior helpers.

// Touch refreshes the updatedAt timestamp.
func (p *PriceHistoryEntity) Touch(at time.Time) {
	if at.IsZero() {
		p.updatedAt = time.Now().UTC()
		return
	}
	p.updatedAt = at
}

func isValidInterval(interval IntervalType) bool {
	switch interval {
	case Interval1m, Interval5m, Interval15m, Interval1h, Interval4h, Interval1d, Interval1w:
		return true
	default:
		return false
	}
}
