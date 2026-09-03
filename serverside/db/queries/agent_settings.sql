-- name: AuthorizeWorkspaceMember :one
SELECT role FROM workspace_members WHERE workspace_id=$1 AND user_id=$2;

-- name: UpsertAgentProfile :one
INSERT INTO agent_profiles(id,workspace_id,agent_id,display_name,avatar_object_key,description,greeting_message,away_message,fallback_message,language)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
ON CONFLICT(workspace_id,agent_id) DO UPDATE SET display_name=EXCLUDED.display_name,avatar_object_key=EXCLUDED.avatar_object_key,description=EXCLUDED.description,greeting_message=EXCLUDED.greeting_message,away_message=EXCLUDED.away_message,fallback_message=EXCLUDED.fallback_message,language=EXCLUDED.language,updated_at=now()
RETURNING id,workspace_id,agent_id,display_name,avatar_object_key,description,greeting_message,away_message,fallback_message,language,created_at,updated_at;

-- name: GetAgentProfile :one
SELECT id,workspace_id,agent_id,display_name,avatar_object_key,description,greeting_message,away_message,fallback_message,language,created_at,updated_at FROM agent_profiles WHERE agent_id=$1;

-- name: UpsertBusinessProfile :one
INSERT INTO business_profiles(id,workspace_id,business_name,industry,business_description,website,email,phone,address,business_hours,timezone,brand_voice,company_values,business_type,target_customer,products_services,communication_style,primary_use_cases,handoff_rules,operating_hours)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
ON CONFLICT(workspace_id) DO UPDATE SET
    business_name=EXCLUDED.business_name,
    industry=EXCLUDED.industry,
    business_description=EXCLUDED.business_description,
    website=EXCLUDED.website,
    email=EXCLUDED.email,
    phone=EXCLUDED.phone,
    address=EXCLUDED.address,
    business_hours=EXCLUDED.business_hours,
    timezone=EXCLUDED.timezone,
    brand_voice=EXCLUDED.brand_voice,
    company_values=EXCLUDED.company_values,
    business_type=EXCLUDED.business_type,
    target_customer=EXCLUDED.target_customer,
    products_services=EXCLUDED.products_services,
    communication_style=EXCLUDED.communication_style,
    primary_use_cases=EXCLUDED.primary_use_cases,
    handoff_rules=EXCLUDED.handoff_rules,
    operating_hours=EXCLUDED.operating_hours,
    updated_at=now()
RETURNING id,workspace_id,business_name,industry,business_description,website,email,phone,address,business_hours,timezone,brand_voice,company_values,business_type,target_customer,products_services,communication_style,primary_use_cases,handoff_rules,operating_hours,created_at,updated_at;

-- name: GetBusinessProfile :one
SELECT id,workspace_id,business_name,industry,business_description,website,email,phone,address,business_hours,timezone,brand_voice,company_values,business_type,target_customer,products_services,communication_style,primary_use_cases,handoff_rules,operating_hours,created_at,updated_at FROM business_profiles WHERE workspace_id=$1;

-- name: UpsertBusinessPolicies :one
INSERT INTO business_policies(id,workspace_id,shipping_policy,refund_policy,return_policy,warranty_policy,payment_policy,complaint_policy)
VALUES($1,$2,$3,$4,$5,$6,$7,$8)
ON CONFLICT(workspace_id) DO UPDATE SET shipping_policy=EXCLUDED.shipping_policy,refund_policy=EXCLUDED.refund_policy,return_policy=EXCLUDED.return_policy,warranty_policy=EXCLUDED.warranty_policy,payment_policy=EXCLUDED.payment_policy,complaint_policy=EXCLUDED.complaint_policy,updated_at=now()
RETURNING id,workspace_id,shipping_policy,refund_policy,return_policy,warranty_policy,payment_policy,complaint_policy,created_at,updated_at;

-- name: GetBusinessPolicies :one
SELECT id,workspace_id,shipping_policy,refund_policy,return_policy,warranty_policy,payment_policy,complaint_policy,created_at,updated_at FROM business_policies WHERE workspace_id=$1;

-- name: GetDefaultAgentForWorkspace :one
SELECT id FROM agents WHERE workspace_id=$1 AND deleted_at IS NULL ORDER BY created_at LIMIT 1;

-- name: GetSecondBrainForWorkspace :one
SELECT vault_key FROM second_brains WHERE workspace_id=$1 AND deleted_at IS NULL ORDER BY created_at LIMIT 1;

-- name: QueueAgentSyncForWorkspace :exec
INSERT INTO outbox_events(id,workspace_id,event_type,aggregate_type,aggregate_id,payload)
SELECT $1,$2,'agent.sync','agent',a.id,$3 FROM agents a WHERE a.workspace_id=$2 AND a.deleted_at IS NULL ORDER BY a.created_at LIMIT 1;

-- name: IncrementWorkspaceAgentConfig :one
UPDATE agents target SET config_version=target.config_version+1,status='syncing',updated_at=now() WHERE target.id=(SELECT candidate.id FROM agents candidate WHERE candidate.workspace_id=$1 AND candidate.deleted_at IS NULL ORDER BY candidate.created_at LIMIT 1) RETURNING target.config_version;
