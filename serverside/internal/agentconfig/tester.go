package agentconfig

import (
	"context"
	"encoding/json"

	"authbackend/generated/sqlc"
	"authbackend/internal/conversation"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Runtime interface {
	Generate(context.Context, conversation.RuntimeInput) (conversation.RuntimeOutput, error)
}

type PostgresAgentTester struct {
	pool    *pgxpool.Pool
	runtime Runtime
}

func NewPostgresAgentTester(pool *pgxpool.Pool, runtime Runtime) *PostgresAgentTester {
	return &PostgresAgentTester{pool: pool, runtime: runtime}
}

func (t *PostgresAgentTester) Test(ctx context.Context, input AgentTestInput) (AgentTestResult, error) {
	workspaceID, err := dbUUID(input.WorkspaceID)
	if err != nil {
		return AgentTestResult{}, err
	}
	agentID, err := dbUUID(input.AgentID)
	if err != nil {
		return AgentTestResult{}, err
	}
	channelUUID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("chatsolv-agent-test:"+input.WorkspaceID))
	channelID := pgtype.UUID{Bytes: channelUUID, Valid: true}
	q := sqlc.New(t.pool)
	if _, err = q.EnsureAgentTestChannel(ctx, sqlc.EnsureAgentTestChannelParams{ID: channelID, WorkspaceID: workspaceID, AgentID: agentID}); err != nil {
		return AgentTestResult{}, err
	}

	if input.Reset {
		// Close previous open test conversations for this user/channel
		_, _ = t.pool.Exec(ctx, "UPDATE conversations SET status='closed' WHERE channel_id=$1 AND status='open'", channelID)
		return AgentTestResult{Content: "Sesi percakapan testing berhasil di-reset. Anda dapat memulai percakapan baru."}, nil
	}

	var conversationID pgtype.UUID
	var isNewConversation bool

	if input.ConversationID != "" {
		if parsedCID, pErr := uuid.Parse(input.ConversationID); pErr == nil {
			var exists bool
			_ = t.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM conversations WHERE id=$1 AND channel_id=$2 AND status='open')", pgtype.UUID{Bytes: parsedCID, Valid: true}, channelID).Scan(&exists)
			if exists {
				conversationID = pgtype.UUID{Bytes: parsedCID, Valid: true}
			}
		}
	}

	if !conversationID.Valid {
		// Check for existing open test conversation
		var openID pgtype.UUID
		err = t.pool.QueryRow(ctx, "SELECT id FROM conversations WHERE channel_id=$1 AND status='open' ORDER BY last_message_at DESC LIMIT 1", channelID).Scan(&openID)
		if err == nil && openID.Valid {
			conversationID = openID
		} else {
			conversationID = pgtype.UUID{Bytes: uuid.New(), Valid: true}
			isNewConversation = true
			if _, err = q.CreateAgentTestConversation(ctx, sqlc.CreateAgentTestConversationParams{ID: conversationID, WorkspaceID: workspaceID, AgentID: agentID, ChannelID: channelID, ExternalUserID: "dashboard:" + input.UserID}); err != nil {
				return AgentTestResult{}, err
			}
		}
	}

	customerContent, _ := json.Marshal(map[string]string{"text": input.Message})
	if err = q.CreateAgentTestMessage(ctx, sqlc.CreateAgentTestMessageParams{ID: pgtype.UUID{Bytes: uuid.New(), Valid: true}, WorkspaceID: workspaceID, ConversationID: conversationID, ChannelID: channelID, SenderType: "customer", Content: customerContent, Provider: pgtype.Text{String: "dashboard", Valid: true}, Status: "received"}); err != nil {
		return AgentTestResult{}, err
	}
	_, _ = t.pool.Exec(ctx, "UPDATE conversations SET last_message_at=now(), updated_at=now() WHERE id=$1", conversationID)

	// Fetch conversation history for multi-turn testing context
	var history []conversation.Message
	if !isNewConversation {
		rows, hErr := sqlc.New(t.pool).ListRecentMessages(ctx, sqlc.ListRecentMessagesParams{WorkspaceID: workspaceID, ConversationID: conversationID, Limit: 30})
		if hErr == nil && len(rows) > 0 {
			for index := len(rows) - 1; index >= 0; index-- {
				var body struct {
					Text string `json:"text"`
				}
				_ = json.Unmarshal(rows[index].Content, &body)
				history = append(history, conversation.Message{Sender: rows[index].SenderType, Content: body.Text})
			}
		}
	}

	output, err := t.runtime.Generate(ctx, conversation.RuntimeInput{
		WorkspaceID:    input.WorkspaceID,
		AgentID:        input.AgentID,
		ConversationID: uuid.UUID(conversationID.Bytes).String(),
		Message:        input.Message,
		History:        history,
	})
	if err != nil {
		return AgentTestResult{}, err
	}
	agentContent, _ := json.Marshal(map[string]string{"text": output.Content})
	if err = q.CreateAgentTestMessage(ctx, sqlc.CreateAgentTestMessageParams{ID: pgtype.UUID{Bytes: uuid.New(), Valid: true}, WorkspaceID: workspaceID, ConversationID: conversationID, ChannelID: channelID, SenderType: "agent", Content: agentContent, Provider: pgtype.Text{String: "hermes", Valid: true}, Status: "created"}); err != nil {
		return AgentTestResult{}, err
	}
	return AgentTestResult{ConversationID: uuid.UUID(conversationID.Bytes).String(), Content: output.Content}, nil
}
