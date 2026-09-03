-- name: FindDefaultPublicAgent :one
SELECT a.id,a.workspace_id
FROM agents a
JOIN second_brains sb ON sb.agent_id=a.id AND sb.workspace_id=a.workspace_id
JOIN subscriptions s ON s.workspace_id=a.workspace_id
JOIN subscription_entitlements e ON e.workspace_id=a.workspace_id AND e.public_api=true
WHERE a.workspace_id=$1 AND a.status='ready' AND sb.status='ready'
  AND s.status IN('trialing','active') AND a.deleted_at IS NULL
ORDER BY a.created_at LIMIT 1;

-- name: FindPublicWebChannel :one
SELECT id FROM channels
WHERE workspace_id=$1 AND agent_id=$2 AND type='web' AND status='connected'
ORDER BY created_at LIMIT 1;

-- name: CreatePublicWebChannel :one
INSERT INTO channels(id,workspace_id,agent_id,type,display_name,status,connected_at)
VALUES($1,$2,$3,'web','Website','connected',now())
RETURNING id;

-- name: CreatePublicAPISession :one
INSERT INTO api_sessions(id,workspace_id,agent_id,external_user_id,token_hash,metadata,expires_at)
VALUES($1,$2,$3,$4,$5,$6,$7)
RETURNING id;

-- name: ResolvePublicAPISession :one
SELECT s.id,s.workspace_id,s.agent_id,s.external_user_id,s.token_hash,s.metadata,s.expires_at,c.id AS channel_id
FROM api_sessions s
JOIN channels c ON c.workspace_id=s.workspace_id AND c.agent_id=s.agent_id AND c.type='web' AND c.status='connected'
WHERE s.id=$1 AND s.token_hash=$2 AND s.expires_at>now()
ORDER BY c.created_at LIMIT 1;