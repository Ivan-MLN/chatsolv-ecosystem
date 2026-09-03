package conversation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeDashboardRepository struct {
	role string
	mode Mode
}

func (f *fakeDashboardRepository) List(context.Context, string, string, ListFilter) ([]DashboardConversation, error) {
	return []DashboardConversation{{ID: "conversation"}}, nil
}
func (f *fakeDashboardRepository) Get(context.Context, string, string) (DashboardConversation, error) {
	return DashboardConversation{ID: "conversation"}, nil
}
func (f *fakeDashboardRepository) Messages(context.Context, string, string, MessageFilter) ([]DashboardMessage, error) {
	return []DashboardMessage{{ID: "message"}}, nil
}
func (f *fakeDashboardRepository) Role(context.Context, string, string) (string, error) {
	return f.role, nil
}
func (f *fakeDashboardRepository) SetMode(_ context.Context, _, _ string, mode Mode) error {
	f.mode = mode
	return nil
}

func TestDashboardListRequiresWorkspace(t *testing.T) {
	svc := NewDashboardService(&fakeDashboardRepository{})
	_, err := svc.List(context.Background(), "user", "", ListFilter{})
	require.ErrorIs(t, err, ErrInvalidDashboardInput)
}
func TestDashboardModeRequiresManager(t *testing.T) {
	repo := &fakeDashboardRepository{role: "member"}
	svc := NewDashboardService(repo)
	err := svc.SetMode(context.Background(), "user", "conversation", ModeHuman)
	require.ErrorIs(t, err, ErrDashboardForbidden)
}
func TestDashboardModeUpdatesForAdmin(t *testing.T) {
	repo := &fakeDashboardRepository{role: "admin"}
	svc := NewDashboardService(repo)
	require.NoError(t, svc.SetMode(context.Background(), "user", "conversation", ModeHuman))
	require.Equal(t, ModeHuman, repo.mode)
}
func TestDashboardModeRejectsUnknownMode(t *testing.T) {
	svc := NewDashboardService(&fakeDashboardRepository{role: "owner"})
	require.ErrorIs(t, svc.SetMode(context.Background(), "user", "conversation", Mode("auto")), ErrInvalidDashboardInput)
}
