package workspace

import (
	"authbackend/generated/sqlc"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreateWithOwnerTrial(ctx context.Context, w Workspace, m Membership, s Subscription, e Entitlement, a AgentSeed, b BrainSeed, event OutboxSeed) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := sqlc.New(tx)
	if _, err = q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{ID: dbUUID(w.ID), Name: w.Name, Slug: w.Slug, OwnerUserID: dbUUID(w.OwnerUserID), Status: string(w.Status), Timezone: w.Timezone}); err != nil {
		return mapPostgresError(err)
	}
	if _, err = q.CreateWorkspaceMember(ctx, sqlc.CreateWorkspaceMemberParams{ID: dbUUID(m.ID), WorkspaceID: dbUUID(m.WorkspaceID), UserID: dbUUID(m.UserID), Role: string(m.Role)}); err != nil {
		return err
	}
	
	subParams := sqlc.CreateSubscriptionParams{
		ID:                 dbUUID(s.ID),
		WorkspaceID:        dbUUID(s.WorkspaceID),
		Status:             string(s.Status),
		PlanID:             s.PlanID,
		BillingCycle:       s.BillingCycle,
		Currency:           s.Currency,
		Amount:             s.Amount,
		CurrentPeriodStart: pgtype.Timestamptz{Time: s.CurrentPeriodStart, Valid: true},
		CurrentPeriodEnd:   pgtype.Timestamptz{Time: s.CurrentPeriodEnd, Valid: true},
		PaymentReference:   pgtype.Text{String: s.PaymentReference, Valid: s.PaymentReference != ""},
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
	}

	if _, err = q.CreateSubscription(ctx, subParams); err != nil {
		return err
	}
	if _, err = q.CreateSubscriptionEntitlement(ctx, sqlc.CreateSubscriptionEntitlementParams{ID: dbUUID(e.ID), SubscriptionID: dbUUID(e.SubscriptionID), WorkspaceID: dbUUID(e.WorkspaceID), MaxAgents: int32(e.MaxAgents), MaxChannels: int32(e.MaxChannels), MaxStorageMb: e.MaxStorageMB, MaxDocuments: int32(e.MaxDocuments), MonthlyMessages: e.MonthlyMessages, PublicApi: e.PublicAPI, Webhooks: e.Webhooks}); err != nil {
		return err
	}
	if _, err = q.CreateAgent(ctx, sqlc.CreateAgentParams{ID: dbUUID(a.ID), WorkspaceID: dbUUID(a.WorkspaceID), Name: a.Name, Status: a.Status}); err != nil {
		return err
	}
	if _, err = q.CreateSecondBrain(ctx, sqlc.CreateSecondBrainParams{ID: dbUUID(b.ID), WorkspaceID: dbUUID(b.WorkspaceID), AgentID: dbUUID(b.AgentID), Status: b.Status}); err != nil {
		return err
	}
	if _, err = q.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{ID: dbUUID(event.ID), WorkspaceID: dbUUID(event.WorkspaceID), EventType: event.Type, AggregateType: event.AggregateType, AggregateID: dbUUID(event.AggregateID), Payload: []byte(`{}`)}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PostgresRepository) GetForMember(ctx context.Context, workspaceID, userID string) (Workspace, Membership, error) {
	workspaceUUID, err := parseDBUUID(workspaceID)
	if err != nil {
		return Workspace{}, Membership{}, ErrNotFound
	}
	userUUID, err := parseDBUUID(userID)
	if err != nil {
		return Workspace{}, Membership{}, ErrNotFound
	}
	row, err := sqlc.New(r.pool).GetWorkspaceForMember(ctx, sqlc.GetWorkspaceForMemberParams{ID: workspaceUUID, UserID: userUUID})
	if errors.Is(err, pgx.ErrNoRows) {
		return Workspace{}, Membership{}, ErrNotFound
	}
	if err != nil {
		return Workspace{}, Membership{}, err
	}
	workspace := Workspace{ID: uuidString(row.ID), Name: row.Name, Slug: row.Slug, OwnerUserID: uuidString(row.OwnerUserID), Status: Status(row.Status), Timezone: row.Timezone, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
	membership := Membership{ID: uuidString(row.MembershipID), WorkspaceID: workspace.ID, UserID: uuidString(row.UserID), Role: Role(row.Role), CreatedAt: row.MembershipCreatedAt.Time, UpdatedAt: row.MembershipUpdatedAt.Time}
	return workspace, membership, nil
}

func (r *PostgresRepository) Update(ctx context.Context, w Workspace) error {
	id, err := parseDBUUID(w.ID)
	if err != nil {
		return ErrNotFound
	}
	_, err = sqlc.New(r.pool).UpdateWorkspace(ctx, sqlc.UpdateWorkspaceParams{ID: id, Name: w.Name, Timezone: w.Timezone})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (r *PostgresRepository) GetSubscriptionForMember(ctx context.Context, workspaceID, userID string) (SubscriptionDetail, error) {
	if _, _, err := r.GetForMember(ctx, workspaceID, userID); err != nil {
		return SubscriptionDetail{}, err
	}
	id, err := parseDBUUID(workspaceID)
	if err != nil {
		return SubscriptionDetail{}, ErrNotFound
	}
	row, err := sqlc.New(r.pool).GetWorkspaceSubscription(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return SubscriptionDetail{}, ErrNotFound
	}
	if err != nil {
		return SubscriptionDetail{}, err
	}
	subscriptionID := uuidString(row.ID)
	var trialStartedAt *time.Time
	if row.TrialStartedAt.Valid {
		t := row.TrialStartedAt.Time
		trialStartedAt = &t
	}
	var trialEndsAt *time.Time
	if row.TrialEndsAt.Valid {
		t := row.TrialEndsAt.Time
		trialEndsAt = &t
	}

	return SubscriptionDetail{
		Subscription: Subscription{
			ID:                 subscriptionID,
			WorkspaceID:        uuidString(row.WorkspaceID),
			Status:             SubscriptionStatus(row.Status),
			PlanID:             row.PlanID,
			BillingCycle:       row.BillingCycle,
			Currency:           row.Currency,
			Amount:             row.Amount,
			CurrentPeriodStart: row.CurrentPeriodStart.Time,
			CurrentPeriodEnd:   row.CurrentPeriodEnd.Time,
			PaymentReference:   row.PaymentReference.String,
			CancelAtPeriodEnd:  row.CancelAtPeriodEnd,
			TrialStartedAt:     trialStartedAt,
			TrialEndsAt:        trialEndsAt,
			CreatedAt:          row.CreatedAt.Time,
			UpdatedAt:          row.UpdatedAt.Time,
		},
		Entitlement: Entitlement{
			ID:              subscriptionID,
			SubscriptionID:  subscriptionID,
			WorkspaceID:     workspaceID,
			MaxAgents:       int(row.MaxAgents),
			MaxChannels:     int(row.MaxChannels),
			MaxStorageMB:    row.MaxStorageMb,
			MaxDocuments:    int(row.MaxDocuments),
			MonthlyMessages: row.MonthlyMessages,
			PublicAPI:       row.PublicApi,
			Webhooks:        row.Webhooks,
		},
	}, nil
}

func parseDBUUID(value string) (pgtype.UUID, error) {
	id, err := uuid.Parse(value)
	return pgtype.UUID{Bytes: id, Valid: err == nil}, err
}

func dbUUID(value string) pgtype.UUID {
	id, _ := parseDBUUID(value)
	return id
}

func uuidString(value pgtype.UUID) string {
	return uuid.UUID(value.Bytes).String()
}

func mapPostgresError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrSlugExists
	}
	return err
}
