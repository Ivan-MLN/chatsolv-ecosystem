package agentconfig

import (
	"context"
	"strings"
)

type Agent struct {
	ID                  string `json:"id"`
	WorkspaceID         string `json:"workspace_id"`
	Name                string `json:"name"`
	Status              string `json:"status"`
	Provider            string `json:"provider"`
	ConfigVersion       int64  `json:"config_version"`
	SyncedConfigVersion int64  `json:"synced_config_version"`
}

type AgentTestInput struct {
	WorkspaceID, AgentID, UserID, Message, Environment, ConversationID string
	Reset                                                              bool
}

type AgentTestResult struct {
	ConversationID string `json:"conversation_id"`
	Content        string `json:"content"`
}

type CanonicalRepository interface {
	ResolveDefaultAgent(context.Context, string, string) (Agent, string, error)
	UpdateAgent(context.Context, Agent) (Agent, error)
}

type AgentTester interface {
	Test(context.Context, AgentTestInput) (AgentTestResult, error)
}

type CanonicalService struct {
	repository  CanonicalRepository
	personality *Service
	settings    *SettingsService
	tester      AgentTester
	generator   *AIGenerator
}

func NewCanonicalService(repository CanonicalRepository, personality *Service, settings *SettingsService, tester AgentTester) *CanonicalService {
	return &CanonicalService{
		repository:  repository,
		personality: personality,
		settings:    settings,
		tester:      tester,
		generator:   NewAIGenerator(),
	}
}

func (s *CanonicalService) GenerateSetup(ctx context.Context, userID, workspaceID, description string) (GeneratedSetup, error) {
	_, err := s.Resolve(ctx, userID, workspaceID)
	if err != nil {
		return GeneratedSetup{}, err
	}
	return s.generator.GenerateSetup(ctx, description)
}

func (s *CanonicalService) Get(ctx context.Context, userID, workspaceID string) (Agent, error) {
	if strings.TrimSpace(workspaceID) == "" {
		return Agent{}, ErrInvalidInput
	}
	agent, _, err := s.repository.ResolveDefaultAgent(ctx, userID, workspaceID)
	return agent, err
}

func (s *CanonicalService) Update(ctx context.Context, userID, workspaceID, name string) (Agent, error) {
	agent, role, err := s.repository.ResolveDefaultAgent(ctx, userID, workspaceID)
	if err != nil {
		return Agent{}, err
	}
	if role != "owner" && role != "admin" {
		return Agent{}, ErrForbidden
	}
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 120 {
		return Agent{}, ErrInvalidInput
	}
	agent.Name = name
	return s.repository.UpdateAgent(ctx, agent)
}

func (s *CanonicalService) Resolve(ctx context.Context, userID, workspaceID string) (Agent, error) {
	return s.Get(ctx, userID, workspaceID)
}

func (s *CanonicalService) GetProfile(ctx context.Context, userID, workspaceID string) (AgentProfile, error) {
	agent, err := s.Resolve(ctx, userID, workspaceID)
	if err != nil {
		return AgentProfile{}, err
	}
	return s.settings.repository.GetAgentProfile(ctx, agent.ID)
}

func (s *CanonicalService) UpdateProfile(ctx context.Context, userID, workspaceID string, profile AgentProfile) (int64, error) {
	agent, err := s.Resolve(ctx, userID, workspaceID)
	if err != nil {
		return 0, err
	}
	return s.settings.UpdateAgentProfile(ctx, userID, agent.ID, profile)
}

func (s *CanonicalService) GetPersonality(ctx context.Context, userID, workspaceID string) (Personality, error) {
	agent, err := s.Resolve(ctx, userID, workspaceID)
	if err != nil {
		return Personality{}, err
	}
	return s.personality.repository.GetPersonality(ctx, agent.ID)
}

func (s *CanonicalService) UpdatePersonality(ctx context.Context, userID, workspaceID string, value Personality) (int64, error) {
	agent, err := s.Resolve(ctx, userID, workspaceID)
	if err != nil {
		return 0, err
	}
	return s.personality.UpdatePersonality(ctx, userID, agent.ID, value)
}

func (s *CanonicalService) Test(ctx context.Context, userID, workspaceID, message, conversationID string, reset bool) (AgentTestResult, error) {
	agent, err := s.Resolve(ctx, userID, workspaceID)
	if err != nil {
		return AgentTestResult{}, err
	}
	message = strings.TrimSpace(message)
	if !reset && (message == "" || len(message) > 4000) || s.tester == nil {
		return AgentTestResult{}, ErrInvalidInput
	}
	return s.tester.Test(ctx, AgentTestInput{
		WorkspaceID:    workspaceID,
		AgentID:        agent.ID,
		UserID:         userID,
		Message:        message,
		Environment:    "test",
		ConversationID: conversationID,
		Reset:          reset,
	})
}
