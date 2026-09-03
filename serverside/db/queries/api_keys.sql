-- name: CreateAPIKey :one
INSERT INTO api_keys(id,workspace_id,prefix,hash,last_four,name,scopes)
VALUES($1,$2,$3,$4,$5,$6,$7)
RETURNING id,workspace_id,prefix,hash,last_four,name,scopes,created_at,last_used_at,revoked_at;

-- name: FindAPIKeyByPrefix :one
SELECT id,workspace_id,prefix,hash,last_four,name,scopes,created_at,last_used_at,revoked_at
FROM api_keys WHERE prefix=$1;

-- name: AuthorizeAPIKeyWorkspace :one
SELECT role FROM workspace_members WHERE workspace_id=$1 AND user_id=$2;

-- name: ListAPIKeysForManager :many
SELECT k.id,k.workspace_id,k.prefix,k.hash,k.last_four,k.name,k.scopes,k.created_at,k.last_used_at,k.revoked_at
FROM api_keys k JOIN workspace_members wm ON wm.workspace_id=k.workspace_id
WHERE k.workspace_id=$1 AND wm.user_id=$2 AND wm.role IN('owner','admin')
ORDER BY k.created_at DESC;

-- name: RevokeAPIKeyForManager :one
UPDATE api_keys k SET revoked_at=COALESCE(k.revoked_at,now())
FROM workspace_members wm
WHERE k.id=$1 AND wm.workspace_id=k.workspace_id AND wm.user_id=$2 AND wm.role IN('owner','admin')
RETURNING k.id;
