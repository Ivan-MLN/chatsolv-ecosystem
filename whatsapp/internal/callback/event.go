package callback

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// MessageSender sends a plain-text reply back to a WhatsApp chat and can download attachments.
// It is satisfied by *whatsapp.Manager.
type MessageSender interface {
	SendPresence(ctx context.Context, channelID string, jid types.JID, state types.ChatPresence) error
	SendText(ctx context.Context, channelID string, jid types.JID, text string, quotedMsgID string, quotedParticipant string, quotedMsg *waE2E.Message) error
	DownloadAttachment(ctx context.Context, channelID string, msg whatsmeow.DownloadableMessage) ([]byte, error)
	ResolvePhoneNumber(ctx context.Context, channelID string, jid types.JID) types.JID
}

// Handler routes whatsmeow events to the appropriate backend callback paths.
// It satisfies the whatsapp.EventHandler signature:
//
//	func(channelID string, evt interface{})
type Handler struct {
	client *Client
	sender MessageSender
	log    *slog.Logger
}

// NewHandler creates a Handler backed by the given Client.
func NewHandler(client *Client, log *slog.Logger) *Handler {
	return &Handler{client: client, log: log}
}

// SetSender sets the MessageSender used to reply to incoming messages.
func (h *Handler) SetSender(sender MessageSender) {
	h.sender = sender
}

// Handle is the entry point. It is called by the whatsapp.Manager for every event.
func (h *Handler) Handle(channelID string, evt interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), h.client.http.Timeout)
	defer cancel()

	switch v := evt.(type) {
	case *events.Message:
		h.onMessage(ctx, channelID, v)
	case *events.Connected:
		h.onStatus(ctx, channelID, "connected", "", channelID)
	case *events.Disconnected:
		h.onStatus(ctx, channelID, "disconnected", "", channelID)
	case *events.LoggedOut:
		h.onStatus(ctx, channelID, "disconnected", "", channelID)
		h.onEvent(ctx, channelID, "logged_out", "", channelID)
	case *events.PairSuccess:
		phone := v.ID.User
		h.onStatus(ctx, channelID, "connected", phone, channelID)
		h.onEvent(ctx, channelID, "pair_success", phone, channelID)
	case *events.QR:
		h.onEvent(ctx, channelID, "qr_refresh", "", channelID)
	}
}

func (h *Handler) onMessage(ctx context.Context, channelID string, msg *events.Message) {
	// Ignore messages from self
	if msg.Info.IsFromMe {
		return
	}

	// STRICT FILTER: Only handle 1-on-1 private chats (DM).
	// Ignore groups (@g.us), status/broadcast (@broadcast), newsletters/channels (@newsletter), etc.
	if msg.Info.IsGroup || msg.Info.IsIncomingBroadcast() || msg.Info.IsNewsletterStatus {
		return
	}
	server := msg.Info.Chat.Server
	if server != types.DefaultUserServer && server != "lid" {
		return
	}

	text := extractText(msg)
	mediaInfo := h.processAttachments(ctx, channelID, msg)
	if mediaInfo != "" {
		if strings.TrimSpace(text) == "" {
			text = mediaInfo
		} else {
			text = text + "\n\n" + mediaInfo
		}
	}

	if strings.TrimSpace(text) == "" {
		return
	}

	// Resolve canonical phone number (decode LID if applicable)
	userJID := msg.Info.Sender
	if userJID.IsEmpty() {
		userJID = msg.Info.Chat
	}
	if !msg.Info.SenderAlt.IsEmpty() && msg.Info.SenderAlt.Server == types.DefaultUserServer {
		userJID = msg.Info.SenderAlt
	} else if h.sender != nil {
		userJID = h.sender.ResolvePhoneNumber(ctx, channelID, userJID)
	}

	senderUser := userJID.User
	if senderUser == "" {
		senderUser = msg.Info.Sender.User
	}
	if senderUser == "" {
		senderUser = msg.Info.Chat.User
	}

	payload := IncomingMessage{
		ChannelID:         channelID,
		ExternalMessageID: msg.Info.ID,
		ExternalUserID:    senderUser,
		MessageType:       "text",
		Content: MessageContent{
			Text: text,
		},
		Timestamp: msg.Info.Timestamp,
	}

	// Immediately show WhatsApp typing status (sedang mengetik...) to the customer
	if h.sender != nil {
		_ = h.sender.SendPresence(ctx, channelID, msg.Info.Chat, types.ChatPresenceComposing)
	}

	var resp BackendResponse[IncomingMessageResult]
	if err := h.client.PostWithResponse(ctx, "/internal/v1/messages/incoming", payload, &resp); err != nil {
		if h.sender != nil {
			_ = h.sender.SendPresence(ctx, channelID, msg.Info.Chat, types.ChatPresencePaused)
		}
		h.log.Error("callback: incoming message failed",
			"channel_id", channelID,
			"msg_id", msg.Info.ID,
			"error", err,
		)
		return
	}

	// PostWithResponse already guarantees a successful 2xx response. The Go
	// backend's canonical envelope does not include a top-level `success` flag.
	if resp.Data.Content != "" && !resp.Data.HandoffRequested && h.sender != nil {
		participant := msg.Info.Sender.ToNonAD().String()
		fullContent := strings.TrimSpace(resp.Data.Content)

		// Split response into natural multi-bubble chats (by --- or paragraphs)
		chunks := splitMessageChunks(fullContent)

		for i, chunk := range chunks {
			chunk = strings.TrimSpace(chunk)
			if chunk == "" {
				continue
			}

			// Keep typing presence active during delay
			_ = h.sender.SendPresence(ctx, channelID, msg.Info.Chat, types.ChatPresenceComposing)

			// Add human typing delay between bubbles
			typingDelay := time.Duration(len(chunk)*20) * time.Millisecond
			if typingDelay < 700*time.Millisecond {
				typingDelay = 700 * time.Millisecond
			} else if typingDelay > 2400*time.Millisecond {
				typingDelay = 2400 * time.Millisecond
			}
			time.Sleep(typingDelay)

			quoteID := ""
			var quoteMsg *waE2E.Message
			quoteParticipant := ""
			if i == 0 {
				quoteID = msg.Info.ID
				quoteMsg = msg.Message
				quoteParticipant = participant
			}

			if err := h.sender.SendText(ctx, channelID, msg.Info.Chat, chunk, quoteID, quoteParticipant, quoteMsg); err != nil {
				h.log.Error("send reply to whatsapp failed",
					"channel_id", channelID,
					"chat_jid", msg.Info.Chat,
					"error", err,
				)
			}
		}

		// Pause typing presence when finished
		_ = h.sender.SendPresence(ctx, channelID, msg.Info.Chat, types.ChatPresencePaused)

		h.log.Info("reply sent to whatsapp",
			"channel_id", channelID,
			"chat_jid", msg.Info.Chat,
			"msg_id", resp.Data.MessageID,
			"bubbles_count", len(chunks),
		)
	} else if h.sender != nil {
		_ = h.sender.SendPresence(ctx, channelID, msg.Info.Chat, types.ChatPresencePaused)
	}
}

