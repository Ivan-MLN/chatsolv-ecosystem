package access

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAccessResolverScenarios(t *testing.T) {
	// Scenario A: Developer bypass - unrestricted access, no active sub required
	devResult := AccessResult{
		UserID:             "dev-user",
		WorkspaceID:        "ws-dev",
		PlatformRole:       "developer",
		AccessMode:         AccessModeDeveloper,
		SubscriptionStatus: "active",
		HasActiveSub:       true,
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
	require.Equal(t, AccessModeDeveloper, devResult.AccessMode)
	require.True(t, devResult.Entitlements.IsUnlimited)
	require.False(t, devResult.Entitlements.SubscriptionRequired)

	// Scenario B: New normal user - platform_role="user", subscription="inactive", paid features blocked
	newUserResult := AccessResult{
		UserID:             "new-user",
		WorkspaceID:        "ws-new",
		PlatformRole:       "user",
		AccessMode:         AccessModeSubscription,
		SubscriptionStatus: "inactive",
		HasActiveSub:       false,
		Entitlements: Entitlements{
			MaxAgents:            1,
			MaxChannels:          1,
			MaxDocuments:         200,
			MaxStorageBytes:      2048 * 1024 * 1024,
			MonthlyMessages:      20000,
			PublicAPI:            true,
			Webhooks:             true,
			SubscriptionRequired: true,
			IsUnlimited:          false,
		},
	}
	require.Equal(t, AccessModeSubscription, newUserResult.AccessMode)
	require.False(t, newUserResult.HasActiveSub)
	require.True(t, newUserResult.Entitlements.SubscriptionRequired)

	// Scenario C: Paid Customer - subscription="active", package quotas enforced
	paidUserResult := AccessResult{
		UserID:             "paid-user",
		WorkspaceID:        "ws-paid",
		PlatformRole:       "user",
		AccessMode:         AccessModeSubscription,
		SubscriptionStatus: "active",
		HasActiveSub:       true,
		Entitlements: Entitlements{
			MaxAgents:            1,
			MaxChannels:          1,
			MaxDocuments:         200,
			MaxStorageBytes:      2048 * 1024 * 1024,
			MonthlyMessages:      20000,
			PublicAPI:            true,
			Webhooks:             true,
			SubscriptionRequired: true,
			IsUnlimited:          false,
		},
	}
	require.True(t, paidUserResult.HasActiveSub)
	require.Equal(t, 1, paidUserResult.Entitlements.MaxAgents)
	require.Equal(t, 1, paidUserResult.Entitlements.MaxChannels)
	require.Equal(t, 200, paidUserResult.Entitlements.MaxDocuments)

	// Scenario D: Expired Customer - subscription="expired"
	expiredUserResult := AccessResult{
		UserID:             "expired-user",
		WorkspaceID:        "ws-expired",
		PlatformRole:       "user",
		AccessMode:         AccessModeSubscription,
		SubscriptionStatus: "expired",
		HasActiveSub:       false,
		Entitlements: Entitlements{
			SubscriptionRequired: true,
			IsUnlimited:          false,
		},
	}
	require.False(t, expiredUserResult.HasActiveSub)
}

func TestCacheInvalidation(t *testing.T) {
	resolver := NewResolver(nil)
	resolver.cache["user1:ws1"] = cachedAccess{
		result:    AccessResult{UserID: "user1", WorkspaceID: "ws1"},
		expiresAt: time.Now().Add(time.Minute),
	}
	require.Len(t, resolver.cache, 1)

	resolver.InvalidateCache("ws1")
	require.Len(t, resolver.cache, 0)
}
