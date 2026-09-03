package conversation

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RelayCommandType string

const (
	RelayCmdNone   RelayCommandType = "NONE"
	RelayCmdAccept RelayCommandType = "ACCEPT"
	RelayCmdDone   RelayCommandType = "DONE"
	RelayCmdRelay  RelayCommandType = "RELAY"
)

type RelayCommand struct {
	Type           RelayCommandType
	ConversationID string
	Text           string
}

// ParseRelayCommand parses admin text commands (#ACC, #DONE, #CLOSE, atau pesan langsung).
// Jika hasActiveSession=true dan bukan command control (#ACC / #DONE / #CLOSE), semua pesan
// langsung dianggap sebagai RelayCmdRelay tanpa perlu prefix '#'.
func ParseRelayCommand(text string, hasActiveSession bool) RelayCommand {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return RelayCommand{Type: RelayCmdNone}
	}

	upper := strings.ToUpper(trimmed)
	if strings.HasPrefix(upper, "#ACC") {
		parts := strings.Fields(trimmed)
		convID := ""
		if len(parts) > 1 {
			convID = strings.TrimPrefix(parts[1], "#")
			convID = strings.TrimPrefix(convID, "CNV-")
			convID = strings.TrimPrefix(convID, "cnv-")
		}
		return RelayCommand{
			Type:           RelayCmdAccept,
			ConversationID: convID,
		}
	}

	if upper == "#DONE" || upper == "#CLOSE" || upper == "#SELESAI" || upper == "#END" {
		return RelayCommand{
			Type: RelayCmdDone,
		}
	}

	// Jika admin memiliki sesi relay aktif, semua pesan (teks/media caption) langsung di-relay
	if hasActiveSession {
		content := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
		if content == "" {
			content = trimmed
		}
		return RelayCommand{
			Type: RelayCmdRelay,
			Text: content,
		}
	}

	// Fallback jika tidak ada sesi aktif tapi pakai prefix '#' eksplisit
	if strings.HasPrefix(trimmed, "#") {
		content := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
		if content != "" {
			return RelayCommand{
				Type: RelayCmdRelay,
				Text: content,
			}
		}
	}

	return RelayCommand{Type: RelayCmdNone}
}

// FormatEscalationBroadcast formats notification text sent to admin WhatsApp
func FormatEscalationBroadcast(customerPhone, conversationID, lastMessage string) string {
	shortID := conversationID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	return fmt.Sprintf("⚠️ [ESKALASI PERCAKAPAN]\nPelanggan: %s\nPercakapan ID: #CNV-%s\nPesan Terakhir: \"%s\"\n\nKetik #ACC %s untuk mengambil alih percakapan ini.", customerPhone, shortID, lastMessage, shortID)
}

// FormatForwardToAdmin formats incoming customer message forwarded to admin WhatsApp
func FormatForwardToAdmin(customerPhone, messageText string) string {
	return fmt.Sprintf("📩 [PESAN DARI PELANGGAN: %s]\n%s\n\nKetik balasan Anda langsung (atau #DONE jika selesai).", customerPhone, messageText)
}

// User Notification Templates
const (
	CustomerConnectedNotification = "Halo kak, saat ini Anda sudah terhubung langsung dengan Customer Support kami. Silakan sampaikan pesan Anda 🙏"
	CustomerClosedNotification    = "Sesi konsultasi dengan Customer Support telah selesai. Terima kasih telah menghubungi kami! Asisten AI kami siap membantu kembali jika ada pertanyaan lainnya 🙏"
)

// RelaySessionStore manages active admin-customer relay sessions in Redis
type RelaySessionStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRelaySessionStore(client *redis.Client, ttl time.Duration) *RelaySessionStore {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &RelaySessionStore{client: client, ttl: ttl}
}

func (s *RelaySessionStore) SetActiveRelay(ctx context.Context, adminPhone, conversationID string) error {
	adminKey := "relay:admin:" + adminPhone
	convKey := "relay:conversation:" + conversationID

	pipe := s.client.Pipeline()
	pipe.Set(ctx, adminKey, conversationID, s.ttl)
	pipe.Set(ctx, convKey, adminPhone, s.ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *RelaySessionStore) GetConversationForAdmin(ctx context.Context, adminPhone string) (string, error) {
	val, err := s.client.Get(ctx, "relay:admin:"+adminPhone).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}
	return val, nil
}

func (s *RelaySessionStore) GetAdminForConversation(ctx context.Context, conversationID string) (string, error) {
	val, err := s.client.Get(ctx, "relay:conversation:"+conversationID).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}
	return val, nil
}

func (s *RelaySessionStore) ClearRelay(ctx context.Context, adminPhone string) error {
	convID, err := s.GetConversationForAdmin(ctx, adminPhone)
	if err != nil || convID == "" {
		return err
	}

	adminKey := "relay:admin:" + adminPhone
	convKey := "relay:conversation:" + convID
	return s.client.Del(ctx, adminKey, convKey).Err()
}
