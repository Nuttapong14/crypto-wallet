package middleware

import (
	"context"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"

	appLogging "github.com/crypto-wallet/backend/internal/infrastructure/logging"
	"github.com/crypto-wallet/backend/internal/infrastructure/security"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// AuthContextKey is the default key used to store JWT claims in the Fiber context.
const AuthContextKey = "auth.claims"

// AuthConfig configures the authentication middleware.
type AuthConfig struct {
	JWTService *security.JWTService
	Logger     *slog.Logger
	ContextKey string
	Skipper    func(*fiber.Ctx) bool
}

// NewAuthMiddleware builds a Fiber middleware that validates JWT bearer tokens.
func NewAuthMiddleware(cfg AuthConfig) fiber.Handler {
	if cfg.JWTService == nil {
		panic("middleware: JWTService is required for auth middleware")
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	contextKey := cfg.ContextKey
	if strings.TrimSpace(contextKey) == "" {
		contextKey = AuthContextKey
	}

	return func(c *fiber.Ctx) error {
		if cfg.Skipper != nil && cfg.Skipper(c) {
			return c.Next()
		}

		token, err := extractBearerToken(c.Get(fiber.HeaderAuthorization))
		if err != nil {
			cfg.Logger.Warn("authorization header invalid", slog.String("error", err.Error()))
			resp, status := utils.ToErrorResponse(fiber.NewError(fiber.StatusUnauthorized, err.Error()))
			return c.Status(status).JSON(resp)
		}

        claims, err := cfg.JWTService.Parse(context.Background(), token)
        if err != nil {
            cfg.Logger.Warn("token validation failed",
                slog.String("error", err.Error()),
                slog.String("path", string(c.Request().URI().Path())),
            )
            resp, status := utils.ToErrorResponse(fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token"))
            return c.Status(status).JSON(resp)
        }

		c.Locals(contextKey, claims)

		subject := strings.TrimSpace(claims.Subject)
		if subject != "" {
			c.Locals("user_id", subject)
			ctx := appLogging.ContextWithUserID(c.UserContext(), subject)
			c.SetUserContext(ctx)
		}

		if claims.Metadata != nil {
			if metadataID, ok := claims.Metadata["user_id"].(string); ok && strings.TrimSpace(metadataID) != "" {
				c.Locals("user_id", metadataID)
				ctx := appLogging.ContextWithUserID(c.UserContext(), metadataID)
				c.SetUserContext(ctx)
			}
		}

        return c.Next()
    }
}

func extractBearerToken(header string) (string, error) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "authorization header missing")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fiber.NewError(fiber.StatusUnauthorized, "authorization header must be Bearer token")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "authorization token is empty")
	}

	return token, nil
}
