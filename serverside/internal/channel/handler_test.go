package channel

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestChannelListRequiresWorkspaceID(t *testing.T) {
	app := fiber.New()
	app.Get("/channels", NewHandler(nil).List)

	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/channels", nil))

	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, response.StatusCode)
}

func TestWhatsAppConnectRequiresWorkspaceID(t *testing.T) {
	app := fiber.New()
	app.Post("/channels/whatsapp/connect", NewHandler(nil).ConnectWhatsApp)

	response, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/channels/whatsapp/connect", nil))

	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, response.StatusCode)
}
