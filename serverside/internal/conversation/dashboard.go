package conversation

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidDashboardInput = errors.New("invalid conversation dashboard input")
	ErrDashboardForbidden    = errors.New("conversation dashboard forbidden")
)

type DashboardConversation struct {
	ID             string    `json:"id"`
	WorkspaceID    string    `json:"workspace_id"`
	AgentID        string    `json:"agent_id"`
	ChannelID      string    `json:"channel_id"`
	ExternalUserID string    `json:"external_user_id"`
	Status         string    `json:"status"`
	Mode           string    `json:"mode"`
	Environment    string    `json:"environment"`
	AssignedUserID *string   `json:"assigned_user_id,omitempty"`
	StartedAt      time.Time `json:"started_at"`
	LastMessageAt  time.Time `json:"last_message_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
type DashboardMessage struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	SenderType     string    `json:"sender_type"`
	ContentType    string    `json:"content_type"`
	Content        string    `json:"content"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}
type ListFilter struct {
	Status, Mode string
	Limit        int
}
type MessageFilter struct {
	Cursor *time.Time
	Limit  int
}
type DashboardRepository interface {
	List(context.Context, string, string, ListFilter) ([]DashboardConversation, error)
	Get(context.Context, string, string) (DashboardConversation, error)
	Messages(context.Context, string, string, MessageFilter) ([]DashboardMessage, error)
	Role(context.Context, string, string) (string, error)
	SetMode(context.Context, string, string, Mode) error
}
type DashboardService struct{ repo DashboardRepository }

func NewDashboardService(repo DashboardRepository) *DashboardService { return &DashboardService{repo} }
func (s *DashboardService) List(ctx context.Context, userID, workspaceID string, filter ListFilter) ([]DashboardConversation, error) {
	if strings.TrimSpace(workspaceID) == "" || !validListFilter(filter) {
		return nil, ErrInvalidDashboardInput
	}
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	return s.repo.List(ctx, workspaceID, userID, filter)
}
func (s *DashboardService) Get(ctx context.Context, userID, id string) (DashboardConversation, error) {
	if strings.TrimSpace(id) == "" {
		return DashboardConversation{}, ErrInvalidDashboardInput
	}
	return s.repo.Get(ctx, id, userID)
}
func (s *DashboardService) Messages(ctx context.Context, userID, id string, filter MessageFilter) ([]DashboardMessage, error) {
	if strings.TrimSpace(id) == "" || filter.Limit < 0 || filter.Limit > 100 {
		return nil, ErrInvalidDashboardInput
	}
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	return s.repo.Messages(ctx, id, userID, filter)
}
func (s *DashboardService) SetMode(ctx context.Context, userID, id string, mode Mode) error {
	if strings.TrimSpace(id) == "" || (mode != ModeAgent && mode != ModeHuman) {
		return ErrInvalidDashboardInput
	}
	role, err := s.repo.Role(ctx, id, userID)
	if err != nil {
		return err
	}
	if role != "owner" && role != "admin" {
		return ErrDashboardForbidden
	}
	return s.repo.SetMode(ctx, id, userID, mode)
}
func validListFilter(f ListFilter) bool {
	return (f.Status == "" || f.Status == "open" || f.Status == "closed") && (f.Mode == "" || f.Mode == "agent" || f.Mode == "human") && f.Limit >= 0 && f.Limit <= 100
}
