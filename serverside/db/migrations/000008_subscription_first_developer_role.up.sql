-- Step 1: Add platform_role to users table with default 'user'
ALTER TABLE users ADD COLUMN IF NOT EXISTS platform_role varchar(20) NOT NULL DEFAULT 'user';

-- Step 2: One-time migration to mark ALL existing users as 'developer'
UPDATE users SET platform_role = 'developer' WHERE platform_role = 'user';

-- Step 3: Add subscription fields for plan, billing cycle, currency, price, current period, and payment reference
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS plan_id varchar(50) NOT NULL DEFAULT 'chatsolv_starter';
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS billing_cycle varchar(20) NOT NULL DEFAULT 'monthly';
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS currency varchar(10) NOT NULL DEFAULT 'IDR';
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS amount bigint NOT NULL DEFAULT 459000;
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS current_period_start timestamptz NOT NULL DEFAULT now();
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS current_period_end timestamptz NOT NULL DEFAULT (now() + interval '30 days');
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS payment_reference text;
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS cancel_at_period_end boolean NOT NULL DEFAULT false;

-- Step 4: Drop old trial constraints if any, adjust subscription status check
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_trial_ends_at_check;
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_status_check;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_status_check CHECK (status IN ('inactive', 'pending_payment', 'active', 'past_due', 'suspended', 'cancelled', 'expired', 'trialing'));

-- Make trial dates nullable for new subscription-first model
ALTER TABLE subscriptions ALTER COLUMN trial_started_at DROP NOT NULL;
ALTER TABLE subscriptions ALTER COLUMN trial_ends_at DROP NOT NULL;

-- Step 5: Payments table for server-side verification and idempotent webhook processing
CREATE TABLE IF NOT EXISTS payments (
    id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    subscription_id uuid NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    provider varchar(50) NOT NULL DEFAULT 'manual',
    provider_transaction_id text UNIQUE,
    payment_reference text NOT NULL UNIQUE,
    amount bigint NOT NULL,
    currency varchar(10) NOT NULL DEFAULT 'IDR',
    status varchar(20) NOT NULL CHECK (status IN ('pending', 'paid', 'failed', 'expired', 'refunded')),
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS payments_workspace_idx ON payments(workspace_id, created_at DESC);
CREATE INDEX IF NOT EXISTS payments_reference_idx ON payments(payment_reference);
CREATE INDEX IF NOT EXISTS payments_provider_tx_idx ON payments(provider_transaction_id);
