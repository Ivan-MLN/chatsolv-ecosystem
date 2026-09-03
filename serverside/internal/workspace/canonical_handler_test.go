package workspace

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestCanonicalGetRequiresExplicitWorkspaceID(t *testing.T) {
	app := fiber.New()
	app.Get("/workspace", NewHandler(nil).CanonicalGet)

	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/workspace", nil))

	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, response.StatusCode)
}

func TestCanonicalUpdateRequiresExplicitWorkspaceID(t *testing.T) {
	app := fiber.New()
	app.Patch("/workspace", NewHandler(nil).CanonicalUpdate)
	request := httptest.NewRequest(fiber.MethodPatch, "/workspace", strings.NewReader(`{"name":"Updated"}`))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	response, err := app.Test(request)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, response.StatusCode)
}
