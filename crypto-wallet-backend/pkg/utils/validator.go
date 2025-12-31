package utils

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ValidationError represents a single field validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.Message)
}

// ValidationErrors aggregates multiple validation failures.
type ValidationErrors []ValidationError

// Error implements the error interface.
func (v ValidationErrors) Error() string {
	switch len(v) {
	case 0:
		return ""
	case 1:
		return v[0].Error()
	default:
		builder := strings.Builder{}
		for i, ve := range v {
			if i > 0 {
				builder.WriteString("; ")
			}
			builder.WriteString(ve.Error())
		}
		return builder.String()
	}
}

// IsEmpty reports whether there are any validation errors.
func (v ValidationErrors) IsEmpty() bool {
	return len(v) == 0
}

// Add appends a new validation error.
func (v *ValidationErrors) Add(field, message string) {
	*v = append(*v, ValidationError{
		Field:   field,
		Message: message,
	})
}

// ToDetails converts validation errors to a map format for API responses.
func (v ValidationErrors) ToDetails() map[string]any {
	if v.IsEmpty() {
		return nil
	}
	details := make(map[string]any, len(v))
	for _, err := range v {
		details[err.Field] = err.Message
	}
	return details
}

// Require ensures a string is non-empty.
func Require(errs *ValidationErrors, field, value string) {
	if strings.TrimSpace(value) == "" {
		errs.Add(field, "is required")
	}
}

// RequireUUID ensures the value is a valid UUID string.
func RequireUUID(errs *ValidationErrors, field, value string) {
	Require(errs, field, value)
	if strings.TrimSpace(value) == "" {
		return
	}
	if _, err := uuid.Parse(value); err != nil {
		errs.Add(field, "must be a valid UUID")
	}
}

// RequireEmail ensures the value is a valid email address.
func RequireEmail(errs *ValidationErrors, field, value string) {
	Require(errs, field, value)
	if strings.TrimSpace(value) == "" {
		return
	}
	if _, err := mail.ParseAddress(value); err != nil {
		errs.Add(field, "must be a valid email address")
	}
}

// RequireMinLength enforces minimum rune length.
func RequireMinLength(errs *ValidationErrors, field, value string, min int) {
	if utf8.RuneCountInString(strings.TrimSpace(value)) < min {
		errs.Add(field, fmt.Sprintf("must be at least %d characters", min))
	}
}

// RequireMaxLength enforces maximum rune length.
func RequireMaxLength(errs *ValidationErrors, field, value string, max int) {
	if utf8.RuneCountInString(value) > max {
		errs.Add(field, fmt.Sprintf("must be at most %d characters", max))
	}
}

// RequireInSet ensures the value belongs to the provided allowed set.
func RequireInSet(errs *ValidationErrors, field, value string, allowed []string) {
	if len(allowed) == 0 {
		return
	}
	for _, candidate := range allowed {
		if value == candidate {
			return
		}
	}
	errs.Add(field, fmt.Sprintf("must be one of %s", strings.Join(allowed, ", ")))
}

// RequirePattern ensures the value matches the supplied regular expression pattern.
func RequirePattern(errs *ValidationErrors, field, value, pattern, message string) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		errs.Add(field, "is required")
		return
	}

	matches, err := regexp.MatchString(pattern, trimmed)
	if err != nil || !matches {
		if strings.TrimSpace(message) == "" {
			message = "is invalid"
		}
		errs.Add(field, message)
	}
}

// RequirePositiveDecimal ensures the provided decimal is strictly positive.
func RequirePositiveDecimal(errs *ValidationErrors, field string, value decimal.Decimal) {
	if !value.GreaterThan(decimal.Zero) {
		errs.Add(field, "must be greater than zero")
	}
}

// RequireNonNegativeDecimal ensures the provided decimal is >= 0.
func RequireNonNegativeDecimal(errs *ValidationErrors, field string, value decimal.Decimal) {
	if value.IsNegative() {
		errs.Add(field, "must be greater than or equal to zero")
	}
}
