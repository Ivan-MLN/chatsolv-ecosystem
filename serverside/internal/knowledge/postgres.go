package knowledge

import (
	"authbackend/generated/sqlc"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository { return &PostgresRepository{pool} }
func (r *PostgresRepository) ResolveWorkspace(ctx context.Context, userID, workspaceID string) (WorkspaceAccess, error) {
	uid, err := knowledgeUUID(userID)
	if err != nil {
		return WorkspaceAccess{}, err
	}
	wid, err := knowledgeUUID(workspaceID)
	if err != nil {
		return WorkspaceAccess{}, err
	}
	row, err := sqlc.New(r.pool).ResolveKnowledgeWorkspace(ctx, sqlc.ResolveKnowledgeWorkspaceParams{WorkspaceID: wid, UserID: uid})
	if err != nil {
		return WorkspaceAccess{}, err
	}
	return WorkspaceAccess{Role: row.Role, SecondBrainID: knowledgeID(row.SecondBrainID), VaultKey: row.VaultKey}, nil
}
func (r *PostgresRepository) CreateSourceAndQueue(ctx context.Context, in SourceRecord, eventType string) (SourceRecord, error) {
	wid, err := knowledgeUUID(in.WorkspaceID)
	if err != nil {
		return SourceRecord{}, err
	}
	bid, err := knowledgeUUID(in.SecondBrainID)
	if err != nil {
		return SourceRecord{}, err
	}
	id, _ := knowledgeUUID(uuid.NewString())
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return SourceRecord{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := sqlc.New(tx)
	row, err := queries.CreateKnowledgeSource(ctx, sqlc.CreateKnowledgeSourceParams{ID: id, WorkspaceID: wid, SecondBrainID: bid, Type: in.Type, Title: in.Title, OriginalFilename: ktext(in.OriginalFilename), MimeType: ktext(in.MIMEType), FileSize: pgtype.Int8{Int64: in.FileSize, Valid: in.FileSize > 0}, OriginalObjectKey: ktext(in.ObjectKey), Checksum: ktext(in.Checksum)})
	if err != nil {
		return SourceRecord{}, mapKnowledgePersistenceError(err)
	}
	if err = queueKnowledgeEvent(ctx, queries, wid, id, eventType); err != nil {
		return SourceRecord{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return SourceRecord{}, mapKnowledgePersistenceError(err)
	}
	return sourceFromDB(row), nil
}
func (r *PostgresRepository) List(ctx context.Context, workspaceID string, limit, offset int) ([]SourceRecord, error) {
	wid, err := knowledgeUUID(workspaceID)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(r.pool).ListKnowledgeSources(ctx, sqlc.ListKnowledgeSourcesParams{WorkspaceID: wid, Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, err
	}
	out := make([]SourceRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, sourceFromDB(row))
	}
	return out, nil
}
func (r *PostgresRepository) GetForMember(ctx context.Context, userID, sourceID string) (SourceRecord, error) {
	uid, err := knowledgeUUID(userID)
	if err != nil {
		return SourceRecord{}, err
	}
	sid, err := knowledgeUUID(sourceID)
	if err != nil {
		return SourceRecord{}, err
	}
	row, err := sqlc.New(r.pool).GetKnowledgeSourceForMember(ctx, sqlc.GetKnowledgeSourceForMemberParams{ID: sid, UserID: uid})
	return sourceFromDB(row), err
}
func (r *PostgresRepository) UpdateAndQueue(ctx context.Context, userID, sourceID, title string) error {
	uid, sid, err := knowledgeMutationIDs(userID, sourceID)
	if err != nil {
		return err
	}
	return r.mutateAndQueue(ctx, sid, "knowledge.ingest", func(q *sqlc.Queries) (pgtype.UUID, error) {
		return q.UpdateKnowledgeTitleForMember(ctx, sqlc.UpdateKnowledgeTitleForMemberParams{ID: sid, UserID: uid, Title: title})
	})
}
func (r *PostgresRepository) DeleteAndQueue(ctx context.Context, userID, sourceID string) error {
	uid, sid, err := knowledgeMutationIDs(userID, sourceID)
	if err != nil {
		return err
	}
	return r.mutateAndQueue(ctx, sid, "knowledge.delete", func(q *sqlc.Queries) (pgtype.UUID, error) {
		return q.MarkKnowledgeDeletingForMember(ctx, sqlc.MarkKnowledgeDeletingForMemberParams{ID: sid, UserID: uid})
	})
}
func (r *PostgresRepository) RetryAndQueue(ctx context.Context, userID, sourceID string) error {
	uid, sid, err := knowledgeMutationIDs(userID, sourceID)
	if err != nil {
		return err
	}
	return r.mutateAndQueue(ctx, sid, "knowledge.ingest", func(q *sqlc.Queries) (pgtype.UUID, error) {
		return q.RetryKnowledgeSourceForMember(ctx, sqlc.RetryKnowledgeSourceForMemberParams{ID: sid, UserID: uid})
	})
}
func (r *PostgresRepository) mutateAndQueue(ctx context.Context, sourceID pgtype.UUID, eventType string, mutate func(*sqlc.Queries) (pgtype.UUID, error)) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := sqlc.New(tx)
	workspaceID, err := mutate(queries)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrForbidden
	}
	if err != nil {
		return err
	}
	if err = queueKnowledgeEvent(ctx, queries, workspaceID, sourceID, eventType); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (r *PostgresRepository) Count(ctx context.Context, workspaceID string) (int64, error) {
	wid, err := knowledgeUUID(workspaceID)
	if err != nil {
		return 0, err
	}
	return sqlc.New(r.pool).CountKnowledgeDocuments(ctx, wid)
}
func (r *PostgresRepository) StorageBytes(ctx context.Context, workspaceID string) (int64, error) {
	wid, err := knowledgeUUID(workspaceID)
	if err != nil {
		return 0, err
	}
	return sqlc.New(r.pool).GetKnowledgeStorageBytes(ctx, wid)
}
func sourceFromDB(row sqlc.KnowledgeSource) SourceRecord {
	return SourceRecord{ID: knowledgeID(row.ID), WorkspaceID: knowledgeID(row.WorkspaceID), SecondBrainID: knowledgeID(row.SecondBrainID), Type: row.Type, Title: row.Title, OriginalFilename: row.OriginalFilename.String, MIMEType: row.MimeType.String, FileSize: row.FileSize.Int64, ObjectKey: row.OriginalObjectKey.String, Checksum: row.Checksum.String, Status: row.Status, ErrorCode: row.ErrorCode.String, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
}
func knowledgeUUID(value string) (pgtype.UUID, error) {
	id, err := uuid.Parse(value)
	return pgtype.UUID{Bytes: id, Valid: err == nil}, err
}
func knowledgeID(value pgtype.UUID) string { return uuid.UUID(value.Bytes).String() }
func ktext(value string) pgtype.Text       { return pgtype.Text{String: value, Valid: value != ""} }
func queueKnowledgeEvent(ctx context.Context, queries *sqlc.Queries, workspaceID, sourceID pgtype.UUID, eventType string) error {
	eventID, _ := knowledgeUUID(uuid.NewString())
	return queries.QueueKnowledgeEvent(ctx, sqlc.QueueKnowledgeEventParams{ID: eventID, WorkspaceID: workspaceID, EventType: eventType, AggregateID: sourceID, Payload: []byte(`{}`)})
}
func knowledgeMutationIDs(userID, sourceID string) (pgtype.UUID, pgtype.UUID, error) {
	uid, err := knowledgeUUID(userID)
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, ErrInvalidInput
	}
	sid, err := knowledgeUUID(sourceID)
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, ErrInvalidInput
	}
	return uid, sid, nil
}
func mapKnowledgePersistenceError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return ErrDocumentDuplicate
	}
	return err
}
