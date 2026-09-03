package template

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"authbackend/generated/sqlc"
	"authbackend/internal/agentconfig"
)

var (
	ErrTemplateNotFound = errors.New("template tidak ditemukan")
	ErrForbidden        = errors.New("akses template ditolak")
	ErrInvalidInput     = errors.New("input template tidak valid")
)

type AgentTemplate struct {
	ID                  string                      `json:"id"`
	Industry            string                      `json:"industry"`
	Title               string                      `json:"title"`
	Description         string                      `json:"description"`
	Icon                string                      `json:"icon"`
	Category            string                      `json:"category"`
	DefaultProfile      agentconfig.AgentProfile    `json:"default_profile"`
	DefaultPersonality  agentconfig.Personality     `json:"default_personality"`
	DefaultUseCases     []string                    `json:"default_use_cases"`
	DefaultHandoffRules *agentconfig.HandoffRulesConfig `json:"default_handoff_rules,omitempty"`
	IsFeatured          bool                        `json:"is_featured"`
	CreatedAt           time.Time                   `json:"created_at"`
}

type Repository interface {
	Authorize(ctx context.Context, userID, workspaceID string) (string, error)
	List(ctx context.Context) ([]AgentTemplate, error)
	GetByID(ctx context.Context, id string) (AgentTemplate, error)
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

func (r *PostgresRepository) List(ctx context.Context) ([]AgentTemplate, error) {
	rows, err := sqlc.New(r.pool).ListAgentTemplates(ctx)
	if err != nil {
		return nil, err
	}
	templates := make([]AgentTemplate, 0, len(rows))
	for _, row := range rows {
		var prof agentconfig.AgentProfile
		_ = json.Unmarshal(row.DefaultProfile, &prof)
		var pers agentconfig.Personality
		_ = json.Unmarshal(row.DefaultPersonality, &pers)
		var useCases []string
		_ = json.Unmarshal(row.DefaultUseCases, &useCases)
		var handoffRules *agentconfig.HandoffRulesConfig
		if len(row.DefaultHandoffRules) > 0 {
			var hr agentconfig.HandoffRulesConfig
			if json.Unmarshal(row.DefaultHandoffRules, &hr) == nil {
				handoffRules = &hr
			}
		}

		templates = append(templates, AgentTemplate{
			ID:                  row.ID,
			Industry:            row.Industry,
			Title:               row.Title,
			Description:         row.Description,
			Icon:                row.Icon,
			Category:            row.Category,
			DefaultProfile:      prof,
			DefaultPersonality:  pers,
			DefaultUseCases:     useCases,
			DefaultHandoffRules: handoffRules,
			IsFeatured:          row.IsFeatured,
			CreatedAt:           row.CreatedAt.Time,
		})
	}
	return templates, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (AgentTemplate, error) {
	row, err := sqlc.New(r.pool).GetAgentTemplateByID(ctx, id)
	if err != nil {
		return AgentTemplate{}, err
	}
	var prof agentconfig.AgentProfile
	_ = json.Unmarshal(row.DefaultProfile, &prof)
	var pers agentconfig.Personality
	_ = json.Unmarshal(row.DefaultPersonality, &pers)
	var useCases []string
	_ = json.Unmarshal(row.DefaultUseCases, &useCases)
	var handoffRules *agentconfig.HandoffRulesConfig
	if len(row.DefaultHandoffRules) > 0 {
		var hr agentconfig.HandoffRulesConfig
		if json.Unmarshal(row.DefaultHandoffRules, &hr) == nil {
			handoffRules = &hr
		}
	}
	return AgentTemplate{
		ID:                  row.ID,
		Industry:            row.Industry,
		Title:               row.Title,
		Description:         row.Description,
		Icon:                row.Icon,
		Category:            row.Category,
		DefaultProfile:      prof,
		DefaultPersonality:  pers,
		DefaultUseCases:     useCases,
		DefaultHandoffRules: handoffRules,
		IsFeatured:          row.IsFeatured,
		CreatedAt:           row.CreatedAt.Time,
	}, nil
}

type AgentConfigUpdater interface {
	UpdateProfile(ctx context.Context, userID, workspaceID string, profile agentconfig.AgentProfile) (int64, error)
	UpdatePersonality(ctx context.Context, userID, workspaceID string, personality agentconfig.Personality) (int64, error)
	GetProfile(ctx context.Context, userID, workspaceID string) (agentconfig.AgentProfile, error)
	GetPersonality(ctx context.Context, userID, workspaceID string) (agentconfig.Personality, error)
}

type BusinessProfileUpdater interface {
	GetBusinessProfile(ctx context.Context, workspaceID string) (agentconfig.BusinessProfile, error)
	UpdateBusinessProfile(ctx context.Context, userID, workspaceID string, profile agentconfig.BusinessProfile) (int64, error)
}

type Service struct {
	repo         Repository
	agentUpdater AgentConfigUpdater
	bizUpdater   BusinessProfileUpdater
}

func NewService(repo Repository, agentUpdater AgentConfigUpdater, bizUpdater BusinessProfileUpdater) *Service {
	return &Service{
		repo:         repo,
		agentUpdater: agentUpdater,
		bizUpdater:   bizUpdater,
	}
}

func (s *Service) List(ctx context.Context, userID, workspaceID string) ([]AgentTemplate, error) {
	if strings.TrimSpace(workspaceID) != "" && strings.TrimSpace(userID) != "" {
		role, err := s.repo.Authorize(ctx, userID, workspaceID)
		if err != nil || role == "" {
			return nil, ErrForbidden
		}
	}
	return s.repo.List(ctx)
}

func (s *Service) ApplyTemplate(ctx context.Context, userID, workspaceID, templateID string) (AgentTemplate, error) {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(templateID) == "" {
		return AgentTemplate{}, ErrInvalidInput
	}
	role, err := s.repo.Authorize(ctx, userID, workspaceID)
	if err != nil || (role != "owner" && role != "admin") {
		return AgentTemplate{}, ErrForbidden
	}

	tmpl, err := s.repo.GetByID(ctx, templateID)
	if err != nil {
		return AgentTemplate{}, ErrTemplateNotFound
	}

	// Read business profile to customize template with business context
	bizName := "Toko Anda"
	if s.bizUpdater != nil {
		if bp, err := s.bizUpdater.GetBusinessProfile(ctx, workspaceID); err == nil && bp.BusinessName != "" {
			bizName = bp.BusinessName
			// Update business profile primary use cases & handoff rules
			if len(tmpl.DefaultUseCases) > 0 {
				bp.PrimaryUseCases = tmpl.DefaultUseCases
			}
			if tmpl.DefaultHandoffRules != nil {
				bp.HandoffRules = tmpl.DefaultHandoffRules
			}
			_, _ = s.bizUpdater.UpdateBusinessProfile(ctx, userID, workspaceID, bp)
		}
	}

	// Apply customized profile & personality
	prof := tmpl.DefaultProfile
	prof.DisplayName = strings.ReplaceAll(prof.DisplayName, "Toko Online", bizName)
	prof.DisplayName = strings.ReplaceAll(prof.DisplayName, "Restoran", bizName)
	prof.DisplayName = strings.ReplaceAll(prof.DisplayName, "Beauty", bizName)

	pers := tmpl.DefaultPersonality
	pers.BotName = strings.ReplaceAll(pers.BotName, "CS Online", "CS "+bizName)
	pers.CustomInstructions = strings.ReplaceAll(pers.CustomInstructions, "toko kami", bizName)

	if s.agentUpdater != nil {
		_, _ = s.agentUpdater.UpdateProfile(ctx, userID, workspaceID, prof)
		_, _ = s.agentUpdater.UpdatePersonality(ctx, userID, workspaceID, pers)
	}

	return tmpl, nil
}
