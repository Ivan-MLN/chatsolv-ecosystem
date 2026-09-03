package apikey

import (
	"context"
	"encoding/json"
	"time"

	"authbackend/generated/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}
func (r *PostgresRepository) Create(ctx context.Context, record Record) error {
	id, err := keyUUID(record.ID)
	if err != nil {
		return err
	}
	wid, err := keyUUID(record.WorkspaceID)
	if err != nil {
		return err
	}
	scopes, _ := json.Marshal(record.Scopes)
	_, err = sqlc.New(r.pool).CreateAPIKey(ctx, sqlc.CreateAPIKeyParams{ID: id, WorkspaceID: wid, Prefix: record.Prefix, Hash: record.Hash, LastFour: record.LastFour, Name: record.Name, Scopes: scopes})
	return err
}
func (r *PostgresRepository) FindByPrefix(ctx context.Context, prefix string) (Record, error) {
	row, err := sqlc.New(r.pool).FindAPIKeyByPrefix(ctx, prefix)
	return keyRecord(row), err
}
func (r *PostgresRepository) AuthorizeWorkspace(ctx context.Context, userID, workspaceID string) (string, error) {
	uid, err := keyUUID(userID)
	if err != nil {
		return "", err
	}
	wid, err := keyUUID(workspaceID)
	if err != nil {
		return "", err
	}
	return sqlc.New(r.pool).AuthorizeAPIKeyWorkspace(ctx, sqlc.AuthorizeAPIKeyWorkspaceParams{WorkspaceID: wid, UserID: uid})
}
func (r *PostgresRepository) List(ctx context.Context, workspaceID, userID string) ([]Record, error) {
	wid, err := keyUUID(workspaceID)
	if err != nil {
		return nil, err
	}
	uid, err := keyUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(r.pool).ListAPIKeysForManager(ctx, sqlc.ListAPIKeysForManagerParams{WorkspaceID: wid, UserID: uid})
	if err != nil {
		return nil, err
	}
	result := make([]Record, 0, len(rows))
	for _, row := range rows {
		result = append(result, keyRecord(row))
	}
	return result, nil
}
func (r *PostgresRepository) Revoke(ctx context.Context, keyID, userID string) error {
	id, err := keyUUID(keyID)
	if err != nil {
		return err
	}
	uid, err := keyUUID(userID)
	if err != nil {
		return err
	}
	_, err = sqlc.New(r.pool).RevokeAPIKeyForManager(ctx, sqlc.RevokeAPIKeyForManagerParams{ID: id, UserID: uid})
	return err
}
func keyUUID(value string) (pgtype.UUID, error) {
	id, err := uuid.Parse(value)
	return pgtype.UUID{Bytes: id, Valid: err == nil}, err
}
func keyRecord(row sqlc.ApiKey) Record {
	var scopes []string
	_ = json.Unmarshal(row.Scopes, &scopes)
	var revoked *time.Time
	if row.RevokedAt.Valid {
		value := row.RevokedAt.Time
		revoked = &value
	}
	return Record{ID: uuid.UUID(row.ID.Bytes).String(), WorkspaceID: uuid.UUID(row.WorkspaceID.Bytes).String(), Prefix: row.Prefix, Hash: row.Hash, LastFour: row.LastFour, Name: row.Name, Scopes: scopes, CreatedAt: row.CreatedAt.Time, RevokedAt: revoked}
}
