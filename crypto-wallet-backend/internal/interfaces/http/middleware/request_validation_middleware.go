package middleware

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// RequestValidationConfig exposes input validation constraints.
type RequestValidationConfig struct {
	MaxBodyBytes int64
	EnforceJSON  bool
}

// NewRequestValidationMiddleware enforces basic payload validation before handlers execute.
func NewRequestValidationMiddleware(cfg RequestValidationConfig) fiber.Handler {
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = 1 << 20 // 1 MiB
	}
	return func(c *fiber.Ctx) error {
		if cfg.EnforceJSON && hasBody(c) && requiresJSONContentType(c.Method()) {
			contentType := strings.ToLower(strings.TrimSpace(c.Get(fiber.HeaderContentType)))
			if !strings.HasPrefix(contentType, fiber.MIMEApplicationJSON) {
				return fiber.NewError(http.StatusUnsupportedMediaType, "content-type must be application/json")
			}
		}

	if hasBody(c) {
		if length := c.Request().Header.ContentLength(); length > 0 && int64(length) > cfg.MaxBodyBytes {
			return fiber.NewError(http.StatusRequestEntityTooLarge, "request body is too large")
		}
		if int64(len(c.Body())) > cfg.MaxBodyBytes {
			return fiber.NewError(http.StatusRequestEntityTooLarge, "request body exceeds allowed size")
		}
	}

		return c.Next()
	}
}

func hasBody(c *fiber.Ctx) bool {
	contentLength := c.Request().Header.ContentLength()
	return contentLength > 0 || len(c.Request().Body()) > 0
}

func requiresJSONContentType(method string) bool {
	switch strings.ToUpper(method) {
	case fiber.MethodPost, fiber.MethodPut, fiber.MethodPatch:
		return true
	default:
		return false
	}
}
