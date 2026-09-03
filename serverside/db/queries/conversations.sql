-- name: FindIncomingResult :one
SELECT response_message.id,response_message.conversation_id,response_message.content
FROM messages incoming JOIN messages response_message ON response_message.conversation_id=incoming.conversation_id AND response_message.sender_type='agent' AND response_message.created_at>=incoming.created_at
WHERE incoming.channel_id=$1 AND incoming.external_message_id=$2 ORDER BY response_message.created_at LIMIT 1;

-- name: ResolveChannelRuntime :one
SELECT c.id AS channel_id,c.workspace_id,c.agent_id,c.status AS channel_status,a.status AS agent_status,a.provider_agent_id,s.status AS subscription_status,s.trial_ends_at,sb.vault_key
FROM channels c JOIN agents a ON a.id=c.agent_id AND a.workspace_id=c.workspace_id JOIN subscriptions s ON s.workspace_id=c.workspace_id JOIN second_brains sb ON sb.agent_id=a.id
WHERE c.id=$1 AND c.type=$2;

-- name: FindOpenConversation :one
SELECT * FROM conversations WHERE channel_id=$1 AND external_user_id=$2 AND status='open';

-- name: GetConversationByID :one
SELECT * FROM conversations WHERE id=$1;

-- name: GetConversationForMember :one
SELECT c.*
FROM conversations c JOIN workspace_members wm ON wm.workspace_id=c.workspace_id
WHERE c.id=$1 AND wm.user_id=$2;

-- name: CreateConversation :one
INSERT INTO conversations(id,workspace_id,agent_id,channel_id,external_user_id,status,mode,environment,metadata)
VALUES($1,$2,$3,$4,$5,'open','agent',$6,$7)
RETURNING *;

-- name: CreateMessage :one
INSERT INTO messages(id,workspace_id,conversation_id,channel_id,sender_type,content_type,content,external_message_id,provider,status)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
RETURNING id,workspace_id,conversation_id,channel_id,sender_type,content_type,content,external_message_id,provider,status,created_at;

-- name: ListRecentMessages :many
SELECT sender_type,content FROM messages WHERE workspace_id=$1 AND conversation_id=$2 ORDER BY created_at DESC LIMIT $3;

-- name: UpdateConversationActivity :exec
UPDATE conversations SET last_message_at=now(),updated_at=now() WHERE id=$1 AND workspace_id=$2;

-- name: SetConversationMode :exec
UPDATE conversations SET mode=$3,assigned_user_id=$4,updated_at=now() WHERE id=$1 AND workspace_id=$2;

-- name: GetConversationManagerRole :one
SELECT wm.role FROM conversations c JOIN workspace_members wm ON wm.workspace_id=c.workspace_id
WHERE c.id=$1 AND wm.user_id=$2;

-- name: SetConversationModeForManager :one
UPDATE conversations c SET mode=sqlc.arg(mode)::varchar(10),assigned_user_id=CASE WHEN sqlc.arg(mode)::varchar(10)='human' THEN sqlc.arg(user_id)::uuid ELSE NULL END,updated_at=now()
FROM workspace_members wm
WHERE c.id=sqlc.arg(id)::uuid AND wm.workspace_id=c.workspace_id AND wm.user_id=sqlc.arg(user_id)::uuid AND wm.role IN('owner','admin')
RETURNING c.id;

-- name: ListConversationsForMember :many
SELECT c.* FROM conversations c JOIN workspace_members wm ON wm.workspace_id=c.workspace_id WHERE c.workspace_id=$1 AND wm.user_id=$2 AND c.hidden_at IS NULL AND c.environment='production' AND ($3::text='' OR c.status=$3) AND ($4::text='' OR c.mode=$4) ORDER BY c.last_message_at DESC LIMIT $5;

-- name: ListMessagesForMember :many
SELECT m.id,m.workspace_id,m.conversation_id,m.channel_id,m.sender_type,m.content_type,m.content,m.external_message_id,m.provider,m.status,m.created_at FROM messages m JOIN conversations c ON c.id=m.conversation_id JOIN workspace_members wm ON wm.workspace_id=c.workspace_id WHERE m.conversation_id=$1 AND wm.user_id=$2 AND ($3::timestamptz IS NULL OR m.created_at<$3) ORDER BY m.created_at DESC LIMIT $4;

-- name: GetRuntimePersonality :one
SELECT bot_name,role,tone,communication_style,primary_language,response_length,emoji_usage,custom_instructions,behavior_rules,escalation_rules,forbidden_topics,fallback_behavior FROM agent_personalities WHERE agent_id=$1;

-- name: GetRuntimeBusiness :one
SELECT business_name,industry,business_description,business_hours,timezone,brand_voice FROM business_profiles WHERE workspace_id=$1;

-- name: RecordUsage :exec
INSERT INTO usage_records(id,workspace_id,metric,quantity,environment) VALUES($1,$2,$3,$4,$5);

-- name: ReserveMonthlyMessage :one
WITH quota_lock AS MATERIALIZED (
    SELECT pg_advisory_xact_lock(hashtextextended(sqlc.arg(workspace_id)::text, 0))
), policy AS MATERIALIZED (
    SELECT u.platform_role, e.monthly_messages
    FROM workspaces w
    JOIN users u ON u.id = w.owner_user_id
    JOIN subscription_entitlements e ON e.workspace_id = w.id
    CROSS JOIN quota_lock
    WHERE w.id = sqlc.arg(workspace_id)::uuid
), usage AS MATERIALIZED (
    SELECT COALESCE(sum(ur.quantity), 0)::bigint AS quantity
    FROM quota_lock
    LEFT JOIN usage_records ur
      ON ur.workspace_id = sqlc.arg(workspace_id)::uuid
     AND ur.metric = 'agent_invocations'
     AND ur.recorded_at >= date_trunc('month', now())
)
INSERT INTO usage_records(id, workspace_id, metric, quantity, environment)
SELECT sqlc.arg(id)::uuid, sqlc.arg(workspace_id)::uuid, 'agent_invocations', 1, sqlc.arg(environment)::varchar
FROM policy, usage
WHERE policy.platform_role = 'developer' OR usage.quantity < policy.monthly_messages
RETURNING id;

-- name: CountMonthlyUsage :one
SELECT COALESCE(sum(quantity),0)::bigint FROM usage_records WHERE workspace_id=$1 AND metric=$2 AND recorded_at>=date_trunc('month',now());

-- name: GetMonthlyMessageEntitlement :one
SELECT monthly_messages FROM subscription_entitlements WHERE workspace_id=$1;

-- name: FindDefaultReadyAgent :one
SELECT a.id,a.workspace_id,a.provider_agent_id,sb.vault_key FROM agents a JOIN second_brains sb ON sb.agent_id=a.id AND sb.workspace_id=a.workspace_id JOIN subscriptions s ON s.workspace_id=a.workspace_id WHERE a.workspace_id=$1 AND a.status='ready' AND sb.status='ready' AND s.status IN('trialing','active') AND a.deleted_at IS NULL ORDER BY a.created_at LIMIT 1;

-- name: FindOrCreateWebChannel :one
INSERT INTO channels(id,workspace_id,agent_id,type,display_name,status) VALUES($1,$2,$3,'web','Website','connected')
ON CONFLICT(workspace_id,id) DO UPDATE SET updated_at=now()
RETURNING id;
