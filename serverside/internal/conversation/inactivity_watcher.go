package conversation

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BotMessenger interface {
	SendTextMessage(ctx context.Context, channelID, recipient, text string) error
}

type InactivityWatcher struct {
	pool      *pgxpool.Pool
	messenger BotMessenger
	log       *slog.Logger
}

func NewInactivityWatcher(pool *pgxpool.Pool, messenger BotMessenger, log *slog.Logger) *InactivityWatcher {
	return &InactivityWatcher{
		pool:      pool,
		messenger: messenger,
		log:       log,
	}
}

// Start begins periodic background checks for 5-minute reminders and 10-minute session timeouts.
func (w *InactivityWatcher) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w.CheckInactivity(ctx)
			}
		}
	}()
}

type inactiveConversationRow struct {
	ID             pgtype.UUID
	WorkspaceID    pgtype.UUID
	ChannelID      pgtype.UUID
	ExternalUserID string
	LastMessageAt  pgtype.Timestamptz
	Metadata       []byte
	ChannelType    string
}

func (w *InactivityWatcher) CheckInactivity(ctx context.Context) {
	query := `
		SELECT c.id, c.workspace_id, c.channel_id, c.external_user_id, c.last_message_at, c.metadata, ch.type
		FROM conversations c
		JOIN channels ch ON ch.id = c.channel_id
		WHERE c.status = 'open'
		  AND c.environment = 'production'
		  AND ch.type = 'whatsapp'
		  AND ch.status = 'connected'
	`

	rows, err := w.pool.Query(ctx, query)
	if err != nil {
		w.log.Error("inactivity watcher query failed", "error", err)
		return
	}
	defer rows.Close()

	var toProcess []inactiveConversationRow
	for rows.Next() {
		var item inactiveConversationRow
		if err := rows.Scan(&item.ID, &item.WorkspaceID, &item.ChannelID, &item.ExternalUserID, &item.LastMessageAt, &item.Metadata, &item.ChannelType); err == nil {
			toProcess = append(toProcess, item)
		}
	}

	for _, item := range toProcess {
		if !item.LastMessageAt.Valid {
			continue
		}

		elapsed := time.Since(item.LastMessageAt.Time)
		convID := conversationID(item.ID)
		channelID := conversationID(item.ChannelID)

		var meta map[string]any
		if len(item.Metadata) > 0 {
			_ = json.Unmarshal(item.Metadata, &meta)
		}
		if meta == nil {
			meta = make(map[string]any)
		}

		// 1. 10-Minute Timeout -> Close session & send closing message
		if elapsed >= 10*time.Minute {
			w.log.Info("auto-closing inactive conversation (10m timeout)", "conversation_id", convID, "user_id", item.ExternalUserID)

			closingText := "Karena tidak ada aktivitas lebih lanjut, sesi percakapan ini kami tutup terlebih dahulu ya kak. Terima kasih sudah menghubungi kami! Kakak bisa chat kembali kapan saja untuk memulai sesi baru 🙏✨"

			// Mark closed in database
			meta["closed_reason"] = "inactivity_10m"
			meta["reminder_5m_sent"] = true
			metaBytes, _ := json.Marshal(meta)

			_, _ = w.pool.Exec(ctx, `
				UPDATE conversations
				SET status = 'closed', closed_at = now(), metadata = $2, updated_at = now()
				WHERE id = $1
			`, item.ID, metaBytes)

			// Record message in DB
			msgID := mustConversationUUID(uuid.NewString())
			contentJSON, _ := json.Marshal(map[string]string{"text": closingText})
			_, _ = w.pool.Exec(ctx, `
				INSERT INTO messages(id, workspace_id, conversation_id, channel_id, sender_type, content_type, content, provider, status)
				VALUES($1, $2, $3, $4, 'agent', 'text', $5, 'system_timeout', 'created')
			`, msgID, item.WorkspaceID, item.ID, item.ChannelID, contentJSON)

			// Send to customer via WhatsApp
			if w.messenger != nil {
				_ = w.messenger.SendTextMessage(ctx, channelID, item.ExternalUserID, closingText)
			}
			continue
		}

		// 2. 5-Minute Reminder -> Send follow-up reminder if not sent yet
		if elapsed >= 5*time.Minute {
			if sent, ok := meta["reminder_5m_sent"].(bool); ok && sent {
				continue
			}

			w.log.Info("sending 5-minute inactivity reminder", "conversation_id", convID, "user_id", item.ExternalUserID)

			reminderText := "Halo kak, apakah masih ada hal lain yang bisa kami bantu? Jangan ragu untuk chat kami kembali jika ada pertanyaan ya 😊"

			meta["reminder_5m_sent"] = true
			metaBytes, _ := json.Marshal(meta)

			_, _ = w.pool.Exec(ctx, `
				UPDATE conversations
				SET metadata = $2, updated_at = now()
				WHERE id = $1
			`, item.ID, metaBytes)

			// Record message in DB
			msgID := mustConversationUUID(uuid.NewString())
			contentJSON, _ := json.Marshal(map[string]string{"text": reminderText})
			_, _ = w.pool.Exec(ctx, `
				INSERT INTO messages(id, workspace_id, conversation_id, channel_id, sender_type, content_type, content, provider, status)
				VALUES($1, $2, $3, $4, 'agent', 'text', $5, 'system_reminder', 'created')
			`, msgID, item.WorkspaceID, item.ID, item.ChannelID, contentJSON)

			// Send to customer via WhatsApp
			if w.messenger != nil {
				_ = w.messenger.SendTextMessage(ctx, channelID, item.ExternalUserID, reminderText)
			}
		}
	}
}
