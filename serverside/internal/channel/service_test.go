package channel

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeRepository struct {
	role       string
	agentID    string
	channels   []Channel
	deletedID  string
	deleteUser string
	count      int64
	maximum    int64
}

func (f *fakeRepository) Authorize(context.Context, string, string) (string, string, error) {
	return f.role, f.agentID, nil
}
func (f *fakeRepository) AuthorizeMutation(context.Context, string, string) error { return nil }
func (f *fakeRepository) Count(context.Context, string) (int64, error)            { return f.count, nil }
func (f *fakeRepository) Max(context.Context, string) (int64, error) {
	if f.maximum == 0 {
		return 1, nil
	}
	return f.maximum, nil
}
func (f *fakeRepository) Create(_ context.Context, value Channel) (Channel, error) {
	value.ID = "channel-a"
	return value, nil
}
func (f *fakeRepository) List(context.Context, string, string) ([]Channel, error) {
	return f.channels, nil
}
func (f *fakeRepository) UpdateStatus(context.Context, string, string, string, string) error {
	return nil
}
func (f *fakeRepository) Delete(_ context.Context, channelID, userID string) error {
	f.deletedID, f.deleteUser = channelID, userID
	return nil
}

type fakeBot struct{ connectErr error }

func (f fakeBot) Connect(context.Context, string, string) (ConnectResponse, error) {
	if f.connectErr != nil {
		return ConnectResponse{}, f.connectErr
	}
	return ConnectResponse{SessionID: "session-a", Status: "waiting_pairing", QR: "qr-data"}, nil
}
func (fakeBot) Disconnect(context.Context, string) error { return nil }
func (fakeBot) GetProfile(context.Context, string) (WhatsAppProfile, error) {
	return WhatsAppProfile{PushName: "Test Bot", Phone: "628123456789"}, nil
}
func (fakeBot) SendTextMessage(context.Context, string, string, string) error { return nil }

func TestListChannelsUsesTenantMembership(t *testing.T) {
	repository := &fakeRepository{role: "member", agentID: "agent-a", channels: []Channel{{ID: "channel-a", WorkspaceID: "workspace-a"}}}
	service := NewService(repository, fakeBot{})

	result, err := service.List(context.Background(), "user-a", "workspace-a")

	require.NoError(t, err)
	require.Len(t, result, 1)
}

func TestDeleteChannelDelegatesTenantScopedMutation(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, fakeBot{})

	err := service.Delete(context.Background(), "user-a", "channel-a")

	require.NoError(t, err)
	require.Equal(t, "channel-a", repository.deletedID)
	require.Equal(t, "user-a", repository.deleteUser)
}

func TestConnectWhatsAppRequiresDisplayName(t *testing.T) {
	service := NewService(&fakeRepository{role: "owner", agentID: "agent-a"}, fakeBot{})

	_, _, err := service.ConnectWhatsApp(context.Background(), "user-a", "workspace-a", " ", "")

	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestConnectWhatsAppCleansUpChannelWhenBotFails(t *testing.T) {
	repository := &fakeRepository{role: "owner", agentID: "agent-a"}
	service := NewService(repository, fakeBot{connectErr: errors.New("bot unavailable")})

	_, _, err := service.ConnectWhatsApp(context.Background(), "user-a", "workspace-a", "Customer Service", "628123456789")

	require.ErrorIs(t, err, ErrConnectionFailed)
	require.Equal(t, "channel-a", repository.deletedID)
	require.Equal(t, "user-a", repository.deleteUser)
}

func TestDeveloperChannelBypassSkipsCommercialQuota(t *testing.T) {
	repository := &fakeRepository{role: "owner", agentID: "agent-a", count: 1, maximum: 1}
	service := NewService(repository, fakeBot{})

	_, _, err := service.ConnectWhatsApp(context.Background(), "user-a", "workspace-a", "Customer Service", "")
	require.ErrorIs(t, err, ErrQuotaExceeded)

	_, _, err = service.ConnectWhatsAppWithBypass(context.Background(), "user-a", "workspace-a", "Customer Service", "", true)
	require.NoError(t, err)
}
