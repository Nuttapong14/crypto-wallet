package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/crypto-wallet/backend/internal/domain/repositories"
)

// AppError represents a structured application error that can be serialised to API responses.
type AppError struct {
	Code    string
	Message string
	Status  int
	Err     error
	Details map[string]any
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap exposes the underlying error for errors.Is / errors.As.
func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// NewAppError constructs an AppError.
func NewAppError(code, message string, status int, err error, details map[string]any) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
		Err:     err,
		Details: details,
	}
}

// ErrorResponse is a serialisable representation of an AppError.
type ErrorResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// ToErrorResponse converts any error into a structured ErrorResponse and status code.
func ToErrorResponse(err error) (ErrorResponse, int) {
	if err == nil {
		return ErrorResponse{Code: "OK", Message: "success"}, http.StatusOK
	}

	status := HTTPStatusFromError(err)
	code := ErrorCodeFromError(err)
	message := SanitizeErrorMessage(err)

	var appErr *AppError
	if errors.As(err, &appErr) && len(appErr.Details) > 0 {
		return ErrorResponse{Code: code, Message: message, Details: appErr.Details}, status
	}

	return ErrorResponse{Code: code, Message: message}, status
}

// HTTPStatusFromError maps an error to an HTTP status code.
func HTTPStatusFromError(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, repositories.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, repositories.ErrDuplicate):
		return http.StatusConflict
	case errors.Is(err, context.Canceled):
		return http.StatusRequestTimeout
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) && fiberErr != nil {
		return fiberErr.Code
	}

	var appErr *AppError
	if errors.As(err, &appErr) && appErr.Status > 0 {
		return appErr.Status
	}

	return http.StatusInternalServerError
}

// ErrorCodeFromError derives a machine readable error code.
func ErrorCodeFromError(err error) string {
	if err == nil {
		return "OK"
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) && fiberErr != nil {
		return fmt.Sprintf("HTTP_%d", fiberErr.Code)
	}

	var appErr *AppError
	if errors.As(err, &appErr) && appErr.Code != "" {
		return strings.ToUpper(appErr.Code)
	}

	switch {
	case errors.Is(err, repositories.ErrNotFound):
		return "NOT_FOUND"
	case errors.Is(err, repositories.ErrDuplicate):
		return "DUPLICATE"
	case errors.Is(err, context.Canceled):
		return "REQUEST_CANCELED"
	case errors.Is(err, context.DeadlineExceeded):
		return "TIMEOUT"
	default:
		return "INTERNAL_ERROR"
	}
}

// SanitizeErrorMessage produces a safe, user-facing error message.
func SanitizeErrorMessage(err error) string {
	if err == nil {
		return "success"
	}

	var appErr *AppError
	if errors.As(err, &appErr) && appErr.Message != "" {
		return appErr.Message
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) && fiberErr != nil {
		return fiberErr.Message
	}

	switch {
	case errors.Is(err, repositories.ErrNotFound):
		return "resource not found"
	case errors.Is(err, repositories.ErrDuplicate):
		return "resource already exists"
	case errors.Is(err, context.Canceled):
		return "request was canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "request timed out"
	default:
		return "an unexpected error occurred"
	}
}

// IsNotFound helper.
func IsNotFound(err error) bool {
	return errors.Is(err, repositories.ErrNotFound)
}

// IsDuplicate helper.
func IsDuplicate(err error) bool {
	return errors.Is(err, repositories.ErrDuplicate)
}
