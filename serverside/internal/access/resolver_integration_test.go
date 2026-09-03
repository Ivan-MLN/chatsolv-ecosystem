package access

import (
	"context"
	"testing"

	"authbackend/internal/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestResolverDeveloperCannotAccessAnotherWorkspace(t *testing.T) {
	pool := testutil.Postgres(t)
	ctx := context.Background()
	developerID := uuid.New()
	ownerID := uuid.New()
	workspaceID := uuid.New()

	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, name, email, password_hash, platform_role)
		VALUES ($1, 'Developer', $2, 'hash', 'developer'),
		       ($3, 'Owner', $4, 'hash', 'user')`,
		developerID, developerID.String()+"@example.test",
		ownerID, ownerID.String()+"@example.test",
	)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO workspaces (id, name, slug, owner_user_id, status, timezone)
		VALUES ($1, 'Private workspace', $2, $3, 'active', 'Asia/Jakarta')`,
		workspaceID, "private-"+workspaceID.String(), ownerID,
	)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_members (id, workspace_id, user_id, role)
		VALUES ($1, $2, $3, 'owner')`, uuid.New(), workspaceID, ownerID)
	require.NoError(t, err)

	_, err = NewResolver(pool).Resolve(ctx, developerID.String(), workspaceID.String())
	require.ErrorContains(t, err, "workspace member not found")
}

func TestResolverDeveloperRetainsBypassInsideOwnWorkspace(t *testing.T) {
	pool := testutil.Postgres(t)
	ctx := context.Background()
	developerID := uuid.New()
	workspaceID := uuid.New()

	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, name, email, password_hash, platform_role)
		VALUES ($1, 'Developer', $2, 'hash', 'developer')`,
		developerID, developerID.String()+"@example.test",
	)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO workspaces (id, name, slug, owner_user_id, status, timezone)
		VALUES ($1, 'Developer workspace', $2, $3, 'active', 'Asia/Jakarta')`,
		workspaceID, "developer-"+workspaceID.String(), developerID,
	)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_members (id, workspace_id, user_id, role)
		VALUES ($1, $2, $3, 'owner')`, uuid.New(), workspaceID, developerID)
	require.NoError(t, err)

	result, err := NewResolver(pool).Resolve(ctx, developerID.String(), workspaceID.String())
	require.NoError(t, err)
	require.Equal(t, AccessModeDeveloper, result.AccessMode)
	require.True(t, result.Entitlements.IsUnlimited)
	require.Equal(t, "owner", result.WorkspaceRole)
}
