package onboarding

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"authbackend/generated/sqlc"
)

var (
	ErrNotFound      = errors.New("onboarding profile not found")
	ErrForbidden     = errors.New("onboarding forbidden")
	ErrInvalidInput  = errors.New("invalid onboarding input")
)

type AdminInput struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type OnboardingData struct {
	BusinessName       string       `json:"business_name"`
	Industry           string       `json:"industry"`
	CustomIndustry     string       `json:"custom_industry,omitempty"`
	BusinessType       string       `json:"business_type"`
	BusinessDescription string      `json:"business_description"`
	TargetCustomer     string       `json:"target_customer,omitempty"`
	ProductsServices   string       `json:"products_services,omitempty"`
	PrimaryUseCases    []string     `json:"primary_use_cases"`
	CommunicationStyle string       `json:"communication_style"`
	CustomStyle        string       `json:"custom_style,omitempty"`
	HandoffRules       []string     `json:"handoff_rules"`
	Admins             []AdminInput `json:"admins,omitempty"`
	SelectedTemplateID string       `json:"selected_template_id,omitempty"`
}

type OnboardingProfile struct {
	ID          string          `json:"id"`
	WorkspaceID string          `json:"workspace_id"`
	UserID      string          `json:"user_id"`
	CurrentStep int             `json:"current_step"`
	IsCompleted bool            `json:"is_completed"`
	Data        OnboardingData  `json:"data"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type Repository interface {
	Authorize(ctx context.Context, userID, workspaceID string) (string, error)
	Get(ctx context.Context, workspaceID string) (OnboardingProfile, error)
	Save(ctx context.Context, profile OnboardingProfile) (OnboardingProfile, error)
	Complete(ctx context.Context, workspaceID string) (OnboardingProfile, error)
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Authorize(ctx context.Context, userID, workspaceID string) (string, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return "", err
	}
	wid, err := uuid.Parse(workspaceID)
	if err != nil {
		return "", err
	}
	return sqlc.New(r.pool).AuthorizeWorkspaceMember(ctx, sqlc.AuthorizeWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: wid, Valid: true},
		UserID:      pgtype.UUID{Bytes: uid, Valid: true},
	})
}

func (r *PostgresRepository) Get(ctx context.Context, workspaceID string) (OnboardingProfile, error) {
	wid, err := uuid.Parse(workspaceID)
	if err != nil {
		return OnboardingProfile{}, err
	}
	row, err := sqlc.New(r.pool).GetOnboardingProfile(ctx, pgtype.UUID{Bytes: wid, Valid: true})
	if err != nil {
		return OnboardingProfile{}, err
	}
	var data OnboardingData
	_ = json.Unmarshal(row.Data, &data)
	var completedAt *time.Time
	if row.CompletedAt.Valid {
		completedAt = &row.CompletedAt.Time
	}
	return OnboardingProfile{
		ID:          uuid.UUID(row.ID.Bytes).String(),
		WorkspaceID: uuid.UUID(row.WorkspaceID.Bytes).String(),
		UserID:      uuid.UUID(row.UserID.Bytes).String(),
		CurrentStep: int(row.CurrentStep),
		IsCompleted: row.IsCompleted,
		Data:        data,
		CompletedAt: completedAt,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}

func (r *PostgresRepository) Save(ctx context.Context, p OnboardingProfile) (OnboardingProfile, error) {
	wid, err := uuid.Parse(p.WorkspaceID)
	if err != nil {
		return OnboardingProfile{}, err
	}
	uid, err := uuid.Parse(p.UserID)
	if err != nil {
		return OnboardingProfile{}, err
	}
	id, err := uuid.Parse(p.ID)
	if err != nil {
		id = uuid.New()
	}
	dataBytes, _ := json.Marshal(p.Data)
	var completedAt pgtype.Timestamptz
	if p.CompletedAt != nil {
		completedAt = pgtype.Timestamptz{Time: *p.CompletedAt, Valid: true}
	}
	row, err := sqlc.New(r.pool).UpsertOnboardingProfile(ctx, sqlc.UpsertOnboardingProfileParams{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		WorkspaceID: pgtype.UUID{Bytes: wid, Valid: true},
		UserID:      pgtype.UUID{Bytes: uid, Valid: true},
		CurrentStep: int32(p.CurrentStep),
		IsCompleted: p.IsCompleted,
		Data:        dataBytes,
		CompletedAt: completedAt,
	})
	if err != nil {
		return OnboardingProfile{}, err
	}
	var data OnboardingData
	_ = json.Unmarshal(row.Data, &data)
	var compAt *time.Time
	if row.CompletedAt.Valid {
		compAt = &row.CompletedAt.Time
	}
	return OnboardingProfile{
		ID:          uuid.UUID(row.ID.Bytes).String(),
		WorkspaceID: uuid.UUID(row.WorkspaceID.Bytes).String(),
		UserID:      uuid.UUID(row.UserID.Bytes).String(),
		CurrentStep: int(row.CurrentStep),
		IsCompleted: row.IsCompleted,
		Data:        data,
		CompletedAt: compAt,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}

func (r *PostgresRepository) Complete(ctx context.Context, workspaceID string) (OnboardingProfile, error) {
	wid, err := uuid.Parse(workspaceID)
	if err != nil {
		return OnboardingProfile{}, err
	}
	row, err := sqlc.New(r.pool).CompleteOnboardingProfile(ctx, pgtype.UUID{Bytes: wid, Valid: true})
	if err != nil {
		return OnboardingProfile{}, err
	}
	var data OnboardingData
	_ = json.Unmarshal(row.Data, &data)
	var compAt *time.Time
	if row.CompletedAt.Valid {
		compAt = &row.CompletedAt.Time
	}
	return OnboardingProfile{
		ID:          uuid.UUID(row.ID.Bytes).String(),
		WorkspaceID: uuid.UUID(row.WorkspaceID.Bytes).String(),
		UserID:      uuid.UUID(row.UserID.Bytes).String(),
		CurrentStep: int(row.CurrentStep),
		IsCompleted: row.IsCompleted,
		Data:        data,
		CompletedAt: compAt,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}
