package jobs

import (
	"authbackend/generated/sqlc"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresQueue struct{ pool *pgxpool.Pool }

func NewPostgresQueue(pool *pgxpool.Pool) *PostgresQueue { return &PostgresQueue{pool} }
func (q *PostgresQueue) Claim(ctx context.Context) (Event, error) {
	row, err := sqlc.New(q.pool).ClaimOutboxEvent(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return Event{}, ErrNoJob
	}
	if err != nil {
		return Event{}, err
	}
	return Event{ID: idString(row.ID), WorkspaceID: idString(row.WorkspaceID), Type: row.EventType, AggregateID: idString(row.AggregateID), Attempts: int(row.Attempts)}, nil
}
func (q *PostgresQueue) Complete(ctx context.Context, eventID string) error {
	id, err := parseID(eventID)
	if err != nil {
		return err
	}
	return sqlc.New(q.pool).CompleteOutboxEvent(ctx, id)
}
func (q *PostgresQueue) Retry(ctx context.Context, eventID string, available time.Time, code string) error {
	id, err := parseID(eventID)
	if err != nil {
		return err
	}
	return sqlc.New(q.pool).RetryOutboxEvent(ctx, sqlc.RetryOutboxEventParams{ID: id, AvailableAt: pgtype.Timestamptz{Time: available, Valid: true}, LastErrorCode: pgtype.Text{String: code, Valid: true}})
}
func (q *PostgresQueue) Dead(ctx context.Context, event Event, code string) error {
	id, err := parseID(event.ID)
	if err != nil {
		return err
	}
	workspaceID, err := parseID(event.WorkspaceID)
	if err != nil {
		return err
	}
	tx, err := q.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := sqlc.New(tx)
	if err = queries.FailOutboxEvent(ctx, sqlc.FailOutboxEventParams{ID: id, LastErrorCode: pgtype.Text{String: code, Valid: true}}); err != nil {
		return err
	}
	failedID, _ := parseID(uuid.NewString())
	if err = queries.CreateFailedJob(ctx, sqlc.CreateFailedJobParams{ID: failedID, WorkspaceID: workspaceID, OutboxEventID: id, JobType: event.Type, ErrorCode: code, Attempts: int32(event.Attempts)}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func parseID(value string) (pgtype.UUID, error) {
	id, err := uuid.Parse(value)
	return pgtype.UUID{Bytes: id, Valid: err == nil}, err
}
func idString(value pgtype.UUID) string { return uuid.UUID(value.Bytes).String() }