func splitMessageChunks(text string) []string {
	if strings.Contains(text, "---") {
		parts := strings.Split(text, "---")
		var valid []string
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				valid = append(valid, trimmed)
			}
		}
		if len(valid) > 1 {
			return valid
		}
	}

	// Split by double newlines into natural paragraphs
	if strings.Contains(text, "\n\n") {
		parts := strings.Split(text, "\n\n")
		var valid []string
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				valid = append(valid, trimmed)
			}
		}
		if len(valid) > 1 && len(valid) <= 4 {
			return valid
		}
	}

	return []string{text}
}

func (h *Handler) processAttachments(ctx context.Context, channelID string, msg *events.Message) string {
	if h.sender == nil || msg.Message == nil {
		return ""
	}

	mediaDir := "/tmp/chatsolv-media"
	if err := os.MkdirAll(mediaDir, 0700); err != nil {
		h.log.Error("create media directory", "error", err)
		return ""
	}

	m := msg.Message
	// 1. Image / Screenshot
	if img := m.GetImageMessage(); img != nil {
		data, err := h.sender.DownloadAttachment(ctx, channelID, img)
		if err == nil && validAttachmentSize(data) {
			fileName, nameErr := randomMediaName("img", ".jpg")
			if nameErr != nil {
				return ""
			}
			filePath := filepath.Join(mediaDir, fileName)
			if os.WriteFile(filePath, data, 0600) != nil {
				return ""
			}
			caption := ""
			if img.Caption != nil && *img.Caption != "" {
				caption = fmt.Sprintf(" (caption: \"%s\")", *img.Caption)
			}
			return fmt.Sprintf("[MEDIA_IMAGE: %s]%s\n[File tersimpan di: %s]", fileName, caption, filePath)
		}
	}

	// 2. Document / PDF / TXT
	if doc := m.GetDocumentMessage(); doc != nil {
		data, err := h.sender.DownloadAttachment(ctx, channelID, doc)
		if err == nil && validAttachmentSize(data) {
			origName := "document"
			if doc.FileName != nil && *doc.FileName != "" {
				origName = filepath.Base(*doc.FileName)
			}
			ext := strings.ToLower(filepath.Ext(origName))
			if len(ext) > 10 {
				ext = ""
			}
			fileName, nameErr := randomMediaName("doc", ext)
			if nameErr != nil {
				return ""
			}
			filePath := filepath.Join(mediaDir, fileName)
			if os.WriteFile(filePath, data, 0600) != nil {
				return ""
			}

			mime := ""
			if doc.Mimetype != nil {
				mime = *doc.Mimetype
			}
			if (strings.HasPrefix(mime, "text/") || strings.HasSuffix(origName, ".txt") || strings.HasSuffix(origName, ".json") || strings.HasSuffix(origName, ".csv") || strings.HasSuffix(origName, ".md")) && len(data) < 500000 {
				return fmt.Sprintf("[MEDIA_DOC: %s|%s]\n[File tersimpan di: %s]\n--- ISI DOKUMEN ---\n%s\n--- AKHIR DOKUMEN ---", fileName, origName, filePath, string(data))
			}
			return fmt.Sprintf("[MEDIA_DOC: %s|%s]\n[File tersimpan di: %s]", fileName, origName, filePath)
		}
	}

	// 3. Audio / Voice Note
	if aud := m.GetAudioMessage(); aud != nil {
		data, err := h.sender.DownloadAttachment(ctx, channelID, aud)
		if err == nil && validAttachmentSize(data) {
			fileName, nameErr := randomMediaName("aud", ".ogg")
			if nameErr != nil {
				return ""
			}
			filePath := filepath.Join(mediaDir, fileName)
			if os.WriteFile(filePath, data, 0600) != nil {
				return ""
			}
			return fmt.Sprintf("[MEDIA_AUDIO: %s]\n[Pesan Suara/Audio tersimpan di: %s]", fileName, filePath)
		}
	}

	// 4. Sticker
	if stk := m.GetStickerMessage(); stk != nil {
		data, err := h.sender.DownloadAttachment(ctx, channelID, stk)
		if err == nil && validAttachmentSize(data) {
			fileName, nameErr := randomMediaName("stk", ".webp")
			if nameErr != nil {
				return ""
			}
			filePath := filepath.Join(mediaDir, fileName)
			if os.WriteFile(filePath, data, 0600) != nil {
				return ""
			}
			return fmt.Sprintf("[MEDIA_STICKER: %s]\n[Stiker tersimpan di: %s]", fileName, filePath)
		}
	}

	// 5. Video
	if vid := m.GetVideoMessage(); vid != nil {
		data, err := h.sender.DownloadAttachment(ctx, channelID, vid)
		if err == nil && validAttachmentSize(data) {
			fileName, nameErr := randomMediaName("vid", ".mp4")
			if nameErr != nil {
				return ""
			}
			filePath := filepath.Join(mediaDir, fileName)
			if os.WriteFile(filePath, data, 0600) != nil {
				return ""
			}
			caption := ""
			if vid.Caption != nil && *vid.Caption != "" {
				caption = fmt.Sprintf(" (caption: \"%s\")", *vid.Caption)
			}
			return fmt.Sprintf("[MEDIA_VIDEO: %s]%s\n[Video tersimpan di: %s]", fileName, caption, filePath)
		}
	}

	return ""
}

