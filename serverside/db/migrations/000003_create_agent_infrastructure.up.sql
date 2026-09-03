CREATE TABLE agents (
    id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name varchar(120) NOT NULL,
    status varchar(20) NOT NULL CHECK (status IN ('pending','provisioning','ready','syncing','suspended','failed','deleting','deleted')),
    provider varchar(20) NOT NULL CHECK (provider = 'hermes'),
    provider_agent_id text,
    config_version bigint NOT NULL DEFAULT 1 CHECK (config_version > 0),
    synced_config_version bigint NOT NULL DEFAULT 0 CHECK (synced_config_version >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    UNIQUE (workspace_id, id)
);
CREATE UNIQUE INDEX agents_one_live_default_idx ON agents(workspace_id) WHERE deleted_at IS NULL;

CREATE TABLE second_brains (
    id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    agent_id uuid NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    provider varchar(20) NOT NULL CHECK (provider = 'obsidian'),
    vault_key text,
    schema_version integer NOT NULL DEFAULT 1 CHECK (schema_version > 0),
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    status varchar(20) NOT NULL CHECK (status IN ('pending','provisioning','ready','syncing','failed','suspended','deleting','deleted')),
    last_synced_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    UNIQUE (workspace_id, id),
    UNIQUE (workspace_id, agent_id)
);

CREATE TABLE outbox_events (
    id uuid PRIMARY KEY,
    workspace_id uuid REFERENCES workspaces(id) ON DELETE CASCADE,
    event_type varchar(80) NOT NULL,
    aggregate_type varchar(40) NOT NULL,
    aggregate_id uuid NOT NULL,
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    status varchar(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','processing','completed','failed')),
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    available_at timestamptz NOT NULL DEFAULT now(),
    processed_at timestamptz,
    last_error_code varchar(80),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX outbox_events_dispatch_idx ON outbox_events(status, available_at, created_at);

CREATE TABLE failed_jobs (
    id uuid PRIMARY KEY,
    workspace_id uuid REFERENCES workspaces(id) ON DELETE SET NULL,
    outbox_event_id uuid REFERENCES outbox_events(id) ON DELETE SET NULL,
    job_type varchar(80) NOT NULL,
    error_code varchar(80) NOT NULL,
    attempts integer NOT NULL,
    failed_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX failed_jobs_workspace_idx ON failed_jobs(workspace_id, failed_at DESC);
CREATE INDEX agents_workspace_idx ON agents(workspace_id, id);
CREATE INDEX second_brains_workspace_idx ON second_brains(workspace_id, id);
ALTER TABLE workspaces ADD CONSTRAINT workspaces_owner_membership_deferred_note CHECK (owner_user_id IS NOT NULL);
