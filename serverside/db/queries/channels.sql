-- name: CreateChannel :one
INSERT INTO channels(id,workspace_id,agent_id,type,display_name,status) VALUES($1,$2,$3,$4,$5,$6)
RETURNING id,workspace_id,agent_id,type,display_name,phone_number,external_channel_id,service_instance_id,status,connected_at,last_seen_at,created_at,updated_at;

-- name: AuthorizeChannelWorkspace :one
SELECT wm.role,a.id AS agent_id FROM workspace_members wm JOIN agents a ON a.workspace_id=wm.workspace_id AND a.deleted_at IS NULL WHERE wm.workspace_id=$1 AND wm.user_id=$2 ORDER BY a.created_at LIMIT 1;

-- name: ListChannelsForMember :many
SELECT c.id,c.workspace_id,c.agent_id,c.type,c.display_name,c.phone_number,c.external_channel_id,c.service_instance_id,c.status,c.connected_at,c.last_seen_at,c.created_at,c.updated_at FROM channels c JOIN workspace_members wm ON wm.workspace_id=c.workspace_id WHERE c.workspace_id=$1 AND wm.user_id=$2 ORDER BY c.created_at DESC;

-- name: UpdateChannelStatusInternal :exec
UPDATE channels
SET status=sqlc.arg(status)::varchar(30),
    phone_number=COALESCE(sqlc.narg(phone_number)::text,phone_number),
    service_instance_id=COALESCE(sqlc.narg(service_instance_id)::text,service_instance_id),
    connected_at=CASE WHEN sqlc.arg(status)::varchar(30)='connected' THEN COALESCE(connected_at,now()) ELSE connected_at END,
    last_seen_at=now(),
    updated_at=now()
WHERE id=sqlc.arg(id);

-- name: DeleteChannelForMember :exec
DELETE FROM channels c USING workspace_members wm WHERE c.id=$1 AND wm.workspace_id=c.workspace_id AND wm.user_id=$2 AND wm.role IN('owner','admin');

-- name: AuthorizeChannelMutation :one
SELECT wm.role FROM channels c JOIN workspace_members wm ON wm.workspace_id=c.workspace_id
WHERE c.id=$1 AND wm.user_id=$2 AND wm.role IN('owner','admin');

-- name: GetChannelByID :one
SELECT id,workspace_id,agent_id,type,display_name,phone_number,external_channel_id,service_instance_id,status,connected_at,last_seen_at,created_at,updated_at FROM channels WHERE id=$1;

-- name: CountWorkspaceChannels :one
SELECT count(*) FROM channels WHERE workspace_id=$1 AND type='whatsapp' AND status NOT IN ('disconnected', 'deleted');

-- name: GetMaxChannels :one
SELECT max_channels FROM subscription_entitlements WHERE workspace_id=$1;

-- name: AuditChannelEvent :exec
INSERT INTO audit_logs(id,workspace_id,event,resource_type,resource_id,metadata) SELECT $1,c.workspace_id,$2,'channel',c.id,$3 FROM channels c WHERE c.id=$4;

-- name: ListRecentAuditLogsForWorkspace :many
SELECT id,workspace_id,user_id,event,resource_type,resource_id,metadata,created_at
FROM audit_logs
WHERE workspace_id=$1
ORDER BY created_at DESC
LIMIT $2;
