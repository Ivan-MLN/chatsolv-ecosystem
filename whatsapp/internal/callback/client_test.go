package callback

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type mockSender struct {
	sentChannelID         string
	sentJID               types.JID
	sentText              string
	sentQuotedMsgID       string
	sentQuotedParticipant string
	sentQuotedMsg         *waProto.Message
}

func (m *mockSender) SendText(ctx context.Context, channelID string, jid types.JID, text string, quotedMsgID string, quotedParticipant string, quotedMsg *waProto.Message) error {
	m.sentChannelID = channelID
	m.sentJID = jid
	m.sentText = text
	m.sentQuotedMsgID = quotedMsgID
	m.sentQuotedParticipant = quotedParticipant
	m.sentQuotedMsg = quotedMsg
	return nil
}

func (m *mockSender) SendPresence(context.Context, string, types.JID, types.ChatPresence) error {
	return nil
}

func (m *mockSender) DownloadAttachment(ctx context.Context, channelID string, msg whatsmeow.DownloadableMessage) ([]byte, error) {
	return nil, nil
}

func (m *mockSender) ResolvePhoneNumber(ctx context.Context, channelID string, jid types.JID) types.JID {
	return jid
}

func TestCallbackClientSendsHMAC(t *testing.T) {
	secret := "test-secret-at-least-32-bytes-long"
	var receivedSig, receivedTS string
	var receivedBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSig = r.Header.Get("X-ChatSolv-Signature")
		receivedTS = r.Header.Get("X-ChatSolv-Timestamp")
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
	}))
	defer srv.Close()

	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	client := New(srv.URL, secret, 5*time.Second, log)
	fixedTime := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	client.now = func() time.Time { return fixedTime }

	payload := map[string]string{"foo": "bar"}
	err := client.Post(context.Background(), "/test", payload)
	require.NoError(t, err)

	expectedBody, _ := json.Marshal(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fixedTime.Format(time.RFC3339) + "." + string(expectedBody)))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	require.Equal(t, fixedTime.Format(time.RFC3339), receivedTS)
	require.Equal(t, expectedSig, receivedSig)
	require.Equal(t, expectedBody, receivedBody)
}

func TestHandlerOnMessageSendsReply(t *testing.T) {
	secret := "test-secret-at-least-32-bytes-long"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/internal/v1/messages/incoming", r.URL.Path)
		var in IncomingMessage
		require.NoError(t, json.NewDecoder(r.Body).Decode(&in))
		require.Equal(t, "ch123", in.ChannelID)
		require.Equal(t, "user123", in.ExternalUserID)
		require.Equal(t, "Hello bot", in.Content.Text)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(BackendResponse[IncomingMessageResult]{
			Message: "Message processed",
			Data: IncomingMessageResult{
				MessageID:        "msg_1",
				ConversationID:   "conv_1",
				Content:          "Hello from Hermes!",
				HandoffRequested: false,
			},
		})
	}))
	defer srv.Close()

	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	client := New(srv.URL, secret, 5*time.Second, log)
	handler := NewHandler(client, log)
	sender := &mockSender{}
	handler.SetSender(sender)

	chatJID := types.NewJID("user123", types.DefaultUserServer)
	msgText := "Hello bot"
	msg := &events.Message{
		Info: types.MessageInfo{
			ID:        "wamid_test_123",
			Sender:    chatJID,
			Chat:      chatJID,
			IsFromMe:  false,
			Timestamp: time.Now(),
		},
		Message: &waProto.Message{
			Conversation: &msgText,
		},
	}

	handler.Handle("ch123", msg)

	require.Equal(t, "ch123", sender.sentChannelID)
	require.Equal(t, chatJID, sender.sentJID)
	require.Equal(t, "Hello from Hermes!", sender.sentText)
}

func TestRandomMediaNameIsOpaqueAndPathSafe(t *testing.T) {
	name, err := randomMediaName("doc", ".pdf")
	require.NoError(t, err)
	require.Regexp(t, `^doc_[a-f0-9]{48}\.pdf$`, name)
	require.NotContains(t, name, "/")
	require.NotContains(t, name, `\`)
}

func TestAttachmentSizeLimit(t *testing.T) {
	require.False(t, validAttachmentSize(nil))
	require.True(t, validAttachmentSize(make([]byte, maxAttachmentBytes)))
	require.False(t, validAttachmentSize(make([]byte, maxAttachmentBytes+1)))
}
