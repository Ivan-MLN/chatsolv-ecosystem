package agentconfig

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeCanonicalRepository struct {
	agent      Agent
	role       string
	resolveErr error
	updated    Agent
}

func (f *fakeCanonicalRepository) ResolveDefaultAgent(context.Context, string, string) (Agent, string, error) {
	return f.agent, f.role, f.resolveErr
}
func (f *fakeCanonicalRepository) UpdateAgent(context.Context, Agent) (Agent, error) {
	f.updated = f.agent
	f.updated.Name = "Updated"
	return f.updated, nil
}

type fakeAgentTester struct {
	input AgentTestInput
	out   AgentTestResult
}

func (f *fakeAgentTester) Test(_ context.Context, input AgentTestInput) (AgentTestResult, error) {
	f.input = input
	return f.out, nil
}

func TestCanonicalAgentResolvesTenantDefaultAgent(t *testing.T) {
	repository := &fakeCanonicalRepository{agent: Agent{ID: "agent-a", WorkspaceID: "workspace-a", Name: "Naya", Status: "ready"}, role: "member"}
	service := NewCanonicalService(repository, nil, nil, nil)

	result, err := service.Get(context.Background(), "user-a", "workspace-a")

	require.NoError(t, err)
	require.Equal(t, "agent-a", result.ID)
}

func TestCanonicalAgentUpdateRequiresOwnerOrAdmin(t *testing.T) {
	repository := &fakeCanonicalRepository{agent: Agent{ID: "agent-a", WorkspaceID: "workspace-a", Name: "Before"}, role: "member"}
	service := NewCanonicalService(repository, nil, nil, nil)

	_, err := service.Update(context.Background(), "user-a", "workspace-a", "Updated")

	require.ErrorIs(t, err, ErrForbidden)
	require.Empty(t, repository.updated.ID)
}

func TestCanonicalAgentTestUsesTestEnvironment(t *testing.T) {
	repository := &fakeCanonicalRepository{agent: Agent{ID: "agent-a", WorkspaceID: "workspace-a", Status: "ready"}, role: "owner"}
	tester := &fakeAgentTester{out: AgentTestResult{ConversationID: "conversation-a", Content: "Bisa kak"}}
	service := NewCanonicalService(repository, nil, nil, tester)

	result, err := service.Test(context.Background(), "user-a", "workspace-a", "Bisa COD?", "", false)

	require.NoError(t, err)
	require.Equal(t, "test", tester.input.Environment)
	require.Equal(t, "agent-a", tester.input.AgentID)
	require.Equal(t, "Bisa kak", result.Content)
}

func TestCanonicalAgentRejectsMissingWorkspaceBeforeRepository(t *testing.T) {
	repository := &fakeCanonicalRepository{resolveErr: errors.New("must not be called")}
	service := NewCanonicalService(repository, nil, nil, nil)

	_, err := service.Get(context.Background(), "user-a", "")

	require.ErrorIs(t, err, ErrInvalidInput)
}
