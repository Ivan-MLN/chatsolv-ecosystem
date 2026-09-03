-- name: GetCurrentUserMemberships :many
SELECT u.id AS user_id,u.name AS user_name,u.email,u.platform_role,u.created_at AS user_created_at,
       w.id AS workspace_id,w.name AS workspace_name,w.slug,w.status,w.timezone,wm.role
FROM users u
LEFT JOIN workspace_members wm ON wm.user_id=u.id
LEFT JOIN workspaces w ON w.id=wm.workspace_id AND w.deleted_at IS NULL
WHERE u.id=$1
ORDER BY wm.created_at,w.id;

-- name: GetDashboardOverview :one
WITH authorized_workspace AS (
    SELECT w.id
    FROM workspaces w
    JOIN workspace_members wm ON wm.workspace_id=w.id
    WHERE w.id=$1 AND wm.user_id=$2 AND w.deleted_at IS NULL
),
agent_state AS (
    SELECT COALESCE((SELECT a.status FROM agents a JOIN authorized_workspace aw ON aw.id=a.workspace_id WHERE a.deleted_at IS NULL ORDER BY a.created_at LIMIT 1),'not_configured')::text AS status
),
brain_state AS (
    SELECT
      COALESCE((SELECT sb.status FROM second_brains sb JOIN authorized_workspace aw ON aw.id=sb.workspace_id WHERE sb.deleted_at IS NULL ORDER BY sb.created_at LIMIT 1),'not_configured')::text AS status,
      (SELECT count(*) FROM knowledge_sources ks JOIN authorized_workspace aw ON aw.id=ks.workspace_id WHERE ks.deleted_at IS NULL AND ks.status NOT IN('deleting','deleted'))::bigint AS knowledge_sources
),
channel_state AS (
    SELECT COALESCE((SELECT c.status FROM channels c JOIN authorized_workspace aw ON aw.id=c.workspace_id WHERE c.type='whatsapp' ORDER BY (c.status='connected') DESC,c.created_at LIMIT 1),'not_connected')::text AS status
),
conversation_state AS (
    SELECT
      (SELECT count(*) FROM conversations c JOIN authorized_workspace aw ON aw.id=c.workspace_id WHERE c.environment='production' AND c.hidden_at IS NULL AND c.started_at>=date_trunc('day',now()))::bigint AS today,
      (SELECT count(*) FROM conversations c JOIN authorized_workspace aw ON aw.id=c.workspace_id WHERE c.environment='production' AND c.hidden_at IS NULL AND c.status='open')::bigint AS open
)
SELECT aw.id AS workspace_id,a.status AS agent_status,b.status AS second_brain_status,b.knowledge_sources,
       ch.status AS channel_status,cv.today,cv.open
FROM authorized_workspace aw
CROSS JOIN agent_state a
CROSS JOIN brain_state b
CROSS JOIN channel_state ch
CROSS JOIN conversation_state cv;
