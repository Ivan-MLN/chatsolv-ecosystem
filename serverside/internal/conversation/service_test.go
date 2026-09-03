package conversation

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

type fakeRepo struct {
	duplicate                 *Result
	savedCustomer, savedAgent bool
	handoff                   bool
	mode                      Mode
}

func (f *fakeRepo) FindResult(context.Context, string, string) (*Result, error) {
	return f.duplicate, nil
}
func (f *fakeRepo) ResolveOrCreate(context.Context, Incoming) (Conversation, error) {
	return Conversation{ID: "cnv", WorkspaceID: "wsp", AgentID: "agt", Mode: f.mode}, nil
}
func (f *fakeRepo) SaveCustomer(context.Context, Conversation, Incoming) error {
	f.savedCustomer = true
	return nil
}
func (f *fakeRepo) RecentMessages(context.Context, string, int) ([]Message, error) { return nil, nil }
func (f *fakeRepo) SaveAgent(_ context.Context, _ Conversation, content string) (Result, error) {
	f.savedAgent = true
	return Result{MessageID: "msg", ConversationID: "cnv", Content: content}, nil
}
func (f *fakeRepo) RequestHandoff(context.Context, Conversation, string) error {
	f.handoff = true
	return nil
}

type fakeRuntime struct{ calls int }

func (f *fakeRuntime) Generate(context.Context, RuntimeInput) (RuntimeOutput, error) {
	f.calls++
	return RuntimeOutput{Content: "Bisa kak"}, nil
}

type handoffRuntime struct{}

func (handoffRuntime) Generate(context.Context, RuntimeInput) (RuntimeOutput, error) {
	return RuntimeOutput{Content: "I will connect you", HandoffRequested: true, HandoffReason: "CUSTOMER_REQUEST"}, nil
}

type fakeLock struct{}

func (fakeLock) WithLock(ctx context.Context, _ string, fn func(context.Context) error) error {
	return fn(ctx)
}
func TestIncomingIsIdempotent(t *testing.T) {
	existing := &Result{MessageID: "old", Content: "previous"}
	repo := &fakeRepo{duplicate: existing}
	runtime := &fakeRuntime{}
	result, err := NewService(repo, runtime, fakeLock{}).Handle(context.Background(), Incoming{ChannelID: "chn", ExternalMessageID: "wamid"})
	require.NoError(t, err)
	require.Equal(t, *existing, result)
	require.Zero(t, runtime.calls)
}
func TestHumanModeDoesNotInvokeAgent(t *testing.T) {
	repo := &fakeRepo{mode: ModeHuman}
	runtime := &fakeRuntime{}
	_, err := NewService(repo, runtime, fakeLock{}).Handle(context.Background(), Incoming{ChannelID: "chn", ExternalMessageID: "wamid", Text: "admin"})
	require.ErrorIs(t, err, ErrHumanMode)
	require.Zero(t, runtime.calls)
	require.True(t, repo.savedCustomer)
}
func TestAgentModePersistsBothMessages(t *testing.T) {
	repo := &fakeRepo{mode: ModeAgent}
	runtime := &fakeRuntime{}
	result, err := NewService(repo, runtime, fakeLock{}).Handle(context.Background(), Incoming{ChannelID: "chn", ExternalMessageID: "wamid", Text: "COD?"})
	require.NoError(t, err)
	require.Equal(t, "Bisa kak", result.Content)
	require.True(t, repo.savedCustomer)
	require.True(t, repo.savedAgent)
	require.Equal(t, 1, runtime.calls)
}
func TestRequestedHandoffChangesConversationMode(t *testing.T) {
	repo := &fakeRepo{mode: ModeAgent}
	result, err := NewService(repo, handoffRuntime{}, fakeLock{}).Handle(context.Background(), Incoming{ChannelID: "chn", ExternalMessageID: "wamid", Text: "admin"})
	require.NoError(t, err)
	require.True(t, repo.handoff)
	require.True(t, result.HandoffRequested)
}
