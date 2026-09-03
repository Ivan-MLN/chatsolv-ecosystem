-- name: AuthorizeWebhookWorkspace :one
SELECT wm.role,e.webhooks FROM workspace_members wm JOIN subscription_entitlements e ON e.workspace_id=wm.workspace_id
WHERE wm.workspace_id=$1 AND wm.user_id=$2;

-- name: CreateWebhookEndpoint :one
INSERT INTO webhook_endpoints(id,workspace_id,url,events,status,secret_ciphertext)
VALUES($1,$2,$3,$4,$5,$6)
RETURNING id,workspace_id,url,events,status,secret_ciphertext,created_at,updated_at,deleted_at;

-- name: ListWebhookEndpointsForManager :many
SELECT w.id,w.workspace_id,w.url,w.events,w.status,w.secret_ciphertext,w.created_at,w.updated_at,w.deleted_at
FROM webhook_endpoints w JOIN workspace_members wm ON wm.workspace_id=w.workspace_id
WHERE w.workspace_id=$1 AND wm.user_id=$2 AND wm.role IN('owner','admin') AND w.deleted_at IS NULL
ORDER BY w.created_at DESC;

-- name: UpdateWebhookEndpointForManager :one
UPDATE webhook_endpoints w SET url=$3,events=$4,status=$5,updated_at=now()
FROM workspace_members wm WHERE w.id=$1 AND wm.workspace_id=w.workspace_id AND wm.user_id=$2 AND wm.role IN('owner','admin') AND w.deleted_at IS NULL
RETURNING w.id,w.workspace_id,w.url,w.events,w.status,w.secret_ciphertext,w.created_at,w.updated_at,w.deleted_at;

-- name: DeleteWebhookEndpointForManager :one
UPDATE webhook_endpoints w SET deleted_at=now(),status='disabled',updated_at=now()
FROM workspace_members wm WHERE w.id=$1 AND wm.workspace_id=w.workspace_id AND wm.user_id=$2 AND wm.role IN('owner','admin') AND w.deleted_at IS NULL
RETURNING w.id;
