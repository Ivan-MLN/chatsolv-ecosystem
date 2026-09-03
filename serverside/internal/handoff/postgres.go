package handoff

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"authbackend/generated/sqlc"
)

var (
	ErrHandoffNotFound     = errors.New("handoff request tidak ditemukan")
	ErrHandoffAlreadyClaimed = errors.New("percakapan sudah diambil oleh admin lain")
	ErrForbidden           = errors.New("akses handoff ditolak")
	ErrInvalidInput        = errors.New("input handoff tidak valid")
)

type HandoffRequest struct {
	ID              string     `json:"id"`
	ShortCode       string     `json:"short_code"`
	WorkspaceID     string     `json:"workspace_id"`
	ConversationID  string     `json:"conversation_id"`
	CustomerPhone   string     `json:"customer_phone"`
	Reason          string     `json:"reason"`
	Status          string     `json:"status"` // 'pending', 'assigned', 'accepted', 'resolved', 'expired', 'cancelled'
	AssignedAdminID *string    `json:"assigned_admin_id,omitempty"`
	AssignedAdminName *string  `json:"assigned_admin_name,omitempty"`
	RequestedAt     time.Time  `json:"requested_at"`
	AssignedAt      *time.Time `json:"assigned_at,omitempty"`
	AcceptedAt      *time.Time `json:"accepted_at,omitempty"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	TimeoutAt       time.Time  `json:"timeout_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ConversationEvent struct {
	ID             string          `json:"id"`
	WorkspaceID    string          `json:"workspace_id"`
	ConversationID string          `json:"conversation_id"`
	EventType      string          `json:"event_type"`
	ActorType      string          `json:"actor_type"`
	ActorID        *string         `json:"actor_id,omitempty"`
	Payload        map[string]any  `json:"payload"`
	CreatedAt      time.Time       `json:"created_at"`
}

type Repository interface {
	Authorize(ctx context.Context, userID, workspaceID string) (string, error)
	CreateHandoff(ctx context.Context, h HandoffRequest) (HandoffRequest, error)
	GetByShortCode(ctx context.Context, code string) (HandoffRequest, error)
	GetByID(ctx context.Context, workspaceID, id string) (HandoffRequest, error)
	List(ctx context.Context, workspaceID string, limit int) ([]HandoffRequest, error)
	AcceptAtomic(ctx context.Context, handoffID, adminID string) (HandoffRequest, error)
	Resolve(ctx context.Context, handoffID string) (HandoffRequest, error)
	Reassign(ctx context.Context, handoffID, adminID string, timeout time.Time) (HandoffRequest, error)
	SetConversationModeAndAdmin(ctx context.Context, conversationID, workspaceID, mode string, adminID *string, handoffID *string, reason string) error
	RecordEvent(ctx context.Context, event ConversationEvent) error
	ListEvents(ctx context.Context, workspaceID, conversationID string) ([]ConversationEvent, error)
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

func generateShortCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(9000))
	return fmt.Sprintf("CS-%d", 1000+n.Int64())
}

func (r *PostgresRepository) CreateHandoff(ctx context.Context, h HandoffRequest) (HandoffRequest, error) {
	wid, err := uuid.Parse(h.WorkspaceID)
	if err != nil {
		return HandoffRequest{}, err
	}
	cid, err := uuid.Parse(h.ConversationID)
	if err != nil {
		return HandoffRequest{}, err
	}
	id, err := uuid.Parse(h.ID)
	if err != nil {
		id = uuid.New()
	}

	code := h.ShortCode
	if code == "" {
		code = generateShortCode()
	}

	var aid pgtype.UUID
	if h.AssignedAdminID != nil {
		if a, err := uuid.Parse(*h.AssignedAdminID); err == nil {
			aid = pgtype.UUID{Bytes: a, Valid: true}
		}
	}

	var assignedAt pgtype.Timestamptz
	if h.AssignedAt != nil {
		assignedAt = pgtype.Timestamptz{Time: *h.AssignedAt, Valid: true}
	}

	timeout := h.TimeoutAt
	if timeout.IsZero() {
		timeout = time.Now().Add(2 * time.Minute)
	}

	row, err := sqlc.New(r.pool).CreateHandoffRequest(ctx, sqlc.CreateHandoffRequestParams{
		ID:              pgtype.UUID{Bytes: id, Valid: true},
		ShortCode:       code,
		WorkspaceID:     pgtype.UUID{Bytes: wid, Valid: true},
		ConversationID:  pgtype.UUID{Bytes: cid, Valid: true},
		CustomerPhone:   h.CustomerPhone,
		Reason:          h.Reason,
		Status:          h.Status,
		AssignedAdminID: aid,
		AssignedAt:      assignedAt,
		TimeoutAt:       pgtype.Timestamptz{Time: timeout, Valid: true},
		Metadata:        []byte(`{}`),
	})
	if err != nil {
		return HandoffRequest{}, err
	}

	return r.mapHandoffRow(row), nil
}

