package hermes

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type recordedCommand struct {
	name string
	args []string
	env  []string
	dir  string
}
type fakeRunner struct {
	calls  []recordedCommand
	output []byte
	err    error
}

func (f *fakeRunner) Run(_ context.Context, name string, args []string, env []string, dir string) ([]byte, error) {
	f.calls = append(f.calls, recordedCommand{name: name, args: append([]string(nil), args...), env: append([]string(nil), env...), dir: dir})
	return f.output, f.err
}

func TestCLIProviderCreatesDedicatedProfileNameWithoutRawPathInput(t *testing.T) {
	r := &fakeRunner{}
	root := t.TempDir()
	provider := NewCLIProvider("hermes", root, "default", r)
	resource, err := provider.CreateAgent(context.Background(), CreateAgentInput{WorkspaceID: "550e8400-e29b-41d4-a716-446655440000", AgentID: "agt", VaultPath: "/vaults/tenant"})
	require.NoError(t, err)
	require.Equal(t, "cs550e8400e29b41d4a716446655440000", resource.ID)
	require.Equal(t, []string{"profile", "create", resource.ID, "--clone-from", "default", "--no-alias", "--no-skills"}, r.calls[0].args)
}

func TestCLIProviderGenerateUsesOnlyResolvedTenantProfileAndVault(t *testing.T) {
	r := &fakeRunner{output: []byte("Jawaban tenant\n")}
	root := t.TempDir()
	profile := "cs550e8400e29b41d4a716446655440000"
	profileHome := filepath.Join(root, "profiles", profile)
	require.NoError(t, os.MkdirAll(profileHome, 0o750))
	provider := NewCLIProvider("hermes", root, "default", r)
	provider.vaults[profile] = "/vaults/wsp_a"
	response, err := provider.Generate(context.Background(), AgentRequest{AgentID: profile, ConversationID: "cnv", Message: "Halo", SystemPrompt: "rules"})
	require.NoError(t, err)
	require.Equal(t, "Jawaban tenant", response.Content)
	require.Equal(t, "/vaults/wsp_a", r.calls[0].dir)
	require.Contains(t, r.calls[0].env, "HERMES_HOME="+profileHome)
	require.Contains(t, r.calls[0].args, "--safe-mode")
	require.Equal(t, "-z", r.calls[0].args[0])
	require.NotContains(t, r.calls[0].args, "--oneshot")
	require.NotContains(t, r.calls[0].args, "--in")
}

func TestCLIProviderRejectsUnknownOrEscapingProfile(t *testing.T) {
	provider := NewCLIProvider("hermes", t.TempDir(), "default", &fakeRunner{})
	_, err := provider.Generate(context.Background(), AgentRequest{AgentID: "../../default", Message: "x"})
	require.Error(t, err)
}
