package apikey

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyRoutesRequireExplicitWorkspaceID(t *testing.T) {
	app := fiber.New()
	app.Get("/api-keys", func(c *fiber.Ctx) error {
		_, err := workspaceID(c)
		return err
	})

	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api-keys", nil))

	require.NoError(t, err)
	t.Cleanup(func() { _ = response.Body.Close() })
	require.Equal(t, fiber.StatusBadRequest, response.StatusCode)
}
