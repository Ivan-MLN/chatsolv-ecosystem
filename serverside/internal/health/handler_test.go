package health

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func check(err error) Check { return func(context.Context) error { return err } }

func TestLiveReturnsOK(t *testing.T) {
	app := fiber.New()
	app.Get("/health/live", NewHandler(check(nil), check(nil)).Live)
	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/health/live", nil))
	require.NoError(t, err)
	t.Cleanup(func() { _ = response.Body.Close() })
	require.Equal(t, fiber.StatusOK, response.StatusCode)
}

func TestReadyReturnsServiceUnavailableWhenDependencyFails(t *testing.T) {
	app := fiber.New()
	app.Get("/health/ready", NewHandler(check(errors.New("down")), check(nil)).Ready)
	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/health/ready", nil))
	require.NoError(t, err)
	t.Cleanup(func() { _ = response.Body.Close() })
	require.Equal(t, fiber.StatusServiceUnavailable, response.StatusCode)
}
