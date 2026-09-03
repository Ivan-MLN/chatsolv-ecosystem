-- name: CreateAgent :one
INSERT INTO agents (id, workspace_id, name, status, provider)
VALUES ($1,$2,$3,$4,'hermes')
RETURNING id,workspace_id,name,status,provider,provider_agent_id,config_version,synced_config_version,created_at,updated_at,deleted_at;

-- name: CreateSecondBrain :one
INSERT INTO second_brains (id,workspace_id,agent_id,provider,status)
VALUES ($1,$2,$3,'obsidian',$4)
RETURNING id,workspace_id,agent_id,provider,vault_key,schema_version,version,status,last_synced_at,created_at,updated_at,deleted_at;

-- name: CreateOutboxEvent :one
INSERT INTO outbox_events (id,workspace_id,event_type,aggregate_type,aggregate_id,payload)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING id,workspace_id,event_type,aggregate_type,aggregate_id,payload,status,attempts,available_at,processed_at,last_error_code,created_at,updated_at;

-- name: GetProvisioningResource :one
SELECT w.id AS workspace_id,a.id AS agent_id,b.id AS second_brain_id,
       COALESCE(a.provider_agent_id,'') AS provider_agent_id,COALESCE(b.vault_key,'') AS vault_key
FROM workspaces w
JOIN agents a ON a.workspace_id=w.id AND a.deleted_at IS NULL
JOIN second_brains b ON b.workspace_id=w.id AND b.agent_id=a.id AND b.deleted_at IS NULL
WHERE w.id=$1 AND w.deleted_at IS NULL;

-- name: MarkAgentProvisioning :exec
UPDATE agents SET status='provisioning',updated_at=now() WHERE workspace_id=$1 AND deleted_at IS NULL;

-- name: MarkBrainProvisioning :exec
UPDATE second_brains SET status='provisioning',updated_at=now() WHERE workspace_id=$1 AND deleted_at IS NULL;

-- name: CompleteAgentProvisioning :exec
UPDATE agents SET status='ready',provider_agent_id=$2,synced_config_version=config_version,updated_at=now() WHERE workspace_id=$1 AND deleted_at IS NULL;

-- name: CompleteBrainProvisioning :exec
UPDATE second_brains SET status='ready',vault_key=$2,last_synced_at=now(),updated_at=now() WHERE workspace_id=$1 AND deleted_at IS NULL;

-- name: ActivateWorkspace :exec
UPDATE workspaces SET status='active',updated_at=now() WHERE id=$1;

-- name: FailAgentProvisioning :exec
UPDATE agents SET status='failed',updated_at=now() WHERE workspace_id=$1 AND deleted_at IS NULL;

-- name: FailBrainProvisioning :exec
UPDATE second_brains SET status='failed',updated_at=now() WHERE workspace_id=$1 AND deleted_at IS NULL;

-- name: RecordProvisioningError :exec
UPDATE outbox_events SET last_error_code=$2,updated_at=now() WHERE workspace_id=$1 AND event_type='workspace.provision' AND status IN ('pending','processing');

-- name: ClaimOutboxEvent :one
UPDATE outbox_events SET status='processing',attempts=attempts+1,updated_at=now()
WHERE id=(SELECT id FROM outbox_events WHERE status='pending' AND available_at<=now() ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1)
RETURNING id,workspace_id,event_type,aggregate_id,payload,attempts;

-- name: CompleteOutboxEvent :exec
UPDATE outbox_events SET status='completed',processed_at=now(),updated_at=now() WHERE id=$1;

-- name: RetryOutboxEvent :exec
UPDATE outbox_events SET status='pending',available_at=$2,last_error_code=$3,updated_at=now() WHERE id=$1;

-- name: FailOutboxEvent :exec
UPDATE outbox_events SET status='failed',last_error_code=$2,updated_at=now() WHERE id=$1;

-- name: CreateFailedJob :exec
INSERT INTO failed_jobs(id,workspace_id,outbox_event_id,job_type,error_code,attempts) VALUES($1,$2,$3,$4,$5,$6);

-- name: GetAgentForMember :one
SELECT a.id,a.workspace_id,a.name,a.status,a.provider,a.config_version,a.synced_config_version,a.created_at,a.updated_at
FROM agents a JOIN workspace_members wm ON wm.workspace_id=a.workspace_id
WHERE a.id=$1 AND wm.user_id=$2 AND a.deleted_at IS NULL;

-- name: GetDefaultAgentForWorkspaceMember :one
SELECT a.id,a.workspace_id,a.name,a.status,a.provider,a.config_version,a.synced_config_version,wm.role
FROM agents a JOIN workspace_members wm ON wm.workspace_id=a.workspace_id
WHERE a.workspace_id=$1 AND wm.user_id=$2 AND a.deleted_at IS NULL
ORDER BY a.created_at LIMIT 1;

-- name: UpdateAgentName :one
UPDATE agents SET name=$2,updated_at=now() WHERE id=$1 AND deleted_at IS NULL
RETURNING id,workspace_id,name,status,provider,config_version,synced_config_version;

-- name: EnsureAgentTestChannel :one
INSERT INTO channels(id,workspace_id,agent_id,type,display_name,status)
VALUES($1,$2,$3,'web','Agent Test Playground','connected')
ON CONFLICT(id) DO UPDATE SET updated_at=now()
RETURNING id;

-- name: CreateAgentTestConversation :one
INSERT INTO conversations(id,workspace_id,agent_id,channel_id,external_user_id,status,mode,environment)
VALUES($1,$2,$3,$4,$5,'open','agent','test')
RETURNING id;

-- name: CreateAgentTestMessage :exec
INSERT INTO messages(id,workspace_id,conversation_id,channel_id,sender_type,content_type,content,provider,status)
VALUES($1,$2,$3,$4,$5,'text',$6,$7,$8);
