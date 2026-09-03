CREATE TABLE workspaces (
    id uuid PRIMARY KEY,
    name varchar(120) NOT NULL,
    slug varchar(63) NOT NULL UNIQUE,
    owner_user_id uuid NOT NULL REFERENCES users(id),
    status varchar(20) NOT NULL CHECK (status IN ('provisioning', 'active', 'suspended', 'deleting', 'deleted')),
    timezone varchar(64) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE workspace_members (
    id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role varchar(20) NOT NULL CHECK (role IN ('owner', 'admin', 'member', 'viewer')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, user_id)
);

CREATE INDEX workspace_members_user_workspace_idx ON workspace_members(user_id, workspace_id);

CREATE TABLE subscriptions (
    id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL UNIQUE REFERENCES workspaces(id) ON DELETE CASCADE,
    status varchar(20) NOT NULL CHECK (status IN ('trialing', 'active', 'past_due', 'suspended', 'cancelled', 'expired')),
    trial_started_at timestamptz NOT NULL,
    trial_ends_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (trial_ends_at > trial_started_at)
);

CREATE TABLE subscription_entitlements (
    id uuid PRIMARY KEY,
    subscription_id uuid NOT NULL UNIQUE REFERENCES subscriptions(id) ON DELETE CASCADE,
    workspace_id uuid NOT NULL UNIQUE REFERENCES workspaces(id) ON DELETE CASCADE,
    max_agents integer NOT NULL CHECK (max_agents > 0),
    max_channels integer NOT NULL CHECK (max_channels > 0),
    max_storage_mb bigint NOT NULL CHECK (max_storage_mb > 0),
    max_documents integer NOT NULL CHECK (max_documents > 0),
    monthly_messages bigint NOT NULL CHECK (monthly_messages > 0),
    public_api boolean NOT NULL DEFAULT false,
    webhooks boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX workspaces_owner_id_idx ON workspaces(owner_user_id, id);
CREATE INDEX subscriptions_workspace_id_idx ON subscriptions(workspace_id, id);
