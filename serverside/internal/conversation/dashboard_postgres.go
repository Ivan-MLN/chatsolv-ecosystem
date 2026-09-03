package conversation

import (
	"context"
	"encoding/json"

	"authbackend/generated/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DashboardPostgresRepository struct{ pool *pgxpool.Pool }

func NewDashboardPostgresRepository(pool *pgxpool.Pool) *DashboardPostgresRepository {
	return &DashboardPostgresRepository{pool: pool}
}
func (r *DashboardPostgresRepository) List(ctx context.Context, workspaceID, userID string, filter ListFilter) ([]DashboardConversation, error) {
	wid, err := conversationUUID(workspaceID)
	if err != nil {
		return nil, err
	}
	uid, err := conversationUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(r.pool).ListConversationsForMember(ctx, sqlc.ListConversationsForMemberParams{WorkspaceID: wid, UserID: uid, Column3: filter.Status, Column4: filter.Mode, Limit: int32(filter.Limit)})
	if err != nil {
		return nil, err
	}
	result := make([]DashboardConversation, 0, len(rows))
	for _, row := range rows {
		result = append(result, dashboardConversation(row))
	}
	return result, nil
}
func (r *DashboardPostgresRepository) Get(ctx context.Context, id, userID string) (DashboardConversation, error) {
	cid, err := conversationUUID(id)
	if err != nil {
		return DashboardConversation{}, err
	}
	uid, err := conversationUUID(userID)
	if err != nil {
		return DashboardConversation{}, err
	}
	row, err := sqlc.New(r.pool).GetConversationForMember(ctx, sqlc.GetConversationForMemberParams{ID: cid, UserID: uid})
	return dashboardConversation(row), err
}
func (r *DashboardPostgresRepository) Messages(ctx context.Context, id, userID string, filter MessageFilter) ([]DashboardMessage, error) {
	cid, err := conversationUUID(id)
	if err != nil {
		return nil, err
	}
	uid, err := conversationUUID(userID)
	if err != nil {
		return nil, err
	}
	cursor := pgtype.Timestamptz{}
	if filter.Cursor != nil {
		cursor = pgtype.Timestamptz{Time: *filter.Cursor, Valid: true}
	}
	rows, err := sqlc.New(r.pool).ListMessagesForMember(ctx, sqlc.ListMessagesForMemberParams{ConversationID: cid, UserID: uid, Column3: cursor, Limit: int32(filter.Limit)})
	if err != nil {
		return nil, err
	}
	result := make([]DashboardMessage, 0, len(rows))
	for _, row := range rows {
		var body struct {
			Text string `json:"text"`
		}
		_ = json.Unmarshal(row.Content, &body)
		result = append(result, DashboardMessage{ID: conversationID(row.ID), ConversationID: conversationID(row.ConversationID), SenderType: row.SenderType, ContentType: row.ContentType, Content: body.Text, Status: row.Status, CreatedAt: row.CreatedAt.Time})
	}
	return result, nil
}
func (r *DashboardPostgresRepository) Role(ctx context.Context, id, userID string) (string, error) {
	cid, err := conversationUUID(id)
	if err != nil {
		return "", err
	}
	uid, err := conversationUUID(userID)
	if err != nil {
		return "", err
	}
	return sqlc.New(r.pool).GetConversationManagerRole(ctx, sqlc.GetConversationManagerRoleParams{ID: cid, UserID: uid})
}
func (r *DashboardPostgresRepository) SetMode(ctx context.Context, id, userID string, mode Mode) error {
	cid, err := conversationUUID(id)
	if err != nil {
		return err
	}
	uid, err := conversationUUID(userID)
	if err != nil {
		return err
	}
	_, err = sqlc.New(r.pool).SetConversationModeForManager(ctx, sqlc.SetConversationModeForManagerParams{
		ID:     cid,
		UserID: uid,
		Mode:   string(mode),
	})
	return err
}
func dashboardConversation(row sqlc.Conversation) DashboardConversation {
	var assigned *string
	if row.AssignedUserID.Valid {
		value := uuid.UUID(row.AssignedUserID.Bytes).String()
		assigned = &value
	}
	return DashboardConversation{ID: conversationID(row.ID), WorkspaceID: conversationID(row.WorkspaceID), AgentID: conversationID(row.AgentID), ChannelID: conversationID(row.ChannelID), ExternalUserID: row.ExternalUserID, Status: row.Status, Mode: row.Mode, Environment: row.Environment, AssignedUserID: assigned, StartedAt: row.StartedAt.Time, LastMessageAt: row.LastMessageAt.Time, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
}
