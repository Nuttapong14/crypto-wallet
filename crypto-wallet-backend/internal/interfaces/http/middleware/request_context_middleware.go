package middleware

import (
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appLogging "github.com/crypto-wallet/backend/internal/infrastructure/logging"
)

// NewRequestContextMiddleware attaches request-scoped logging metadata to the context.
func NewRequestContextMiddleware(logger *slog.Logger) fiber.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(c *fiber.Ctx) error {
		requestID := strings.TrimSpace(c.Get("X-Request-ID"))
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Locals("request_id", requestID)

		ctx := c.UserContext()
		requestLogger := logger.With(slog.String("request_id", requestID))
		ctx = appLogging.ContextWithRequestID(ctx, requestID)
		ctx = appLogging.ContextWithLogger(ctx, requestLogger)

		c.SetUserContext(ctx)
		c.Response().Header.Set("X-Request-ID", requestID)

		return c.Next()
	}
}
