package channel

import (
	"context"
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
func (r *PostgresRepository) Authorize(ctx context.Context, userID, workspaceID string) (string, string, error) {
	uid, err := channelUUID(userID)
	if err != nil {
		return "", "", ErrInvalidInput
	}
	wid, err := channelUUID(workspaceID)
	if err != nil {
		return "", "", ErrInvalidInput
	}
	row, err := sqlc.New(r.pool).AuthorizeChannelWorkspace(ctx, sqlc.AuthorizeChannelWorkspaceParams{WorkspaceID: wid, UserID: uid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", ErrNotFound
		}
		return "", "", err
	}
	return row.Role, channelID(row.AgentID), nil
}
func (r *PostgresRepository) AuthorizeMutation(ctx context.Context, userID, channelIDValue string) error {
	uid, err := channelUUID(userID)
	if err != nil {
		return ErrInvalidInput
	}
	id, err := channelUUID(channelIDValue)
	if err != nil {
		return ErrInvalidInput
	}
	_, err = sqlc.New(r.pool).AuthorizeChannelMutation(ctx, sqlc.AuthorizeChannelMutationParams{ID: id, UserID: uid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
func (r *PostgresRepository) Count(ctx context.Context, workspaceID string) (int64, error) {
	wid, err := channelUUID(workspaceID)
	if err != nil {
		return 0, err
	}
	return sqlc.New(r.pool).CountWorkspaceChannels(ctx, wid)
}
func (r *PostgresRepository) Max(ctx context.Context, workspaceID string) (int64, error) {
	wid, err := channelUUID(workspaceID)
	if err != nil {
		return 0, err
	}
	maximum, err := sqlc.New(r.pool).GetMaxChannels(ctx, wid)
	return int64(maximum), err
}
func (r *PostgresRepository) Create(ctx context.Context, value Channel) (Channel, error) {
	id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	wid, err := channelUUID(value.WorkspaceID)
	if err != nil {
		return Channel{}, err
	}
	aid, err := channelUUID(value.AgentID)
	if err != nil {
		return Channel{}, err
	}
	row, err := sqlc.New(r.pool).CreateChannel(ctx, sqlc.CreateChannelParams{ID: id, WorkspaceID: wid, AgentID: aid, Type: value.Type, DisplayName: value.DisplayName, Status: value.Status})
	return fromDB(row), err
}
func (r *PostgresRepository) List(ctx context.Context, workspaceID, userID string) ([]Channel, error) {
	wid, err := channelUUID(workspaceID)
	if err != nil {
		return nil, err
	}
	uid, err := channelUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(r.pool).ListChannelsForMember(ctx, sqlc.ListChannelsForMemberParams{WorkspaceID: wid, UserID: uid})
	if err != nil {
		return nil, err
	}
	result := make([]Channel, 0, len(rows))
	for _, row := range rows {
		result = append(result, fromDB(row))
	}
	return result, nil
}
func (r *PostgresRepository) UpdateStatus(ctx context.Context, channelIDValue, status, phone, session string) error {
	id, err := channelUUID(channelIDValue)
	if err != nil {
		return err
	}
	return sqlc.New(r.pool).UpdateChannelStatusInternal(ctx, sqlc.UpdateChannelStatusInternalParams{ID: id, Status: status, PhoneNumber: nullableText(phone), ServiceInstanceID: nullableText(session)})
}
func (r *PostgresRepository) Delete(ctx context.Context, channelIDValue, userID string) error {
	id, err := channelUUID(channelIDValue)
	if err != nil {
		return err
	}
	uid, err := channelUUID(userID)
	if err != nil {
		return err
	}
	// Soft-hide all conversations for this channel from the UI so conversations tab looks fresh,
	// while keeping full message history and second brain references intact.
	_, _ = r.pool.Exec(ctx, "UPDATE conversations SET hidden_at=now(), status='closed' WHERE channel_id=$1", id)
	return sqlc.New(r.pool).DeleteChannelForMember(ctx, sqlc.DeleteChannelForMemberParams{ID: id, UserID: uid})
}
func channelUUID(value string) (pgtype.UUID, error) {
	id, err := uuid.Parse(value)
	return pgtype.UUID{Bytes: id, Valid: err == nil}, err
}
func channelID(value pgtype.UUID) string    { return uuid.UUID(value.Bytes).String() }
func nullableText(value string) pgtype.Text { return pgtype.Text{String: value, Valid: value != ""} }
func fromDB(row sqlc.Channel) Channel {
	return Channel{ID: channelID(row.ID), WorkspaceID: channelID(row.WorkspaceID), AgentID: channelID(row.AgentID), Type: row.Type, DisplayName: row.DisplayName, PhoneNumber: row.PhoneNumber.String, Status: row.Status, ServiceInstanceID: row.ServiceInstanceID.String}
}
