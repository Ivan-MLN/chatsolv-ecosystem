-- name: ResolveInternalRuntimeContext :one
SELECT c.channel_id,ch.type AS channel_type,c.external_user_id
FROM conversations c
JOIN channels ch ON ch.id=c.channel_id AND ch.workspace_id=c.workspace_id AND ch.agent_id=c.agent_id
WHERE c.id=$1 AND c.agent_id=$2;

-- name: GetInternalAgentHealth :one
SELECT a.id,a.status,sb.status AS brain_status,
       (a.status='ready' AND sb.status='ready' AND a.provider_agent_id IS NOT NULL AND sb.vault_key IS NOT NULL) AS ready
FROM agents a
JOIN second_brains sb ON sb.agent_id=a.id AND sb.workspace_id=a.workspace_id AND sb.deleted_at IS NULL
WHERE a.id=$1 AND a.deleted_at IS NULL;
