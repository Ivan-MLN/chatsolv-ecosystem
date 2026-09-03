package provisioning

import (
	"context"
	"errors"
	"testing"

	"authbackend/internal/brain/obsidian"
	"authbackend/internal/hermes"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	resource          Resource
	completed, failed bool
}

func (f *fakeRepo) Get(context.Context, string) (Resource, error)  { return f.resource, nil }
func (f *fakeRepo) MarkProvisioning(context.Context, string) error { return nil }
func (f *fakeRepo) Complete(_ context.Context, _ string, providerID, vaultKey string) error {
	f.completed = true
	f.resource.ProviderAgentID = providerID
	f.resource.VaultKey = vaultKey
	return nil
}
func (f *fakeRepo) Fail(context.Context, string, string) error { f.failed = true; return nil }

type fakeBrain struct {
	vault obsidian.Vault
	calls int
}

func (f *fakeBrain) CreateVault(context.Context, string) (obsidian.Vault, error) {
	f.calls++
	return f.vault, nil
}
func (f *fakeBrain) WriteNote(context.Context, string, obsidian.Note) error { return nil }
func (f *fakeBrain) ReadNote(context.Context, string, string) (obsidian.Note, error) {
	return obsidian.Note{}, nil
}
func (f *fakeBrain) DeleteNote(context.Context, string, string) error           { return nil }
func (f *fakeBrain) ListNotes(context.Context, string) ([]obsidian.Note, error) { return nil, nil }
func (f *fakeBrain) DeleteVault(context.Context, string) error                  { return nil }

type fakeHermes struct {
	calls int
	err   error
}

func (f *fakeHermes) CreateAgent(context.Context, hermes.CreateAgentInput) (hermes.AgentResource, error) {
	f.calls++
	return hermes.AgentResource{ID: "tenant-profile"}, f.err
}
func (f *fakeHermes) UpdateAgent(context.Context, string, hermes.UpdateAgentInput) error { return nil }
func (f *fakeHermes) ConfigureBrain(context.Context, string, hermes.BrainConfig) error   { return nil }
func (f *fakeHermes) Generate(context.Context, hermes.AgentRequest) (hermes.AgentResponse, error) {
	return hermes.AgentResponse{}, nil
}
func (f *fakeHermes) DeleteAgent(context.Context, string) error { return nil }
func (f *fakeHermes) Health(context.Context, string) error      { return nil }

func TestProvisionCreatesDedicatedBrainAndHermesResource(t *testing.T) {
	repo := &fakeRepo{resource: Resource{WorkspaceID: "wsp_a", AgentID: "agt_a", SecondBrainID: "brn_a"}}
	brain := &fakeBrain{vault: obsidian.Vault{ID: "wsp_a", Path: "/vaults/wsp_a"}}
	h := &fakeHermes{}
	err := NewService(repo, brain, h).Provision(context.Background(), "wsp_a")
	require.NoError(t, err)
	require.True(t, repo.completed)
	require.Equal(t, 1, brain.calls)
	require.Equal(t, 1, h.calls)
}

func TestProvisionMarksFailureWithoutCreatingDuplicateResource(t *testing.T) {
	repo := &fakeRepo{resource: Resource{WorkspaceID: "wsp_a", AgentID: "agt_a", SecondBrainID: "brn_a"}}
	h := &fakeHermes{err: errors.New("hermes unavailable")}
	err := NewService(repo, &fakeBrain{vault: obsidian.Vault{ID: "wsp_a"}}, h).Provision(context.Background(), "wsp_a")
	require.Error(t, err)
	require.True(t, repo.failed)
	require.False(t, repo.completed)
}
