package auth

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestRegisterRejectsInvalidInputAtBoundary(t *testing.T) {
	app := fiber.New()
	app.Post("/", NewHandler(nil).Register)
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, 400, res.StatusCode)
}

func TestBindAcceptsJSONContentTypeCharset(t *testing.T) {
	h := NewHandler(nil)
	app := fiber.New()
	app.Post("/", func(c *fiber.Ctx) error {
		var r registerRequest
		_ = h.bind(c, &r)
		return nil
	})
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	res, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, 400, res.StatusCode)
}
