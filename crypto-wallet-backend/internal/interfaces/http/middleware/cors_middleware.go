package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSConfig wraps Fiber's cors.Config with sensible defaults for the project.
type CORSConfig struct {
	AllowOrigins     string
	AllowMethods     string
	AllowHeaders     string
	ExposeHeaders    string
	AllowCredentials bool
	MaxAge           int
}

// NewCORSMiddleware returns a configured CORS middleware.
func NewCORSMiddleware(cfg CORSConfig) fiber.Handler {
	config := cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	}

	if config.AllowOrigins == "" {
		config.AllowOrigins = "*"
	}
	if config.AllowMethods == "" {
		config.AllowMethods = "GET,POST,PUT,PATCH,DELETE,OPTIONS"
	}
	if config.AllowHeaders == "" {
		config.AllowHeaders = "Authorization,Content-Type,Accept,X-Request-ID"
	}

	return cors.New(config)
}
