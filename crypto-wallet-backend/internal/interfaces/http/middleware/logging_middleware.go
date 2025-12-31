package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	appLogging "github.com/crypto-wallet/backend/internal/infrastructure/logging"
)

// NewLoggingMiddleware provides structured request logging using slog.
func NewLoggingMiddleware(logger *slog.Logger) fiber.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)
		status := c.Response().StatusCode()

		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}

		ctx := c.UserContext()
		reqLogger := appLogging.LoggerFromContext(ctx, logger)

		attrs := []slog.Attr{
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("ip", c.IP()),
			slog.String("user_agent", string(c.Request().Header.UserAgent())),
		}

		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}

		reqLogger.LogAttrs(context.Background(), level, "request completed", attrs...)
		return err
	}
}
