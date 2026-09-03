package knowledge

import (
	"context"
	"strings"
	"testing"

	"authbackend/internal/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type knowledgeFixture struct {
	userID, workspaceID, brainID string
}

func TestCreateSourceAndQueuePersistsBothAtomically(t *testing.T) {
	pool := testutil.Postgres(t)
	fixture := createKnowledgeFixture(t, pool, "owner")
	repository := NewPostgresRepository(pool)

	source, err := repository.CreateSourceAndQueue(context.Background(), SourceRecord{
		WorkspaceID: fixture.workspaceID, SecondBrainID: fixture.brainID,
		Type: "document", Title: "Refund", MIMEType: "text/plain",
		ObjectKey: "objects/refund.txt", Checksum: strings.Repeat("a", 64), FileSize: 10,
	}, "knowledge.ingest")

	require.NoError(t, err)
	var sourceCount, eventCount int
	require.NoError(t, pool.QueryRow(context.Background(), `SELECT count(*) FROM knowledge_sources WHERE id=$1`, source.ID).Scan(&sourceCount))
	require.NoError(t, pool.QueryRow(context.Background(), `SELECT count(*) FROM outbox_events WHERE aggregate_id=$1 AND event_type='knowledge.ingest'`, source.ID).Scan(&eventCount))
	require.Equal(t, 1, sourceCount)
	require.Equal(t, 1, eventCount)
}

func TestCreateSourceAndQueuePreservesDuplicateChecksumError(t *testing.T) {
	pool := testutil.Postgres(t)
	fixture := createKnowledgeFixture(t, pool, "owner")
	repository := NewPostgresRepository(pool)
	input := SourceRecord{
		WorkspaceID: fixture.workspaceID, SecondBrainID: fixture.brainID,
		Type: "document", Title: "Refund", MIMEType: "text/plain",
		ObjectKey: "objects/refund.txt", Checksum: strings.Repeat("b", 64), FileSize: 10,
	}

	_, err := repository.CreateSourceAndQueue(context.Background(), input, "knowledge.ingest")
	require.NoError(t, err)
	_, err = repository.CreateSourceAndQueue(context.Background(), input, "knowledge.ingest")

	require.ErrorIs(t, err, ErrDocumentDuplicate)
	var sourceCount, eventCount int
	require.NoError(t, pool.QueryRow(context.Background(), `SELECT count(*) FROM knowledge_sources WHERE workspace_id=$1`, fixture.workspaceID).Scan(&sourceCount))
	require.NoError(t, pool.QueryRow(context.Background(), `SELECT count(*) FROM outbox_events WHERE workspace_id=$1`, fixture.workspaceID).Scan(&eventCount))
	require.Equal(t, 1, sourceCount)
	require.Equal(t, 1, eventCount)
}

func TestKnowledgeMutationRejectsAnotherTenant(t *testing.T) {
	pool := testutil.Postgres(t)
	owner := createKnowledgeFixture(t, pool, "owner")
	otherTenant := createKnowledgeFixture(t, pool, "owner")
	repository := NewPostgresRepository(pool)
	source, err := repository.CreateSourceAndQueue(context.Background(), SourceRecord{
		WorkspaceID: owner.workspaceID, SecondBrainID: owner.brainID,
		Type: "text", Title: "Original", MIMEType: "text/plain", ObjectKey: "inline:body",
	}, "knowledge.ingest")
	require.NoError(t, err)

	err = repository.UpdateAndQueue(context.Background(), otherTenant.userID, source.ID, "Cross-tenant update")

	require.ErrorIs(t, err, ErrForbidden)
	var title string
	var eventCount int
	require.NoError(t, pool.QueryRow(context.Background(), `SELECT title FROM knowledge_sources WHERE id=$1`, source.ID).Scan(&title))
	require.NoError(t, pool.QueryRow(context.Background(), `SELECT count(*) FROM outbox_events WHERE aggregate_id=$1`, source.ID).Scan(&eventCount))
	require.Equal(t, "Original", title)
	require.Equal(t, 1, eventCount)
}

func createKnowledgeFixture(t *testing.T, pool *pgxpool.Pool, role string) knowledgeFixture {
	t.Helper()
	ctx := context.Background()
	userID := uuid.NewString()
	workspaceID := uuid.NewString()
	agentID := uuid.NewString()
	brainID := uuid.NewString()
	_, err := pool.Exec(ctx, `INSERT INTO users(id,name,email,password_hash) VALUES($1,'Test User',$2,'hash')`, userID, userID+"@example.com")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO workspaces(id,name,slug,owner_user_id,status,timezone) VALUES($1,'Test Workspace',$2,$3,'active','UTC')`, workspaceID, "workspace-"+workspaceID, userID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO workspace_members(id,workspace_id,user_id,role) VALUES($1,$2,$3,$4)`, uuid.NewString(), workspaceID, userID, role)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO agents(id,workspace_id,name,status,provider) VALUES($1,$2,'Test Agent','ready','hermes')`, agentID, workspaceID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO second_brains(id,workspace_id,agent_id,provider,vault_key,status) VALUES($1,$2,$3,'obsidian',$4,'ready')`, brainID, workspaceID, agentID, "vault-"+brainID)
	require.NoError(t, err)
	return knowledgeFixture{userID: userID, workspaceID: workspaceID, brainID: brainID}
}
