package dashboard

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeRepository struct {
	me             Me
	overview       Overview
	meErr          error
	overviewErr    error
	gotUserID      string
	gotWorkspaceID string
}

func (f *fakeRepository) GetMe(_ context.Context, userID string) (Me, error) {
	f.gotUserID = userID
	return f.me, f.meErr
}

func (f *fakeRepository) GetOverview(_ context.Context, userID, workspaceID string) (Overview, error) {
	f.gotUserID = userID
	f.gotWorkspaceID = workspaceID
	return f.overview, f.overviewErr
}

func TestMeReturnsIdentityAndAllWorkspaceMemberships(t *testing.T) {
	repository := &fakeRepository{me: Me{
		User:       User{ID: "user-a", Name: "Ayu", Email: "ayu@example.com"},
		Workspaces: []WorkspaceMembership{{WorkspaceID: "workspace-a", Name: "Toko A", Role: "owner"}},
	}}
	service := NewService(repository)

	result, err := service.Me(context.Background(), "user-a")

	require.NoError(t, err)
	require.Equal(t, "user-a", repository.gotUserID)
	require.Equal(t, "ayu@example.com", result.User.Email)
	require.Len(t, result.Workspaces, 1)
}

func TestDashboardRequiresExplicitWorkspace(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.Overview(context.Background(), "user-a", "")

	require.ErrorIs(t, err, ErrInvalidWorkspace)
	require.Empty(t, repository.gotUserID)
}

func TestDashboardReturnsTenantScopedAggregate(t *testing.T) {
	repository := &fakeRepository{overview: Overview{
		WorkspaceID:   "workspace-a",
		Agent:         ResourceStatus{Status: "ready"},
		SecondBrain:   SecondBrainStatus{Status: "ready", KnowledgeSources: 16},
		Channel:       ResourceStatus{Status: "connected"},
		Conversations: ConversationSummary{Today: 42, Open: 8},
	}}
	service := NewService(repository)

	result, err := service.Overview(context.Background(), "user-a", "workspace-a")

	require.NoError(t, err)
	require.Equal(t, "user-a", repository.gotUserID)
	require.Equal(t, "workspace-a", repository.gotWorkspaceID)
	require.Equal(t, int64(16), result.SecondBrain.KnowledgeSources)
	require.Equal(t, int64(42), result.Conversations.Today)
}

func TestDashboardPreservesRepositoryAuthorizationFailure(t *testing.T) {
	denied := errors.New("workspace not found")
	service := NewService(&fakeRepository{overviewErr: denied})

	_, err := service.Overview(context.Background(), "user-a", "workspace-b")

	require.ErrorIs(t, err, denied)
}
