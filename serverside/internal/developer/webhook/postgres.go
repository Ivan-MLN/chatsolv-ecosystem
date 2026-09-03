package webhook

import (
	"authbackend/generated/sqlc"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(p *pgxpool.Pool) *PostgresRepository { return &PostgresRepository{p} }
func (r *PostgresRepository) Authorize(ctx context.Context, userID, workspaceID string) (string, bool, error) {
	uid, e := webhookUUID(userID)
	if e != nil {
		return "", false, e
	}
	wid, e := webhookUUID(workspaceID)
	if e != nil {
		return "", false, e
	}
	row, e := sqlc.New(r.pool).AuthorizeWebhookWorkspace(ctx, sqlc.AuthorizeWebhookWorkspaceParams{WorkspaceID: wid, UserID: uid})
	return row.Role, row.Webhooks, e
}
func (r *PostgresRepository) Create(ctx context.Context, v Endpoint) error {
	id, e := webhookUUID(v.ID)
	if e != nil {
		return e
	}
	wid, e := webhookUUID(v.WorkspaceID)
	if e != nil {
		return e
	}
	events, _ := json.Marshal(v.Events)
	_, e = sqlc.New(r.pool).CreateWebhookEndpoint(ctx, sqlc.CreateWebhookEndpointParams{ID: id, WorkspaceID: wid, Url: v.URL, Events: events, Status: v.Status, SecretCiphertext: v.SecretCiphertext})
	return e
}
func (r *PostgresRepository) List(ctx context.Context, workspaceID, userID string) ([]Endpoint, error) {
	wid, e := webhookUUID(workspaceID)
	if e != nil {
		return nil, e
	}
	uid, e := webhookUUID(userID)
	if e != nil {
		return nil, e
	}
	rows, e := sqlc.New(r.pool).ListWebhookEndpointsForManager(ctx, sqlc.ListWebhookEndpointsForManagerParams{WorkspaceID: wid, UserID: uid})
	if e != nil {
		return nil, e
	}
	result := make([]Endpoint, 0, len(rows))
	for _, row := range rows {
		result = append(result, endpoint(row))
	}
	return result, nil
}
func (r *PostgresRepository) Update(ctx context.Context, id, userID string, in UpdateInput) (Endpoint, error) {
	wid, e := webhookUUID(id)
	if e != nil {
		return Endpoint{}, e
	}
	uid, e := webhookUUID(userID)
	if e != nil {
		return Endpoint{}, e
	}
	events, _ := json.Marshal(in.Events)
	row, e := sqlc.New(r.pool).UpdateWebhookEndpointForManager(ctx, sqlc.UpdateWebhookEndpointForManagerParams{ID: wid, UserID: uid, Url: in.URL, Events: events, Status: in.Status})
	return endpoint(row), e
}
func (r *PostgresRepository) Delete(ctx context.Context, id, userID string) error {
	wid, e := webhookUUID(id)
	if e != nil {
		return e
	}
	uid, e := webhookUUID(userID)
	if e != nil {
		return e
	}
	_, e = sqlc.New(r.pool).DeleteWebhookEndpointForManager(ctx, sqlc.DeleteWebhookEndpointForManagerParams{ID: wid, UserID: uid})
	return e
}
func webhookUUID(v string) (pgtype.UUID, error) {
	id, e := uuid.Parse(v)
	return pgtype.UUID{Bytes: id, Valid: e == nil}, e
}
func endpoint(row sqlc.WebhookEndpoint) Endpoint {
	var events []string
	_ = json.Unmarshal(row.Events, &events)
	return Endpoint{ID: uuid.UUID(row.ID.Bytes).String(), WorkspaceID: uuid.UUID(row.WorkspaceID.Bytes).String(), URL: row.Url, Events: events, Status: row.Status, SecretCiphertext: row.SecretCiphertext, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
}
