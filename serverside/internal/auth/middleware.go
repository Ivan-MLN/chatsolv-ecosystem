package auth

import (
	"authbackend/pkg/response"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const authenticatedUserIDKey = "authenticated_user_id"

func RequireAccessToken(manager *JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		scheme, token, ok := strings.Cut(c.Get(fiber.HeaderAuthorization), " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
			return response.Fail(c, fiber.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		}
		userID, err := manager.Parse(strings.TrimSpace(token))
		if err != nil {
			return response.Fail(c, fiber.StatusUnauthorized, "Invalid access token", "UNAUTHORIZED")
		}
		c.Locals(authenticatedUserIDKey, userID)
		return c.Next()
	}
}

func AuthenticatedUserID(c *fiber.Ctx) string {
	value, _ := c.Locals(authenticatedUserIDKey).(string)
	return value
}
