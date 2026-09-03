package conversation

import (
	"authbackend/generated/sqlc"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository { return &PostgresRepository{pool} }
func (r *PostgresRepository) FindResult(ctx context.Context, channelID, externalID string) (*Result, error) {
	cid, err := conversationUUID(channelID)
	if err != nil {
		return nil, err
	}
	row, err := sqlc.New(r.pool).FindIncomingResult(ctx, sqlc.FindIncomingResultParams{ChannelID: cid, ExternalMessageID: ctext(externalID)})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var body struct {
		Text string `json:"text"`
	}
	_ = json.Unmarshal(row.Content, &body)
	return &Result{MessageID: conversationID(row.ID), ConversationID: conversationID(row.ConversationID), Content: body.Text}, nil
}
func (r *PostgresRepository) ResolveOrCreate(ctx context.Context, in Incoming) (Conversation, error) {
	cid, err := conversationUUID(in.ChannelID)
	if err != nil {
		return Conversation{}, err
	}
	q := sqlc.New(r.pool)
	runtime, err := q.ResolveChannelRuntime(ctx, sqlc.ResolveChannelRuntimeParams{ID: cid, Type: in.ChannelType})
	if err != nil {
		return Conversation{}, err
	}
	if runtime.ChannelStatus != "connected" || runtime.AgentStatus != "ready" || !runtime.ProviderAgentID.Valid || !runtime.VaultKey.Valid {
		return Conversation{}, ErrRuntimeDisabled
	}
	if runtime.SubscriptionStatus != "active" && runtime.SubscriptionStatus != "trialing" {
		return Conversation{}, ErrRuntimeDisabled
	}
	row, err := q.FindOpenConversation(ctx, sqlc.FindOpenConversationParams{ChannelID: cid, ExternalUserID: in.ExternalUserID})
	if err == nil {
		// Session Inactivity Policy:
		// If last message was more than 15 minutes ago, auto-close the previous session and start a new fresh session.
		if time.Since(row.LastMessageAt.Time) > 15*time.Minute {
			_, _ = r.pool.Exec(ctx, "UPDATE conversations SET status='closed', closed_at=now(), updated_at=now() WHERE id=$1", row.ID)
		} else {
			return conversationFromDB(row), nil
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return Conversation{}, err
	}
	id, _ := conversationUUID(uuid.NewString())
	row, err = q.CreateConversation(ctx, sqlc.CreateConversationParams{ID: id, WorkspaceID: runtime.WorkspaceID, AgentID: runtime.AgentID, ChannelID: cid, ExternalUserID: in.ExternalUserID, Environment: environment(in.Environment), Metadata: []byte(`{}`)})
	return conversationFromDB(row), err
}
func (r *PostgresRepository) SaveCustomer(ctx context.Context, c Conversation, in Incoming) error {
	wid, _ := conversationUUID(c.WorkspaceID)
	cid, _ := conversationUUID(c.ID)
	channelID, _ := conversationUUID(in.ChannelID)
	id, _ := conversationUUID(uuid.NewString())
	content, _ := json.Marshal(map[string]string{"text": in.Text})
	_, err := sqlc.New(r.pool).CreateMessage(ctx, sqlc.CreateMessageParams{ID: id, WorkspaceID: wid, ConversationID: cid, ChannelID: channelID, SenderType: "customer", ContentType: "text", Content: content, ExternalMessageID: ctext(in.ExternalMessageID), Provider: ctext(in.Provider), Status: "received"})
	if err != nil {
		return err
	}
	return sqlc.New(r.pool).UpdateConversationActivity(ctx, sqlc.UpdateConversationActivityParams{ID: cid, WorkspaceID: wid})
}
func (r *PostgresRepository) RecentMessages(ctx context.Context, conversationIDValue string, limit int) ([]Message, error) {
	cid, err := conversationUUID(conversationIDValue)
	if err != nil {
		return nil, err
	}
	conversation, err := r.getConversation(ctx, cid)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(r.pool).ListRecentMessages(ctx, sqlc.ListRecentMessagesParams{WorkspaceID: conversation.WorkspaceID, ConversationID: cid, Limit: int32(limit)})
	if err != nil {
		return nil, err
	}
	out := make([]Message, 0, len(rows))
	for index := len(rows) - 1; index >= 0; index-- {
		var body struct {
			Text string `json:"text"`
		}
		_ = json.Unmarshal(rows[index].Content, &body)
		out = append(out, Message{Sender: rows[index].SenderType, Content: body.Text})
	}
	return out, nil
}
func (r *PostgresRepository) SaveAgent(ctx context.Context, c Conversation, contentValue string) (Result, error) {
	wid, _ := conversationUUID(c.WorkspaceID)
	cid, _ := conversationUUID(c.ID)
	id, _ := conversationUUID(uuid.NewString())
	content, _ := json.Marshal(map[string]string{"text": contentValue})
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := sqlc.New(tx)
	if _, err = q.ReserveMonthlyMessage(ctx, sqlc.ReserveMonthlyMessageParams{
		WorkspaceID: wid,
		ID:          mustConversationUUID(uuid.NewString()),
		Environment: c.Environment,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Result{}, ErrMessageLimitReached
		}
		return Result{}, err
	}
	row, err := q.CreateMessage(ctx, sqlc.CreateMessageParams{ID: id, WorkspaceID: wid, ConversationID: cid, SenderType: "agent", ContentType: "text", Content: content, Provider: ctext("hermes"), Status: "created"})
	if err != nil {
		return Result{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Result{}, err
	}
	return Result{MessageID: conversationID(row.ID), ConversationID: c.ID, Content: contentValue}, nil
}
func (r *PostgresRepository) RequestHandoff(ctx context.Context, c Conversation, _ string) error {
	id, err := conversationUUID(c.ID)
	if err != nil {
		return err
	}
	wid, err := conversationUUID(c.WorkspaceID)
	if err != nil {
		return err
	}
	return sqlc.New(r.pool).SetConversationMode(ctx, sqlc.SetConversationModeParams{ID: id, WorkspaceID: wid, Mode: string(ModeHuman)})
}
func (r *PostgresRepository) getConversation(ctx context.Context, id pgtype.UUID) (sqlc.Conversation, error) {
	return sqlc.New(r.pool).GetConversationByID(ctx, id)
}
func conversationFromDB(row sqlc.Conversation) Conversation {
	return Conversation{ID: conversationID(row.ID), WorkspaceID: conversationID(row.WorkspaceID), AgentID: conversationID(row.AgentID), ChannelID: conversationID(row.ChannelID), Mode: Mode(row.Mode), Environment: row.Environment}
}
func conversationUUID(value string) (pgtype.UUID, error) {
	id, err := uuid.Parse(value)
	return pgtype.UUID{Bytes: id, Valid: err == nil}, err
}
func mustConversationUUID(value string) pgtype.UUID { id, _ := conversationUUID(value); return id }
func conversationID(value pgtype.UUID) string       { return uuid.UUID(value.Bytes).String() }
func ctext(value string) pgtype.Text                { return pgtype.Text{String: value, Valid: value != ""} }
func environment(value string) string {
	if value == "test" {
		return "test"
	}
	return "production"
}
