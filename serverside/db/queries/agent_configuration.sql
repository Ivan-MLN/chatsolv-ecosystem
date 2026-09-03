-- name: AuthorizeAgentMember :one
SELECT wm.role,a.workspace_id FROM agents a JOIN workspace_members wm ON wm.workspace_id=a.workspace_id WHERE a.id=$1 AND wm.user_id=$2 AND a.deleted_at IS NULL;

-- name: UpsertPersonality :one
INSERT INTO agent_personalities(id,workspace_id,agent_id,bot_name,role,tone,communication_style,primary_language,response_length,emoji_usage,greeting_style,closing_style,custom_instructions,behavior_rules,escalation_rules,forbidden_topics,fallback_behavior)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
ON CONFLICT(workspace_id,agent_id) DO UPDATE SET bot_name=EXCLUDED.bot_name,role=EXCLUDED.role,tone=EXCLUDED.tone,communication_style=EXCLUDED.communication_style,primary_language=EXCLUDED.primary_language,response_length=EXCLUDED.response_length,emoji_usage=EXCLUDED.emoji_usage,greeting_style=EXCLUDED.greeting_style,closing_style=EXCLUDED.closing_style,custom_instructions=EXCLUDED.custom_instructions,behavior_rules=EXCLUDED.behavior_rules,escalation_rules=EXCLUDED.escalation_rules,forbidden_topics=EXCLUDED.forbidden_topics,fallback_behavior=EXCLUDED.fallback_behavior,updated_at=now()
RETURNING id,workspace_id,agent_id,bot_name,role,tone,communication_style,primary_language,response_length,emoji_usage,greeting_style,closing_style,custom_instructions,behavior_rules,escalation_rules,forbidden_topics,fallback_behavior,created_at,updated_at;

-- name: IncrementAgentConfigVersion :one
UPDATE agents SET config_version=config_version+1,status='syncing',updated_at=now() WHERE id=$1 RETURNING config_version;

-- name: GetPersonality :one
SELECT id,workspace_id,agent_id,bot_name,role,tone,communication_style,primary_language,response_length,emoji_usage,greeting_style,closing_style,custom_instructions,behavior_rules,escalation_rules,forbidden_topics,fallback_behavior,created_at,updated_at FROM agent_personalities WHERE agent_id=$1;

-- name: CreateAgentSyncEvent :exec
INSERT INTO outbox_events(id,workspace_id,event_type,aggregate_type,aggregate_id,payload) VALUES($1,$2,'agent.sync','agent',$3,$4);

-- name: GetAgentSyncResource :one
SELECT a.id,a.workspace_id,a.provider_agent_id,a.config_version,b.vault_key FROM agents a JOIN second_brains b ON b.agent_id=a.id AND b.workspace_id=a.workspace_id WHERE a.id=$1 AND a.deleted_at IS NULL AND b.deleted_at IS NULL;

-- name: CompleteAgentSync :exec
UPDATE agents SET status='ready',synced_config_version=config_version,updated_at=now() WHERE id=$1;
