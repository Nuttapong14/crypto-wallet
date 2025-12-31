package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimitConfig configures the rate limiter middleware.
type RateLimitConfig struct {
	Enabled     bool
	MaxRequests int
	Window      time.Duration
	ExcludePaths []string
}

// NewRateLimitMiddleware creates a rate limiting middleware with sensible defaults.
func NewRateLimitMiddleware(cfg RateLimitConfig) fiber.Handler {
	if !cfg.Enabled {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	if cfg.MaxRequests <= 0 {
		cfg.MaxRequests = 100
	}

	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}

	baseHandler := limiter.New(limiter.Config{
		Max:        cfg.MaxRequests,
		Expiration: cfg.Window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() + ":" + c.Path()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return fiber.NewError(fiber.StatusTooManyRequests, "rate limit exceeded")
		},
	})

	return func(c *fiber.Ctx) error {
		for _, excluded := range cfg.ExcludePaths {
			if strings.HasPrefix(c.Path(), excluded) {
				return c.Next()
			}
		}
		return baseHandler(c)
	}
}
