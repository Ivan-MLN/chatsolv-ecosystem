package publicapi

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestBearerTokenRejectsMissingAuthorization(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		_, err := bearerToken(c)
		return err
	})

	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))

	require.NoError(t, err)
	t.Cleanup(func() { _ = response.Body.Close() })
	require.Equal(t, fiber.StatusUnauthorized, response.StatusCode)
}

func TestBearerTokenAcceptsBearerScheme(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		token, err := bearerToken(c)
		if err != nil {
			return err
		}
		return c.SendString(token)
	})
	request := httptest.NewRequest(fiber.MethodGet, "/", nil)
	request.Header.Set(fiber.HeaderAuthorization, "Bearer csc_secret")

	response, err := app.Test(request)

	require.NoError(t, err)
	t.Cleanup(func() { _ = response.Body.Close() })
	require.Equal(t, fiber.StatusOK, response.StatusCode)
}
