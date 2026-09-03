package internalapi

import (
	"context"

	"authbackend/generated/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository { return &PostgresRepository{pool} }
func (r *PostgresRepository) UpdateChannelStatus(ctx context.Context, channelID, status, phone, session string) error {
	id, err := internalUUID(channelID)
	if err != nil {
		return err
	}
	return sqlc.New(r.pool).UpdateChannelStatusInternal(ctx, sqlc.UpdateChannelStatusInternalParams{ID: id, Status: status, PhoneNumber: internalText(phone), ServiceInstanceID: internalText(session)})
}
func (r *PostgresRepository) ResolveRuntimeContext(ctx context.Context, agentID, conversationID string) (RuntimeContext, error) {
	aid, err := internalUUID(agentID)
	if err != nil {
		return RuntimeContext{}, err
	}
	cid, err := internalUUID(conversationID)
	if err != nil {
		return RuntimeContext{}, err
	}
	row, err := sqlc.New(r.pool).ResolveInternalRuntimeContext(ctx, sqlc.ResolveInternalRuntimeContextParams{ID: cid, AgentID: aid})
	return RuntimeContext{ChannelID: internalID(row.ChannelID), ChannelType: row.ChannelType, ExternalUserID: row.ExternalUserID}, err
}
func (r *PostgresRepository) GetAgentHealth(ctx context.Context, agentID string) (AgentHealth, error) {
	aid, err := internalUUID(agentID)
	if err != nil {
		return AgentHealth{}, err
	}
	row, err := sqlc.New(r.pool).GetInternalAgentHealth(ctx, aid)
	return AgentHealth{AgentID: internalID(row.ID), Status: row.Status, BrainStatus: row.BrainStatus, Ready: row.Ready.Bool}, err
}
func internalUUID(v string) (pgtype.UUID, error) {
	id, err := uuid.Parse(v)
	return pgtype.UUID{Bytes: id, Valid: err == nil}, err
}
func internalID(v pgtype.UUID) string   { return uuid.UUID(v.Bytes).String() }
func internalText(v string) pgtype.Text { return pgtype.Text{String: v, Valid: v != ""} }
