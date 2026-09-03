package access

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccessResultEntitlements(t *testing.T) {
	devAccess := AccessResult{
		UserID:       "user-dev",
		WorkspaceID:  "ws-1",
		PlatformRole: "developer",
		AccessMode:   AccessModeDeveloper,
		HasActiveSub: true,
		Entitlements: Entitlements{
			MaxAgents:            -1,
			MaxChannels:          -1,
			MaxDocuments:         -1,
			MaxStorageBytes:      -1,
			MonthlyMessages:      -1,
			PublicAPI:            true,
			Webhooks:             true,
			SubscriptionRequired: false,
			IsUnlimited:          true,
		},
	}

	require.True(t, devAccess.Entitlements.IsUnlimited)
	require.False(t, devAccess.Entitlements.SubscriptionRequired)
	require.Equal(t, -1, devAccess.Entitlements.MaxAgents)

	normalInactiveAccess := AccessResult{
		UserID:       "user-norm",
		WorkspaceID:  "ws-2",
		PlatformRole: "user",
		AccessMode:   AccessModeSubscription,
		HasActiveSub: false,
		Entitlements: Entitlements{
			MaxAgents:            1,
			MaxChannels:          1,
			MaxDocuments:         200,
			MaxStorageBytes:      2 * 1024 * 1024 * 1024,
			MonthlyMessages:      20000,
			PublicAPI:            true,
			Webhooks:             true,
			SubscriptionRequired: true,
			IsUnlimited:          false,
		},
	}

	require.False(t, normalInactiveAccess.Entitlements.IsUnlimited)
	require.True(t, normalInactiveAccess.Entitlements.SubscriptionRequired)
	require.False(t, normalInactiveAccess.HasActiveSub)
}
