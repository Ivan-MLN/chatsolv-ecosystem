package internalapi

import (
	"context"
	"testing"

	"authbackend/internal/conversation"

	"github.com/stretchr/testify/require"
)

type fakeRepository struct {
	status  string
	context RuntimeContext
	health  AgentHealth
}

func (f *fakeRepository) UpdateChannelStatus(_ context.Context, _, status, _, _ string) error {
	f.status = status
	return nil
}
func (f *fakeRepository) ResolveRuntimeContext(context.Context, string, string) (RuntimeContext, error) {
	return f.context, nil
}
func (f *fakeRepository) GetAgentHealth(context.Context, string) (AgentHealth, error) {
	return f.health, nil
}

type fakeConversationService struct{ incoming conversation.Incoming }

func (f *fakeConversationService) Handle(_ context.Context, in conversation.Incoming) (conversation.Result, error) {
	f.incoming = in
	return conversation.Result{Content: "reply"}, nil
}

func TestIncomingMessageUsesWhatsAppRuntime(t *testing.T) {
	runtime := &fakeConversationService{}
	svc := NewService(&fakeRepository{}, runtime)
	result, err := svc.Incoming(context.Background(), IncomingMessage{ChannelID: "channel", ExternalMessageID: "wamid", ExternalUserID: "62812", MessageType: "text", Content: MessageContent{Text: "hello"}})
	require.NoError(t, err)
	require.Equal(t, "whatsapp", runtime.incoming.ChannelType)
	require.Equal(t, "production", runtime.incoming.Environment)
	require.Equal(t, "reply", result.Content)
}
func TestRespondResolvesConversationContext(t *testing.T) {
	runtime := &fakeConversationService{}
	repo := &fakeRepository{context: RuntimeContext{ChannelID: "channel", ChannelType: "web", ExternalUserID: "visitor"}}
	svc := NewService(repo, runtime)
	_, err := svc.Respond(context.Background(), "agent", RespondInput{ConversationID: "conversation", Message: "hello"})
	require.NoError(t, err)
	require.Equal(t, "channel", runtime.incoming.ChannelID)
	require.Equal(t, "web", runtime.incoming.ChannelType)
	require.Equal(t, "visitor", runtime.incoming.ExternalUserID)
}
func TestChannelStatusRejectsUnknownState(t *testing.T) {
	svc := NewService(&fakeRepository{}, &fakeConversationService{})
	err := svc.ChannelStatus(context.Background(), ChannelStatusInput{ChannelID: "channel", Status: "invalid"})
	require.ErrorIs(t, err, ErrInvalidInput)
}
func TestChannelEventTranslatesPairSuccess(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewService(repo, &fakeConversationService{})
	err := svc.ChannelEvent(context.Background(), ChannelEventInput{ChannelID: "channel", Event: "pair_success", PhoneNumber: "62812"})
	require.NoError(t, err)
	require.Equal(t, "connected", repo.status)
}
func TestAgentHealthReturnsRepositoryState(t *testing.T) {
	svc := NewService(&fakeRepository{health: AgentHealth{Status: "ready", Ready: true}}, &fakeConversationService{})
	health, err := svc.Health(context.Background(), "agent")
	require.NoError(t, err)
	require.True(t, health.Ready)
}
