package auth

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestJWTManagerParseReturnsSubject(t *testing.T) {
	manager := NewJWTManager([]byte("01234567890123456789012345678901"), time.Minute)
	token, _, err := manager.Generate("user-1")
	require.NoError(t, err)

	subject, err := manager.Parse(token)

	require.NoError(t, err)
	require.Equal(t, "user-1", subject)
}

func TestRequireAccessTokenSetsAuthenticatedUser(t *testing.T) {
	manager := NewJWTManager([]byte("01234567890123456789012345678901"), time.Minute)
	token, _, err := manager.Generate("user-1")
	require.NoError(t, err)
	app := fiber.New()
	app.Get("/private", RequireAccessToken(manager), func(c *fiber.Ctx) error {
		return c.SendString(AuthenticatedUserID(c))
	})
	req := httptest.NewRequest("GET", "/private", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, res.StatusCode)
}

func TestRequireAccessTokenRejectsMissingAndInvalidToken(t *testing.T) {
	manager := NewJWTManager([]byte("01234567890123456789012345678901"), time.Minute)
	app := fiber.New()
	app.Get("/private", RequireAccessToken(manager), func(c *fiber.Ctx) error { return c.SendStatus(200) })

	for _, header := range []string{"", "Basic abc", "Bearer invalid"} {
		req := httptest.NewRequest("GET", "/private", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		res, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusUnauthorized, res.StatusCode)
	}
}