const maxAttachmentBytes = 20 << 20

func validAttachmentSize(data []byte) bool {
	return len(data) > 0 && len(data) <= maxAttachmentBytes
}

func randomMediaName(prefix, extension string) (string, error) {
	random := make([]byte, 24)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(random) + extension, nil
}

func (h *Handler) onStatus(ctx context.Context, channelID, status, phone, sessionID string) {
	payload := StatusPayload{
		ChannelID:   channelID,
		Status:      status,
		PhoneNumber: phone,
		SessionID:   sessionID,
	}
	if err := h.client.Post(ctx, "/internal/v1/channels/status", payload); err != nil {
		h.log.Error("callback: status update failed",
			"channel_id", channelID,
			"status", status,
			"error", err,
		)
	}
}

func (h *Handler) onEvent(ctx context.Context, channelID, event, phone, sessionID string) {
	payload := EventPayload{
		ChannelID:   channelID,
		Event:       event,
		PhoneNumber: phone,
		SessionID:   sessionID,
	}
	if err := h.client.Post(ctx, "/internal/v1/channels/events", payload); err != nil {
		h.log.Error("callback: channel event failed",
			"channel_id", channelID,
			"event", event,
			"error", err,
		)
	}
}

// extractText returns the plain-text body from a whatsmeow Message event.
func extractText(msg *events.Message) string {
	m := msg.Message
	if m == nil {
		return ""
	}
	switch {
	case m.Conversation != nil:
		return *m.Conversation
	case m.ExtendedTextMessage != nil && m.ExtendedTextMessage.Text != nil:
		return *m.ExtendedTextMessage.Text
	case m.ImageMessage != nil && m.ImageMessage.Caption != nil:
		return *m.ImageMessage.Caption
	case m.VideoMessage != nil && m.VideoMessage.Caption != nil:
		return *m.VideoMessage.Caption
	case m.DocumentMessage != nil && m.DocumentMessage.Caption != nil:
		return *m.DocumentMessage.Caption
	}
	return ""
}