func (r *PostgresRepository) GetByShortCode(ctx context.Context, code string) (HandoffRequest, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	row, err := sqlc.New(r.pool).GetHandoffByShortCode(ctx, code)
	if err != nil {
		return HandoffRequest{}, err
	}
	return r.mapHandoffRow(sqlc.HandoffRequest(row)), nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, workspaceID, id string) (HandoffRequest, error) {
	wid, err := uuid.Parse(workspaceID)
	if err != nil {
		return HandoffRequest{}, err
	}
	hid, err := uuid.Parse(id)
	if err != nil {
		return HandoffRequest{}, err
	}
	row, err := sqlc.New(r.pool).GetHandoffByID(ctx, sqlc.GetHandoffByIDParams{
		WorkspaceID: pgtype.UUID{Bytes: wid, Valid: true},
		ID:          pgtype.UUID{Bytes: hid, Valid: true},
	})
	if err != nil {
		return HandoffRequest{}, err
	}
	return r.mapHandoffRow(sqlc.HandoffRequest(row)), nil
}

func (r *PostgresRepository) List(ctx context.Context, workspaceID string, limit int) ([]HandoffRequest, error) {
	wid, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := sqlc.New(r.pool).ListHandoffRequests(ctx, sqlc.ListHandoffRequestsParams{
		WorkspaceID: pgtype.UUID{Bytes: wid, Valid: true},
		Limit:       int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]HandoffRequest, 0, len(rows))
	for _, row := range rows {
		out = append(out, r.mapHandoffRow(sqlc.HandoffRequest(row)))
	}
	return out, nil
}

func (r *PostgresRepository) AcceptAtomic(ctx context.Context, handoffID, adminID string) (HandoffRequest, error) {
	hid, err := uuid.Parse(handoffID)
	if err != nil {
		return HandoffRequest{}, err
	}
	aid, err := uuid.Parse(adminID)
	if err != nil {
		return HandoffRequest{}, err
	}
	row, err := sqlc.New(r.pool).AcceptHandoffRequestAtomic(ctx, sqlc.AcceptHandoffRequestAtomicParams{
		ID:              pgtype.UUID{Bytes: hid, Valid: true},
		AssignedAdminID: pgtype.UUID{Bytes: aid, Valid: true},
	})
	if err != nil {
		return HandoffRequest{}, ErrHandoffAlreadyClaimed
	}
	return r.mapHandoffRow(sqlc.HandoffRequest(row)), nil
}

func (r *PostgresRepository) Resolve(ctx context.Context, handoffID string) (HandoffRequest, error) {
	hid, err := uuid.Parse(handoffID)
	if err != nil {
		return HandoffRequest{}, err
	}
	row, err := sqlc.New(r.pool).ResolveHandoffRequest(ctx, pgtype.UUID{Bytes: hid, Valid: true})
	if err != nil {
		return HandoffRequest{}, err
	}
	return r.mapHandoffRow(sqlc.HandoffRequest(row)), nil
}

func (r *PostgresRepository) Reassign(ctx context.Context, handoffID, adminID string, timeout time.Time) (HandoffRequest, error) {
	hid, err := uuid.Parse(handoffID)
	if err != nil {
		return HandoffRequest{}, err
	}
	aid, err := uuid.Parse(adminID)
	if err != nil {
		return HandoffRequest{}, err
	}
	row, err := sqlc.New(r.pool).ReassignHandoffRequest(ctx, sqlc.ReassignHandoffRequestParams{
		ID:              pgtype.UUID{Bytes: hid, Valid: true},
		AssignedAdminID: pgtype.UUID{Bytes: aid, Valid: true},
		TimeoutAt:       pgtype.Timestamptz{Time: timeout, Valid: true},
	})
	if err != nil {
		return HandoffRequest{}, err
	}
	return r.mapHandoffRow(sqlc.HandoffRequest(row)), nil
}

