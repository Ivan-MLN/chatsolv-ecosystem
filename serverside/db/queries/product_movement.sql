-- name: GetOnboardingProfile :one
SELECT id, workspace_id, user_id, current_step, is_completed, data, completed_at, created_at, updated_at
FROM onboarding_profiles
WHERE workspace_id = $1;

-- name: UpsertOnboardingProfile :one
INSERT INTO onboarding_profiles (
    id, workspace_id, user_id, current_step, is_completed, data, completed_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, now(), now()
)
ON CONFLICT (workspace_id) DO UPDATE SET
    current_step = EXCLUDED.current_step,
    is_completed = EXCLUDED.is_completed,
    data = EXCLUDED.data,
    completed_at = COALESCE(EXCLUDED.completed_at, onboarding_profiles.completed_at),
    updated_at = now()
RETURNING id, workspace_id, user_id, current_step, is_completed, data, completed_at, created_at, updated_at;

-- name: CompleteOnboardingProfile :one
UPDATE onboarding_profiles
SET is_completed = true, completed_at = now(), updated_at = now()
WHERE workspace_id = $1
RETURNING id, workspace_id, user_id, current_step, is_completed, data, completed_at, created_at, updated_at;

-- name: ListWorkspaceAdmins :many
SELECT id, workspace_id, user_id, name, phone, role, status, is_active, rotation_priority, last_assigned_at, total_handled_today, last_active_date, created_at, updated_at
FROM workspace_admins
WHERE workspace_id = $1
ORDER BY is_active DESC, name ASC;

-- name: GetWorkspaceAdminByID :one
SELECT id, workspace_id, user_id, name, phone, role, status, is_active, rotation_priority, last_assigned_at, total_handled_today, last_active_date, created_at, updated_at
FROM workspace_admins
WHERE workspace_id = $1 AND id = $2;

-- name: FindWorkspaceAdminByPhone :one
SELECT id, workspace_id, user_id, name, phone, role, status, is_active, rotation_priority, last_assigned_at, total_handled_today, last_active_date, created_at, updated_at
FROM workspace_admins
WHERE phone = $1 AND is_active = true
LIMIT 1;

-- name: CreateWorkspaceAdmin :one
INSERT INTO workspace_admins (
    id, workspace_id, user_id, name, phone, role, status, is_active, rotation_priority, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now()
)
RETURNING id, workspace_id, user_id, name, phone, role, status, is_active, rotation_priority, last_assigned_at, total_handled_today, last_active_date, created_at, updated_at;

-- name: UpdateWorkspaceAdmin :one
UPDATE workspace_admins
SET name = $3, phone = $4, role = $5, status = $6, is_active = $7, rotation_priority = $8, updated_at = now()
WHERE workspace_id = $1 AND id = $2
RETURNING id, workspace_id, user_id, name, phone, role, status, is_active, rotation_priority, last_assigned_at, total_handled_today, last_active_date, created_at, updated_at;

-- name: DeleteWorkspaceAdmin :exec
DELETE FROM workspace_admins
WHERE workspace_id = $1 AND id = $2;

-- name: GetNextEligibleAdminForRotation :one
SELECT id, workspace_id, user_id, name, phone, role, status, is_active, rotation_priority, last_assigned_at, total_handled_today, last_active_date, created_at, updated_at
FROM workspace_admins
WHERE workspace_id = $1 AND is_active = true AND status = 'online'
ORDER BY COALESCE(last_assigned_at, '1970-01-01'::timestamptz) ASC, rotation_priority DESC
LIMIT 1;

-- name: RecordAdminAssignment :exec
UPDATE workspace_admins
SET last_assigned_at = now(),
    total_handled_today = CASE WHEN last_active_date = CURRENT_DATE THEN total_handled_today + 1 ELSE 1 END,
    last_active_date = CURRENT_DATE,
    updated_at = now()
WHERE id = $1;

-- name: CreateHandoffRequest :one
INSERT INTO handoff_requests (
    id, short_code, workspace_id, conversation_id, customer_phone, reason, status, assigned_admin_id, requested_at, assigned_at, timeout_at, metadata, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, now(), $9, $10, $11, now(), now()
)
RETURNING id, short_code, workspace_id, conversation_id, customer_phone, reason, status, assigned_admin_id, requested_at, assigned_at, accepted_at, resolved_at, timeout_at, metadata, created_at, updated_at;

-- name: GetHandoffByShortCode :one
SELECT id, short_code, workspace_id, conversation_id, customer_phone, reason, status, assigned_admin_id, requested_at, assigned_at, accepted_at, resolved_at, timeout_at, metadata, created_at, updated_at
FROM handoff_requests
WHERE short_code = $1;

-- name: GetHandoffByID :one
SELECT id, short_code, workspace_id, conversation_id, customer_phone, reason, status, assigned_admin_id, requested_at, assigned_at, accepted_at, resolved_at, timeout_at, metadata, created_at, updated_at
FROM handoff_requests
WHERE workspace_id = $1 AND id = $2;

-- name: ListHandoffRequests :many
SELECT id, short_code, workspace_id, conversation_id, customer_phone, reason, status, assigned_admin_id, requested_at, assigned_at, accepted_at, resolved_at, timeout_at, metadata, created_at, updated_at
FROM handoff_requests
WHERE workspace_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: AcceptHandoffRequestAtomic :one
UPDATE handoff_requests
SET status = 'accepted', assigned_admin_id = $2, accepted_at = now(), updated_at = now()
WHERE id = $1 AND status IN ('pending', 'assigned')
RETURNING id, short_code, workspace_id, conversation_id, customer_phone, reason, status, assigned_admin_id, requested_at, assigned_at, accepted_at, resolved_at, timeout_at, metadata, created_at, updated_at;

-- name: ResolveHandoffRequest :one
UPDATE handoff_requests
SET status = 'resolved', resolved_at = now(), updated_at = now()
WHERE id = $1
RETURNING id, short_code, workspace_id, conversation_id, customer_phone, reason, status, assigned_admin_id, requested_at, assigned_at, accepted_at, resolved_at, timeout_at, metadata, created_at, updated_at;

-- name: ReassignHandoffRequest :one
UPDATE handoff_requests
SET assigned_admin_id = $2, assigned_at = now(), timeout_at = $3, status = 'assigned', updated_at = now()
WHERE id = $1 AND status IN ('pending', 'assigned')
RETURNING id, short_code, workspace_id, conversation_id, customer_phone, reason, status, assigned_admin_id, requested_at, assigned_at, accepted_at, resolved_at, timeout_at, metadata, created_at, updated_at;

-- name: RecordConversationEvent :exec
INSERT INTO conversation_events (
    id, workspace_id, conversation_id, event_type, actor_type, actor_id, payload, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, now()
);

-- name: ListConversationEvents :many
SELECT id, workspace_id, conversation_id, event_type, actor_type, actor_id, payload, created_at
FROM conversation_events
WHERE workspace_id = $1 AND conversation_id = $2
ORDER BY created_at ASC;

-- name: ListAgentTemplates :many
SELECT id, industry, title, description, icon, category, default_profile, default_personality, default_use_cases, default_handoff_rules, is_featured, created_at
FROM agent_templates
ORDER BY is_featured DESC, title ASC;

-- name: GetAgentTemplateByID :one
SELECT id, industry, title, description, icon, category, default_profile, default_personality, default_use_cases, default_handoff_rules, is_featured, created_at
FROM agent_templates
WHERE id = $1;
