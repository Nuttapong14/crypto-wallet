package httpserver

import (
	"github.com/gofiber/fiber/v2"

	"github.com/crypto-wallet/backend/internal/interfaces/http/handlers"
	"github.com/crypto-wallet/backend/internal/interfaces/http/middleware"
)

// SetupExchangeRoutes registers all exchange-related routes.
func SetupExchangeRoutes(app *fiber.App, exchangeHandler *handlers.ExchangeHandler) {
	// Public routes for exchange rates and trading pairs
	api := app.Group("/api/v1/exchange")

	// Get exchange rate for a trading pair
	api.Get("/rate", exchangeHandler.GetExchangeRate)

	// Get all active trading pairs
	api.Get("/pairs", exchangeHandler.GetActiveTradingPairs)

	// Protected routes (require authentication)
	protected := api.Group("/", middleware.AuthMiddleware())

	// Quote generation
	protected.Post("/quote", exchangeHandler.GetQuote)

	// Execute swap
	protected.Post("/execute", exchangeHandler.ExecuteSwap)

	// Cancel swap
	protected.Post("/cancel", exchangeHandler.CancelSwap)

	// User-specific routes
	userRoutes := protected.Group("/user/:userID")

	// Get exchange history for a user
	userRoutes.Get("/history", exchangeHandler.GetExchangeHistory)

	// Get exchange statistics for a user
	userRoutes.Get("/stats", exchangeHandler.GetExchangeStats)
}
