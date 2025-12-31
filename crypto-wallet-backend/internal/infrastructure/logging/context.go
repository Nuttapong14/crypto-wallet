package logging

import (
	"context"
	"log/slog"
)

type contextKey string

const (
	loggerKey    contextKey = "logging.logger"
	requestIDKey contextKey = "logging.request_id"
	userIDKey    contextKey = "logging.user_id"
)

// ContextWithLogger stores a logger instance on the context.
func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, loggerKey, logger)
}

// ContextWithRequestID stores the request identifier on the context.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, requestIDKey, requestID)
}

// ContextWithUserID stores an authenticated user identifier on the context.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, userIDKey, userID)
}

// RequestIDFromContext extracts the request identifier if available.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if value, ok := ctx.Value(requestIDKey).(string); ok {
		return value
	}
	return ""
}

// UserIDFromContext extracts the authenticated user identifier if present.
func UserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if value, ok := ctx.Value(userIDKey).(string); ok {
		return value
	}
	return ""
}

// LoggerFromContext returns a logger enriched with context attributes.
func LoggerFromContext(ctx context.Context, fallback *slog.Logger) *slog.Logger {
	logger := fallback
	if ctx != nil {
		if value, ok := ctx.Value(loggerKey).(*slog.Logger); ok && value != nil {
			logger = value
		}
	}

	if logger == nil {
		logger = slog.Default()
	}

	requestID := RequestIDFromContext(ctx)
	if requestID != "" {
		logger = logger.With(slog.String("request_id", requestID))
	}

	if userID := UserIDFromContext(ctx); userID != "" {
		logger = logger.With(slog.String("user_id", userID))
	}

	return logger
}
