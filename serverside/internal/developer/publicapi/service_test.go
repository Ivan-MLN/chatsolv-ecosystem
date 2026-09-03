package publicapi

import (
	"context"
	"testing"

	"authbackend/internal/conversation"
	"authbackend/internal/developer/apikey"

	"github.com/stretchr/testify/require"
)

type fakeKeyAuthenticator struct{ record apikey.Record }

func (f fakeKeyAuthenticator) Authenticate(context.Context, string, string) (apikey.Record, error) {
	return f.record, nil
}

type fakeSessionRepository struct {
	created  Session
	resolved Session
}

func (f *fakeSessionRepository) Create(_ context.Context, session Session) error {
	f.created = session
	return nil
}
func (f *fakeSessionRepository) Resolve(context.Context, string, string) (Session, error) {
	return f.resolved, nil
}

type fakeConversation struct{ incoming conversation.Incoming }

func (f *fakeConversation) Handle(_ context.Context, incoming conversation.Incoming) (conversation.Result, error) {
	f.incoming = incoming
	return conversation.Result{ConversationID: "conversation", MessageID: "message", Content: "hello"}, nil
}

func TestCreateSessionStoresOnlyClientTokenHash(t *testing.T) {
	repo := &fakeSessionRepository{}
	svc := NewService(fakeKeyAuthenticator{record: apikey.Record{WorkspaceID: "workspace"}}, repo, &fakeConversation{})

	created, err := svc.CreateSession(context.Background(), "cs_live_secret", CreateSessionInput{ExternalUserID: "visitor-1"})

	require.NoError(t, err)
	require.NotEmpty(t, created.ClientToken)
	require.NotEqual(t, created.ClientToken, repo.created.TokenHash)
	require.NotContains(t, repo.created.TokenHash, created.ClientToken)
	require.Equal(t, int64(3600), created.ExpiresIn)
}

func TestSendMessageUsesResolvedTenantSession(t *testing.T) {
	repo := &fakeSessionRepository{resolved: Session{ID: "session", WorkspaceID: "workspace", AgentID: "agent", ChannelID: "channel", ExternalUserID: "visitor"}}
	conversationService := &fakeConversation{}
	svc := NewService(fakeKeyAuthenticator{}, repo, conversationService)

	result, err := svc.SendMessage(context.Background(), "session", "client-token", "Do you support stores?")

	require.NoError(t, err)
	require.Equal(t, "production", conversationService.incoming.Environment)
	require.Equal(t, "channel", conversationService.incoming.ChannelID)
	require.Equal(t, "visitor", conversationService.incoming.ExternalUserID)
	require.Equal(t, "web", conversationService.incoming.ChannelType)
	require.Equal(t, "hello", result.Content)
}

func TestSendMessageRejectsBlankMessage(t *testing.T) {
	svc := NewService(fakeKeyAuthenticator{}, &fakeSessionRepository{}, &fakeConversation{})
	_, err := svc.SendMessage(context.Background(), "session", "token", "   ")
	require.ErrorIs(t, err, ErrInvalidInput)
}
