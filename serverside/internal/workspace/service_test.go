package workspace

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeRepository struct {
	created             Workspace
	createdMembership   Membership
	createdSubscription Subscription
	createdEntitlement  Entitlement
	createdAgent        AgentSeed
	createdBrain        BrainSeed
	createdEvent        OutboxSeed
	workspace           Workspace
	membership          Membership
	getErr, updateErr   error
}

func (f *fakeRepository) CreateWithOwnerTrial(_ context.Context, w Workspace, m Membership, s Subscription, e Entitlement, a AgentSeed, b BrainSeed, event OutboxSeed) error {
	f.created = w
	f.createdMembership = m
	f.createdSubscription = s
	f.createdEntitlement = e
	f.createdAgent = a
	f.createdBrain = b
	f.createdEvent = event
	return nil
}

func (f *fakeRepository) GetForMember(_ context.Context, workspaceID, userID string) (Workspace, Membership, error) {
	if f.getErr != nil {
		return Workspace{}, Membership{}, f.getErr
	}
	if f.workspace.ID != workspaceID || f.membership.UserID != userID {
		return Workspace{}, Membership{}, ErrNotFound
	}
	return f.workspace, f.membership, nil
}

func (f *fakeRepository) Update(_ context.Context, w Workspace) error {
	f.workspace = w
	return f.updateErr
}

func TestCreateCreatesWorkspaceAndSubscription(t *testing.T) {
	repo := &fakeRepository{}
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	svc := NewService(repo, nil, func() time.Time { return now })

	result, err := svc.Create(context.Background(), "user-1", CreateInput{
		Name: " Toko Maju ", Slug: "Toko Maju", Timezone: "Asia/Jakarta",
	})

	require.NoError(t, err)
	require.Equal(t, "Toko Maju", repo.created.Name)
	require.Equal(t, "toko-maju", repo.created.Slug)
	require.Equal(t, StatusProvisioning, repo.created.Status)
	require.Equal(t, RoleOwner, repo.createdMembership.Role)
	require.Equal(t, "user-1", repo.createdMembership.UserID)
	require.Equal(t, SubscriptionInactive, repo.createdSubscription.Status)
	require.Equal(t, 1, repo.createdEntitlement.MaxAgents)
	require.Equal(t, 1, repo.createdEntitlement.MaxChannels)
	require.Equal(t, int64(20_000), repo.createdEntitlement.MonthlyMessages)
	require.True(t, repo.createdEntitlement.PublicAPI)
	require.Equal(t, repo.created.ID, repo.createdAgent.WorkspaceID)
	require.Equal(t, repo.createdAgent.ID, repo.createdBrain.AgentID)
	require.Equal(t, "workspace.provision", repo.createdEvent.Type)
	require.Equal(t, repo.created.ID, repo.createdEvent.AggregateID)
	require.Equal(t, repo.created.ID, result.Workspace.ID)
}

func TestGetRejectsCrossTenantAccess(t *testing.T) {
	repo := &fakeRepository{
		workspace:  Workspace{ID: "workspace-b"},
		membership: Membership{WorkspaceID: "workspace-b", UserID: "user-b", Role: RoleOwner},
	}

	_, err := NewService(repo, nil, time.Now).Get(context.Background(), "user-a", "workspace-b")

	require.ErrorIs(t, err, ErrNotFound)
}

func TestUpdateRequiresOwnerOrAdmin(t *testing.T) {
	for _, role := range []Role{RoleMember, RoleViewer} {
		t.Run(string(role), func(t *testing.T) {
			repo := &fakeRepository{
				workspace:  Workspace{ID: "workspace-1", Name: "Before", Slug: "before", Timezone: "UTC"},
				membership: Membership{WorkspaceID: "workspace-1", UserID: "user-1", Role: role},
			}
			_, err := NewService(repo, nil, time.Now).Update(context.Background(), "user-1", "workspace-1", UpdateInput{Name: "After"})
			require.ErrorIs(t, err, ErrForbidden)
			require.Equal(t, "Before", repo.workspace.Name)
		})
	}
}
