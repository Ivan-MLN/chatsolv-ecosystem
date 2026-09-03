package hermes

import "context"

type CreateAgentInput struct{ WorkspaceID, AgentID, VaultPath string }
type AgentResource struct{ ID string }
type UpdateAgentInput struct{ SystemPrompt string }
type BrainConfig struct{ VaultPath string }
type AgentRequest struct{ AgentID, ConversationID, Message, SystemPrompt, VaultPath string }
type AgentResponse struct {
	Content                   string
	InputTokens, OutputTokens int64
}

type AgentProvider interface {
	CreateAgent(context.Context, CreateAgentInput) (AgentResource, error)
	UpdateAgent(context.Context, string, UpdateAgentInput) error
	ConfigureBrain(context.Context, string, BrainConfig) error
	Generate(context.Context, AgentRequest) (AgentResponse, error)
	DeleteAgent(context.Context, string) error
	Health(context.Context, string) error
}
