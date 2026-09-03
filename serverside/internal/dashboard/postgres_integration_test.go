package dashboard

import (
	"context"
	"testing"

	"authbackend/internal/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestPostgresRepositoryReturnsTenantScopedMeAndDashboard(t *testing.T) {
	pool := testutil.Postgres(t)
	ctx := context.Background()
	userA, workspaceA := createDashboardFixture(t, pool, "A")
	_, workspaceB := createDashboardFixture(t, pool, "B")
	repository := NewPostgresRepository(pool)

	me, err := repository.GetMe(ctx, userA)
	require.NoError(t, err)
	require.Len(t, me.Workspaces, 1)
	require.Equal(t, workspaceA, me.Workspaces[0].WorkspaceID)

	overview, err := repository.GetOverview(ctx, userA, workspaceA)
	require.NoError(t, err)
	require.Equal(t, workspaceA, overview.WorkspaceID)

	_, err = repository.GetOverview(ctx, userA, workspaceB)
	require.ErrorIs(t, err, ErrNotFound)
}

func createDashboardFixture(t *testing.T, pool *pgxpool.Pool, suffix string) (string, string) {
	t.Helper()
	ctx := context.Background()
	userID := uuid.NewString()
	workspaceID := uuid.NewString()
	agentID := uuid.NewString()
	brainID := uuid.NewString()
	_, err := pool.Exec(ctx, `INSERT INTO users(id,name,email,password_hash) VALUES($1,$2,$3,'hash')`, userID, "User "+suffix, userID+"@example.com")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO workspaces(id,name,slug,owner_user_id,status,timezone) VALUES($1,$2,$3,$4,'active','UTC')`, workspaceID, "Workspace "+suffix, "workspace-"+workspaceID, userID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO workspace_members(id,workspace_id,user_id,role) VALUES($1,$2,$3,'owner')`, uuid.NewString(), workspaceID, userID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO agents(id,workspace_id,name,status,provider) VALUES($1,$2,'Agent','ready','hermes')`, agentID, workspaceID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO second_brains(id,workspace_id,agent_id,provider,vault_key,status) VALUES($1,$2,$3,'obsidian',$4,'ready')`, brainID, workspaceID, agentID, "vault-"+brainID)
	require.NoError(t, err)
	return userID, workspaceID
}
