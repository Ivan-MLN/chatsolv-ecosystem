package provisioning

import (
	"authbackend/generated/sqlc"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository { return &PostgresRepository{pool} }
func (r *PostgresRepository) Get(ctx context.Context, workspaceID string) (Resource, error) {
	id, err := dbUUID(workspaceID)
	if err != nil {
		return Resource{}, err
	}
	row, err := sqlc.New(r.pool).GetProvisioningResource(ctx, id)
	if err != nil {
		return Resource{}, err
	}
	return Resource{WorkspaceID: uuidString(row.WorkspaceID), AgentID: uuidString(row.AgentID), SecondBrainID: uuidString(row.SecondBrainID), ProviderAgentID: row.ProviderAgentID, VaultKey: row.VaultKey}, nil
}
func (r *PostgresRepository) MarkProvisioning(ctx context.Context, workspaceID string) error {
	id, err := dbUUID(workspaceID)
	if err != nil {
		return err
	}
	q := sqlc.New(r.pool)
	if err = q.MarkAgentProvisioning(ctx, id); err != nil {
		return err
	}
	return q.MarkBrainProvisioning(ctx, id)
}
func (r *PostgresRepository) Complete(ctx context.Context, workspaceID, providerID, vaultKey string) error {
	id, err := dbUUID(workspaceID)
	if err != nil {
		return err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := sqlc.New(tx)
	if err = q.CompleteAgentProvisioning(ctx, sqlc.CompleteAgentProvisioningParams{WorkspaceID: id, ProviderAgentID: pgtype.Text{String: providerID, Valid: true}}); err != nil {
		return err
	}
	if err = q.CompleteBrainProvisioning(ctx, sqlc.CompleteBrainProvisioningParams{WorkspaceID: id, VaultKey: pgtype.Text{String: vaultKey, Valid: true}}); err != nil {
		return err
	}
	if err = q.ActivateWorkspace(ctx, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (r *PostgresRepository) Fail(ctx context.Context, workspaceID, code string) error {
	id, err := dbUUID(workspaceID)
	if err != nil {
		return err
	}
	q := sqlc.New(r.pool)
	if err = q.FailAgentProvisioning(ctx, id); err != nil {
		return err
	}
	if err = q.FailBrainProvisioning(ctx, id); err != nil {
		return err
	}
	return q.RecordProvisioningError(ctx, sqlc.RecordProvisioningErrorParams{WorkspaceID: id, LastErrorCode: pgtype.Text{String: code, Valid: true}})
}
func dbUUID(value string) (pgtype.UUID, error) {
	id, err := uuid.Parse(value)
	return pgtype.UUID{Bytes: id, Valid: err == nil}, err
}
func uuidString(value pgtype.UUID) string { return uuid.UUID(value.Bytes).String() }
