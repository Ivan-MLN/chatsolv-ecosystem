package publicapi

import (
	"context"
	"encoding/json"
	"errors"

	"authbackend/generated/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, session Session) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := sqlc.New(tx)
	workspaceID, err := publicUUID(session.WorkspaceID)
	if err != nil {
		return err
	}
	agent, err := q.FindDefaultPublicAgent(ctx, workspaceID)
	if err != nil {
		return err
	}
	channelID, err := q.FindPublicWebChannel(ctx, sqlc.FindPublicWebChannelParams{WorkspaceID: workspaceID, AgentID: agent.ID})
	if errors.Is(err, pgx.ErrNoRows) {
		channelID, err = q.CreatePublicWebChannel(ctx, sqlc.CreatePublicWebChannelParams{ID: mustPublicUUID(uuid.NewString()), WorkspaceID: workspaceID, AgentID: agent.ID})
	}
	if err != nil {
		return err
	}
	metadata, err := json.Marshal(session.Metadata)
	if err != nil {
		return err
	}
	_, err = q.CreatePublicAPISession(ctx, sqlc.CreatePublicAPISessionParams{ID: mustPublicUUID(session.ID), WorkspaceID: workspaceID, AgentID: agent.ID, ExternalUserID: session.ExternalUserID, TokenHash: session.TokenHash, Metadata: metadata, ExpiresAt: pgtype.Timestamptz{Time: session.ExpiresAt, Valid: true}})
	if err != nil {
		return err
	}
	_ = channelID
	return tx.Commit(ctx)
}

func (r *PostgresRepository) Resolve(ctx context.Context, sessionID, hash string) (Session, error) {
	id, err := publicUUID(sessionID)
	if err != nil {
		return Session{}, err
	}
	row, err := sqlc.New(r.pool).ResolvePublicAPISession(ctx, sqlc.ResolvePublicAPISessionParams{ID: id, TokenHash: hash})
	if err != nil {
		return Session{}, err
	}
	var metadata map[string]any
	_ = json.Unmarshal(row.Metadata, &metadata)
	return Session{ID: uuidString(row.ID), WorkspaceID: uuidString(row.WorkspaceID), AgentID: uuidString(row.AgentID), ChannelID: uuidString(row.ChannelID), ExternalUserID: row.ExternalUserID, TokenHash: row.TokenHash, Metadata: metadata, ExpiresAt: row.ExpiresAt.Time}, nil
}

func publicUUID(value string) (pgtype.UUID, error) {
	id, err := uuid.Parse(value)
	return pgtype.UUID{Bytes: id, Valid: err == nil}, err
}
func mustPublicUUID(value string) pgtype.UUID { id, _ := publicUUID(value); return id }
func uuidString(value pgtype.UUID) string     { return uuid.UUID(value.Bytes).String() }
