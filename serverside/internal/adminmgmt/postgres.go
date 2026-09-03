package adminmgmt

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"authbackend/generated/sqlc"
)

var (
	ErrAdminNotFound    = errors.New("admin tidak ditemukan")
	ErrForbidden        = errors.New("akses admin ditolak")
	ErrInvalidInput     = errors.New("data admin tidak valid")
	ErrDuplicatePhone   = errors.New("nomor WhatsApp admin sudah terdaftar")
	phoneRegex          = regexp.MustCompile(`^\+?[0-9]{8,18}$`)
)

func NormalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	if strings.HasPrefix(phone, "08") {
		phone = "628" + phone[2:]
	} else if strings.HasPrefix(phone, "+") {
		phone = phone[1:]
	}
	return phone
}

type WorkspaceAdmin struct {
	ID                 string     `json:"id"`
	WorkspaceID        string     `json:"workspace_id"`
	UserID             *string    `json:"user_id,omitempty"`
	Name               string     `json:"name"`
	Phone              string     `json:"phone"`
	Role               string     `json:"role"`
	Status             string     `json:"status"` // 'online', 'busy', 'offline'
	IsActive           bool       `json:"is_active"`
	RotationPriority   int        `json:"rotation_priority"`
	LastAssignedAt     *time.Time `json:"last_assigned_at,omitempty"`
	TotalHandledToday  int        `json:"total_handled_today"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type Repository interface {
	Authorize(ctx context.Context, userID, workspaceID string) (string, error)
	List(ctx context.Context, workspaceID string) ([]WorkspaceAdmin, error)
	GetByID(ctx context.Context, workspaceID, adminID string) (WorkspaceAdmin, error)
	FindByPhone(ctx context.Context, phone string) (WorkspaceAdmin, error)
	Create(ctx context.Context, admin WorkspaceAdmin) (WorkspaceAdmin, error)
	Update(ctx context.Context, admin WorkspaceAdmin) (WorkspaceAdmin, error)
	Delete(ctx context.Context, workspaceID, adminID string) error
	GetNextForRotation(ctx context.Context, workspaceID string) (WorkspaceAdmin, error)
	RecordAssignment(ctx context.Context, adminID string) error
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

func (r *PostgresRepository) List(ctx context.Context, workspaceID string) ([]WorkspaceAdmin, error) {
	wid, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(r.pool).ListWorkspaceAdmins(ctx, pgtype.UUID{Bytes: wid, Valid: true})
	if err != nil {
		return nil, err
	}
	admins := make([]WorkspaceAdmin, 0, len(rows))
	for _, row := range rows {
		var uid *string
		if row.UserID.Valid {
			u := uuid.UUID(row.UserID.Bytes).String()
			uid = &u
		}
		var assignedAt *time.Time
		if row.LastAssignedAt.Valid {
			assignedAt = &row.LastAssignedAt.Time
		}
		admins = append(admins, WorkspaceAdmin{
			ID:                uuid.UUID(row.ID.Bytes).String(),
			WorkspaceID:       uuid.UUID(row.WorkspaceID.Bytes).String(),
			UserID:            uid,
			Name:              row.Name,
			Phone:             row.Phone,
			Role:              row.Role,
			Status:            row.Status,
			IsActive:          row.IsActive,
			RotationPriority:  int(row.RotationPriority),
			LastAssignedAt:    assignedAt,
			TotalHandledToday: int(row.TotalHandledToday),
			CreatedAt:         row.CreatedAt.Time,
			UpdatedAt:         row.UpdatedAt.Time,
		})
	}
	return admins, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, workspaceID, adminID string) (WorkspaceAdmin, error) {
	wid, err := uuid.Parse(workspaceID)
	if err != nil {
		return WorkspaceAdmin{}, err
	}
	aid, err := uuid.Parse(adminID)
	if err != nil {
		return WorkspaceAdmin{}, err
	}
	row, err := sqlc.New(r.pool).GetWorkspaceAdminByID(ctx, sqlc.GetWorkspaceAdminByIDParams{
		WorkspaceID: pgtype.UUID{Bytes: wid, Valid: true},
		ID:          pgtype.UUID{Bytes: aid, Valid: true},
	})
	if err != nil {
		return WorkspaceAdmin{}, err
	}
	var uid *string
	if row.UserID.Valid {
		u := uuid.UUID(row.UserID.Bytes).String()
		uid = &u
	}
	var assignedAt *time.Time
	if row.LastAssignedAt.Valid {
		assignedAt = &row.LastAssignedAt.Time
	}
	return WorkspaceAdmin{
		ID:                uuid.UUID(row.ID.Bytes).String(),
		WorkspaceID:       uuid.UUID(row.WorkspaceID.Bytes).String(),
		UserID:            uid,
		Name:              row.Name,
		Phone:             row.Phone,
		Role:              row.Role,
		Status:            row.Status,
		IsActive:          row.IsActive,
		RotationPriority:  int(row.RotationPriority),
		LastAssignedAt:    assignedAt,
		TotalHandledToday: int(row.TotalHandledToday),
		CreatedAt:         row.CreatedAt.Time,
		UpdatedAt:         row.UpdatedAt.Time,
	}, nil
}

func (r *PostgresRepository) FindByPhone(ctx context.Context, phone string) (WorkspaceAdmin, error) {
	norm := NormalizePhone(phone)
	row, err := sqlc.New(r.pool).FindWorkspaceAdminByPhone(ctx, norm)
	if err != nil {
		return WorkspaceAdmin{}, err
	}
	var uid *string
	if row.UserID.Valid {
		u := uuid.UUID(row.UserID.Bytes).String()
		uid = &u
	}
	return WorkspaceAdmin{
		ID:          uuid.UUID(row.ID.Bytes).String(),
		WorkspaceID: uuid.UUID(row.WorkspaceID.Bytes).String(),
		UserID:      uid,
		Name:        row.Name,
		Phone:       row.Phone,
		Role:        row.Role,
		Status:      row.Status,
		IsActive:    row.IsActive,
	}, nil
}

func (r *PostgresRepository) Create(ctx context.Context, a WorkspaceAdmin) (WorkspaceAdmin, error) {
	wid, err := uuid.Parse(a.WorkspaceID)
	if err != nil {
		return WorkspaceAdmin{}, err
	}
	id, err := uuid.Parse(a.ID)
	if err != nil {
		id = uuid.New()
	}
	var uid pgtype.UUID
	if a.UserID != nil {
		if u, err := uuid.Parse(*a.UserID); err == nil {
			uid = pgtype.UUID{Bytes: u, Valid: true}
		}
	}
	status := a.Status
	if status == "" {
		status = "online"
	}
	role := a.Role
	if role == "" {
		role = "customer_service"
	}
	row, err := sqlc.New(r.pool).CreateWorkspaceAdmin(ctx, sqlc.CreateWorkspaceAdminParams{
		ID:               pgtype.UUID{Bytes: id, Valid: true},
		WorkspaceID:      pgtype.UUID{Bytes: wid, Valid: true},
		UserID:           uid,
		Name:             a.Name,
		Phone:            NormalizePhone(a.Phone),
		Role:             role,
		Status:           status,
		IsActive:         a.IsActive,
		RotationPriority: int32(a.RotationPriority),
	})
	if err != nil {
		return WorkspaceAdmin{}, err
	}
	return WorkspaceAdmin{
		ID:                uuid.UUID(row.ID.Bytes).String(),
		WorkspaceID:       uuid.UUID(row.WorkspaceID.Bytes).String(),
		Name:              row.Name,
		Phone:             row.Phone,
		Role:              row.Role,
		Status:            row.Status,
		IsActive:          row.IsActive,
		RotationPriority:  int(row.RotationPriority),
		TotalHandledToday: int(row.TotalHandledToday),
		CreatedAt:         row.CreatedAt.Time,
		UpdatedAt:         row.UpdatedAt.Time,
	}, nil
}

func (r *PostgresRepository) Update(ctx context.Context, a WorkspaceAdmin) (WorkspaceAdmin, error) {
	wid, err := uuid.Parse(a.WorkspaceID)
	if err != nil {
		return WorkspaceAdmin{}, err
	}
	aid, err := uuid.Parse(a.ID)
	if err != nil {
		return WorkspaceAdmin{}, err
	}
	row, err := sqlc.New(r.pool).UpdateWorkspaceAdmin(ctx, sqlc.UpdateWorkspaceAdminParams{
		WorkspaceID:      pgtype.UUID{Bytes: wid, Valid: true},
		ID:               pgtype.UUID{Bytes: aid, Valid: true},
		Name:             a.Name,
		Phone:            NormalizePhone(a.Phone),
		Role:             a.Role,
		Status:           a.Status,
		IsActive:         a.IsActive,
		RotationPriority: int32(a.RotationPriority),
	})
	if err != nil {
		return WorkspaceAdmin{}, err
	}
	return WorkspaceAdmin{
		ID:                uuid.UUID(row.ID.Bytes).String(),
		WorkspaceID:       uuid.UUID(row.WorkspaceID.Bytes).String(),
		Name:              row.Name,
		Phone:             row.Phone,
		Role:              row.Role,
		Status:            row.Status,
		IsActive:          row.IsActive,
		RotationPriority:  int(row.RotationPriority),
		TotalHandledToday: int(row.TotalHandledToday),
		CreatedAt:         row.CreatedAt.Time,
		UpdatedAt:         row.UpdatedAt.Time,
	}, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, workspaceID, adminID string) error {
	wid, err := uuid.Parse(workspaceID)
	if err != nil {
		return err
	}
	aid, err := uuid.Parse(adminID)
	if err != nil {
		return err
	}
	return sqlc.New(r.pool).DeleteWorkspaceAdmin(ctx, sqlc.DeleteWorkspaceAdminParams{
		WorkspaceID: pgtype.UUID{Bytes: wid, Valid: true},
		ID:          pgtype.UUID{Bytes: aid, Valid: true},
	})
}

func (r *PostgresRepository) GetNextForRotation(ctx context.Context, workspaceID string) (WorkspaceAdmin, error) {
	wid, err := uuid.Parse(workspaceID)
	if err != nil {
		return WorkspaceAdmin{}, err
	}
	row, err := sqlc.New(r.pool).GetNextEligibleAdminForRotation(ctx, pgtype.UUID{Bytes: wid, Valid: true})
	if err != nil {
		return WorkspaceAdmin{}, err
	}
	return WorkspaceAdmin{
		ID:          uuid.UUID(row.ID.Bytes).String(),
		WorkspaceID: uuid.UUID(row.WorkspaceID.Bytes).String(),
		Name:        row.Name,
		Phone:       row.Phone,
		Role:        row.Role,
		Status:      row.Status,
		IsActive:    row.IsActive,
	}, nil
}

func (r *PostgresRepository) RecordAssignment(ctx context.Context, adminID string) error {
	aid, err := uuid.Parse(adminID)
	if err != nil {
		return err
	}
	return sqlc.New(r.pool).RecordAdminAssignment(ctx, pgtype.UUID{Bytes: aid, Valid: true})
}
