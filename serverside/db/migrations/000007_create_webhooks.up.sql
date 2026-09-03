CREATE TABLE webhook_endpoints (
    id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    url text NOT NULL,
    events jsonb NOT NULL,
    status varchar(16) NOT NULL CHECK (status IN ('active','disabled')),
    secret_ciphertext bytea NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);
CREATE INDEX webhook_endpoints_workspace_idx ON webhook_endpoints(workspace_id,created_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE webhook_deliveries (
    id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    webhook_endpoint_id uuid NOT NULL REFERENCES webhook_endpoints(id) ON DELETE CASCADE,
    event_id uuid NOT NULL,
    event varchar(80) NOT NULL,
    http_status integer,
    attempt integer NOT NULL DEFAULT 0 CHECK(attempt >= 0),
    status varchar(16) NOT NULL CHECK(status IN ('pending','delivered','retrying','failed')),
    next_retry_at timestamptz,
    response_body text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX webhook_deliveries_retry_idx ON webhook_deliveries(status,next_retry_at) WHERE status IN ('pending','retrying');
