package middleware

import (
	"context"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// KYCEnforcerConfig configures the KYC enforcement middleware.
type KYCEnforcerConfig struct {
	Repository repositories.KYCRepository
	Logger     *slog.Logger
	ContextKey string
}

// KYCEnforcer enforces minimum verification levels for protected routes.
type KYCEnforcer struct {
	repo       repositories.KYCRepository
	logger     *slog.Logger
	contextKey string
}

// NewKYCEnforcer constructs a KYC enforcement middleware helper.
func NewKYCEnforcer(cfg KYCEnforcerConfig) *KYCEnforcer {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	contextKey := cfg.ContextKey
	if strings.TrimSpace(contextKey) == "" {
		contextKey = AuthContextKey
	}
	return &KYCEnforcer{
		repo:       cfg.Repository,
		logger:     cfg.Logger,
		contextKey: contextKey,
	}
}

// Require returns a Fiber middleware enforcing the supplied verification level.
func (e *KYCEnforcer) Require(level entities.VerificationLevel) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if e.repo == nil {
			return c.Next()
		}

		rawClaims := c.Locals(e.contextKey)
		userID, err := extractUserIDFromContext(rawClaims)
		if err != nil {
			resp, status := utils.ToErrorResponse(utils.NewAppError(
				"KYC_AUTH_REQUIRED",
				"authentication required",
				fiber.StatusUnauthorized,
				err,
				nil,
			))
			return c.Status(status).JSON(resp)
		}

		profile, err := e.repo.GetProfileByUserID(context.Background(), userID)
		if err != nil {
			resp, status := utils.ToErrorResponse(utils.NewAppError(
				"KYC_PROFILE_REQUIRED",
				"complete identity verification to access this feature",
				fiber.StatusForbidden,
				err,
				nil,
			))
			return c.Status(status).JSON(resp)
		}

		if compareLevel(profile.GetVerificationLevel(), level) < 0 || profile.GetStatus() != entities.KYCStatusApproved {
			resp, status := utils.ToErrorResponse(utils.NewAppError(
				"KYC_LEVEL_INSUFFICIENT",
				"additional verification required to access this feature",
				fiber.StatusForbidden,
				nil,
				map[string]any{
					"requiredLevel": level,
					"currentLevel":  profile.GetVerificationLevel(),
					"status":        profile.GetStatus(),
				},
			))
			return c.Status(status).JSON(resp)
		}

		return c.Next()
	}
}

func compareLevel(current, required entities.VerificationLevel) int {
	return levelWeight(current) - levelWeight(required)
}

func levelWeight(level entities.VerificationLevel) int {
	switch level {
	case entities.VerificationLevelFull:
		return 3
	case entities.VerificationLevelBasic:
		return 2
	case entities.VerificationLevelUnverified:
		return 1
	default:
		return 0
	}
}

func extractUserIDFromContext(claims any) (uuid.UUID, error) {
	switch value := claims.(type) {
	case map[string]any:
		if v, ok := value["user_id"].(string); ok && v != "" {
			return uuid.Parse(v)
		}
		if v, ok := value["sub"].(string); ok && v != "" {
			return uuid.Parse(v)
		}
	case string:
		if value != "" {
			return uuid.Parse(value)
		}
	}
	return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "user id missing from context")
}
