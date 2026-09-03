package agentconfig

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestCanonicalBusinessRoutesRequireExplicitWorkspaceID(t *testing.T) {
	app := fiber.New()
	app.Get("/business", func(c *fiber.Ctx) error {
		_, err := canonicalWorkspaceID(c)
		return err
	})

	result, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/business", nil))

	require.NoError(t, err)
	t.Cleanup(func() { _ = result.Body.Close() })
	require.Equal(t, fiber.StatusBadRequest, result.StatusCode)
}
