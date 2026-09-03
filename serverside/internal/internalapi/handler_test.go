package internalapi

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestIncomingHandlerReadsSnakeCaseBotPayload(t *testing.T) {
	runtime := &fakeConversationService{}
	app := fiber.New()
	app.Post("/", NewHandler(NewService(&fakeRepository{}, runtime)).Incoming)
	request := httptest.NewRequest(fiber.MethodPost, "/", strings.NewReader(`{
		"channel_id":"channel-id",
		"external_message_id":"message-id",
		"external_user_id":"customer-id",
		"message_type":"text",
		"content":{"text":"Halo"}
	}`))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	response, err := app.Test(request)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, response.StatusCode)
	require.Equal(t, "channel-id", runtime.incoming.ChannelID)
	require.Equal(t, "message-id", runtime.incoming.ExternalMessageID)
	require.Equal(t, "customer-id", runtime.incoming.ExternalUserID)
}
