package agentconfig

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestCanonicalAgentRoutesRequireExplicitWorkspaceID(t *testing.T) {
	handler := NewCanonicalHandler(nil)
	app := fiber.New()
	app.Get("/agent", handler.Get)
	app.Patch("/agent", handler.Update)
	app.Post("/agent/test", handler.Test)

	for _, request := range []*http.Request{
		httptest.NewRequest(fiber.MethodGet, "/agent", nil),
		httptest.NewRequest(fiber.MethodPatch, "/agent", strings.NewReader(`{"name":"Naya"}`)),
		httptest.NewRequest(fiber.MethodPost, "/agent/test", strings.NewReader(`{"message":"hello"}`)),
	} {
		response, err := app.Test(request)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, response.StatusCode)
	}
}
