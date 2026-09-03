package dashboard

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

func (r *PostgresRepository) GetMe(ctx context.Context, userID string) (Me, error) {
	id, err := dashboardUUID(userID)
	if err != nil {
		return Me{}, ErrNotFound
	}
	rows, err := sqlc.New(r.pool).GetCurrentUserMemberships(ctx, id)
	if err != nil {
		return Me{}, err
	}
	if len(rows) == 0 {
		return Me{}, ErrNotFound
	}

	role := rows[0].PlatformRole
	if role == "" {
		role = "user"
	}
	accessMode := "subscription"
	if role == "developer" {
		accessMode = "developer"
	}

	result := Me{
		User: User{
			ID:           dashboardID(rows[0].UserID),
			Name:         rows[0].UserName,
			Email:        rows[0].Email,
			PlatformRole: role,
			AccessMode:   accessMode,
			CreatedAt:    rows[0].UserCreatedAt.Time,
		},
		Workspaces: make([]WorkspaceMembership, 0, len(rows)),
	}
	for _, row := range rows {
		if !row.WorkspaceID.Valid {
			continue
		}
		result.Workspaces = append(result.Workspaces, WorkspaceMembership{
			WorkspaceID: dashboardID(row.WorkspaceID),
			Name:        row.WorkspaceName.String,
			Slug:        row.Slug.String,
			Status:      row.Status.String,
			Timezone:    row.Timezone.String,
			Role:        row.Role.String,
		})
	}
	return result, nil
}

func (r *PostgresRepository) GetOverview(ctx context.Context, userID, workspaceID string) (Overview, error) {
	uid, err := dashboardUUID(userID)
	if err != nil {
		return Overview{}, ErrNotFound
	}
	wid, err := dashboardUUID(workspaceID)
	if err != nil {
		return Overview{}, ErrNotFound
	}
	row, err := sqlc.New(r.pool).GetDashboardOverview(ctx, sqlc.GetDashboardOverviewParams{ID: wid, UserID: uid})
	if errors.Is(err, pgx.ErrNoRows) {
		return Overview{}, ErrNotFound
	}
	if err != nil {
		return Overview{}, err
	}
	return Overview{
		WorkspaceID:   dashboardID(row.WorkspaceID),
		Agent:         ResourceStatus{Status: row.AgentStatus},
		SecondBrain:   SecondBrainStatus{Status: row.SecondBrainStatus, KnowledgeSources: row.KnowledgeSources},
		Channel:       ResourceStatus{Status: row.ChannelStatus},
		Conversations: ConversationSummary{Today: row.Today, Open: row.Open},
	}, nil
}

func dashboardUUID(value string) (pgtype.UUID, error) {
	id, err := uuid.Parse(value)
	return pgtype.UUID{Bytes: id, Valid: err == nil}, err
}

func dashboardID(value pgtype.UUID) string { return uuid.UUID(value.Bytes).String() }