func (r *PostgresRepository) SetConversationModeAndAdmin(ctx context.Context, conversationID, workspaceID, mode string, adminID *string, handoffID *string, reason string) error {
	cid, err := uuid.Parse(conversationID)
	if err != nil {
		return err
	}
	wid, err := uuid.Parse(workspaceID)
	if err != nil {
		return err
	}

	var aid *uuid.UUID
	if adminID != nil {
		if a, err := uuid.Parse(*adminID); err == nil {
			aid = &a
		}
	}

	var hid *uuid.UUID
	if handoffID != nil {
		if h, err := uuid.Parse(*handoffID); err == nil {
			hid = &h
		}
	}

	var humanStart *time.Time
	var humanEnd *time.Time
	now := time.Now()
	if mode == "human" {
		humanStart = &now
	} else if mode == "agent" {
		humanEnd = &now
	}

	query := `
		UPDATE conversations SET
			mode = $3,
			assigned_admin_id = $4,
			current_handoff_id = $5,
			handoff_reason = COALESCE(NULLIF($6, ''), handoff_reason),
			human_started_at = CASE WHEN $3 = 'human' THEN COALESCE(human_started_at, $7) ELSE human_started_at END,
			human_ended_at = CASE WHEN $3 = 'agent' THEN $8 ELSE human_ended_at END,
			updated_at = now()
		WHERE id = $1 AND workspace_id = $2
	`
	_, err = r.pool.Exec(ctx, query, cid, wid, mode, aid, hid, reason, humanStart, humanEnd)
	return err
}

func (r *PostgresRepository) RecordEvent(ctx context.Context, ev ConversationEvent) error {
	wid, err := uuid.Parse(ev.WorkspaceID)
	if err != nil {
		return err
	}
	cid, err := uuid.Parse(ev.ConversationID)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(ev.ID)
	if err != nil {
		id = uuid.New()
	}
	payloadBytes, _ := json.Marshal(ev.Payload)
	var actorID pgtype.Text
	if ev.ActorID != nil {
		actorID = pgtype.Text{String: *ev.ActorID, Valid: true}
	}
	return sqlc.New(r.pool).RecordConversationEvent(ctx, sqlc.RecordConversationEventParams{
		ID:             pgtype.UUID{Bytes: id, Valid: true},
		WorkspaceID:    pgtype.UUID{Bytes: wid, Valid: true},
		ConversationID: pgtype.UUID{Bytes: cid, Valid: true},
		EventType:      ev.EventType,
		ActorType:      ev.ActorType,
		ActorID:        actorID,
		Payload:        payloadBytes,
	})
}

func (r *PostgresRepository) ListEvents(ctx context.Context, workspaceID, conversationID string) ([]ConversationEvent, error) {
	wid, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, err
	}
	cid, err := uuid.Parse(conversationID)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(r.pool).ListConversationEvents(ctx, sqlc.ListConversationEventsParams{
		WorkspaceID:    pgtype.UUID{Bytes: wid, Valid: true},
		ConversationID: pgtype.UUID{Bytes: cid, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	out := make([]ConversationEvent, 0, len(rows))
	for _, row := range rows {
		var payload map[string]any
		_ = json.Unmarshal(row.Payload, &payload)
		var actID *string
		if row.ActorID.Valid {
			actID = &row.ActorID.String
		}
		out = append(out, ConversationEvent{
			ID:             uuid.UUID(row.ID.Bytes).String(),
			WorkspaceID:    uuid.UUID(row.WorkspaceID.Bytes).String(),
			ConversationID: uuid.UUID(row.ConversationID.Bytes).String(),
			EventType:      row.EventType,
			ActorType:      row.ActorType,
			ActorID:        actID,
			Payload:        payload,
			CreatedAt:      row.CreatedAt.Time,
		})
	}
	return out, nil
}

func (r *PostgresRepository) mapHandoffRow(row sqlc.HandoffRequest) HandoffRequest {
	var assignedAdmin *string
	if row.AssignedAdminID.Valid {
		a := uuid.UUID(row.AssignedAdminID.Bytes).String()
		assignedAdmin = &a
	}
	var assignedAt *time.Time
	if row.AssignedAt.Valid {
		assignedAt = &row.AssignedAt.Time
	}
	var acceptedAt *time.Time
	if row.AcceptedAt.Valid {
		acceptedAt = &row.AcceptedAt.Time
	}
	var resolvedAt *time.Time
	if row.ResolvedAt.Valid {
		resolvedAt = &row.ResolvedAt.Time
	}

	return HandoffRequest{
		ID:              uuid.UUID(row.ID.Bytes).String(),
		ShortCode:       row.ShortCode,
		WorkspaceID:     uuid.UUID(row.WorkspaceID.Bytes).String(),
		ConversationID:  uuid.UUID(row.ConversationID.Bytes).String(),
		CustomerPhone:   row.CustomerPhone,
		Reason:          row.Reason,
		Status:          row.Status,
		AssignedAdminID: assignedAdmin,
		RequestedAt:     row.RequestedAt.Time,
		AssignedAt:      assignedAt,
		AcceptedAt:      acceptedAt,
		ResolvedAt:      resolvedAt,
		TimeoutAt:       row.TimeoutAt.Time,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}
}
