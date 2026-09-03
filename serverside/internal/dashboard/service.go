package dashboard

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidWorkspace = errors.New("workspace is required")
	ErrNotFound         = errors.New("dashboard resource not found")
)

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PlatformRole string    `json:"platform_role"`
	AccessMode   string    `json:"access_mode"`
	CreatedAt    time.Time `json:"created_at"`
}

type WorkspaceMembership struct {
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Status      string `json:"status"`
	Timezone    string `json:"timezone"`
	Role        string `json:"role"`
}

type Me struct {
	User       User                  `json:"user"`
	Workspaces []WorkspaceMembership `json:"workspaces"`
}

type ResourceStatus struct {
	Status string `json:"status"`
}

type SecondBrainStatus struct {
	Status           string `json:"status"`
	KnowledgeSources int64  `json:"knowledge_sources"`
}

type ConversationSummary struct {
	Today int64 `json:"today"`
	Open  int64 `json:"open"`
}

type Overview struct {
	WorkspaceID   string              `json:"workspace_id"`
	Agent         ResourceStatus      `json:"agent"`
	SecondBrain   SecondBrainStatus   `json:"second_brain"`
	Channel       ResourceStatus      `json:"channel"`
	Conversations ConversationSummary `json:"conversations"`
}

type Repository interface {
	GetMe(context.Context, string) (Me, error)
	GetOverview(context.Context, string, string) (Overview, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) Me(ctx context.Context, userID string) (Me, error) {
	return s.repository.GetMe(ctx, userID)
}

func (s *Service) Overview(ctx context.Context, userID, workspaceID string) (Overview, error) {
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return Overview{}, ErrInvalidWorkspace
	}
	return s.repository.GetOverview(ctx, userID, workspaceID)
}
