-- name: CreateWorkspace :one
INSERT INTO workspaces (id, name, slug, owner_user_id, status, timezone)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, name, slug, owner_user_id, status, timezone, created_at, updated_at, deleted_at;

-- name: CreateWorkspaceMember :one
INSERT INTO workspace_members (id, workspace_id, user_id, role)
VALUES ($1, $2, $3, $4)
RETURNING id, workspace_id, user_id, role, created_at, updated_at;

-- name: CreateSubscription :one
INSERT INTO subscriptions (
    id, workspace_id, status, plan_id, billing_cycle, currency, amount,
    current_period_start, current_period_end, payment_reference, cancel_at_period_end
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, workspace_id, status, plan_id, billing_cycle, currency, amount,
          current_period_start, current_period_end, payment_reference, cancel_at_period_end,
          trial_started_at, trial_ends_at, created_at, updated_at;

-- name: CreateSubscriptionEntitlement :one
INSERT INTO subscription_entitlements (
    id, subscription_id, workspace_id, max_agents, max_channels,
    max_storage_mb, max_documents, monthly_messages, public_api, webhooks
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, subscription_id, workspace_id, max_agents, max_channels,
          max_storage_mb, max_documents, monthly_messages, public_api, webhooks,
          created_at, updated_at;

-- name: GetWorkspaceForMember :one
SELECT w.id, w.name, w.slug, w.owner_user_id, w.status, w.timezone,
       w.created_at, w.updated_at, w.deleted_at,
       wm.id AS membership_id, wm.user_id, wm.role,
       wm.created_at AS membership_created_at, wm.updated_at AS membership_updated_at
FROM workspaces w
JOIN workspace_members wm ON wm.workspace_id = w.id
WHERE w.id = $1 AND wm.user_id = $2 AND w.deleted_at IS NULL;

-- name: UpdateWorkspace :one
UPDATE workspaces
SET name = $2, timezone = $3, updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, name, slug, owner_user_id, status, timezone, created_at, updated_at, deleted_at;

-- name: GetWorkspaceSubscription :one
SELECT s.id, s.workspace_id, s.status, s.plan_id, s.billing_cycle, s.currency, s.amount,
       s.current_period_start, s.current_period_end, s.payment_reference, s.cancel_at_period_end,
       s.trial_started_at, s.trial_ends_at, s.created_at, s.updated_at,
       e.max_agents, e.max_channels, e.max_storage_mb, e.max_documents,
       e.monthly_messages, e.public_api, e.webhooks
FROM subscriptions s
JOIN subscription_entitlements e ON e.subscription_id = s.id
WHERE s.workspace_id = $1;

-- name: UpdateSubscriptionStatus :one
UPDATE subscriptions
SET status = $2,
    plan_id = COALESCE(NULLIF($3::varchar, ''), plan_id),
    current_period_start = COALESCE($4::timestamptz, current_period_start),
    current_period_end = COALESCE($5::timestamptz, current_period_end),
    payment_reference = COALESCE(NULLIF($6::text, ''), payment_reference),
    updated_at = now()
WHERE workspace_id = $1
RETURNING id, workspace_id, status, plan_id, billing_cycle, currency, amount,
          current_period_start, current_period_end, payment_reference, cancel_at_period_end,
          trial_started_at, trial_ends_at, created_at, updated_at;

-- name: CreatePayment :one
INSERT INTO payments (
    id, workspace_id, subscription_id, provider, provider_transaction_id,
    payment_reference, amount, currency, status, metadata
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, workspace_id, subscription_id, provider, provider_transaction_id,
          payment_reference, amount, currency, status, metadata, created_at, updated_at;

-- name: GetPaymentByReference :one
SELECT id, workspace_id, subscription_id, provider, provider_transaction_id,
       payment_reference, amount, currency, status, metadata, created_at, updated_at
FROM payments
WHERE payment_reference = $1;

-- name: GetPaymentByProviderTxID :one
SELECT id, workspace_id, subscription_id, provider, provider_transaction_id,
       payment_reference, amount, currency, status, metadata, created_at, updated_at
FROM payments
WHERE provider_transaction_id = $1;

-- name: UpdatePaymentStatus :one
UPDATE payments
SET status = $2,
    provider_transaction_id = COALESCE(NULLIF($3::text, ''), provider_transaction_id),
    metadata = COALESCE(NULLIF($4::jsonb, '{}'::jsonb), metadata),
    updated_at = now()
WHERE id = $1
RETURNING id, workspace_id, subscription_id, provider, provider_transaction_id,
          payment_reference, amount, currency, status, metadata, created_at, updated_at;

-- name: CountWorkspaceAgents :one
SELECT count(*) FROM agents WHERE workspace_id = $1 AND deleted_at IS NULL;

-- name: GetKnowledgeStorageBytes :one
SELECT COALESCE(sum(size), 0)::bigint FROM object_metadata WHERE workspace_id = $1;
