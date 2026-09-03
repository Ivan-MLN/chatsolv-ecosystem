CREATE TABLE agent_personalities (
 id uuid PRIMARY KEY, workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE, agent_id uuid NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
 bot_name varchar(100) NOT NULL, role varchar(100) NOT NULL, tone varchar(30) NOT NULL, communication_style varchar(40) NOT NULL,
 primary_language varchar(10) NOT NULL, response_length varchar(20) NOT NULL, emoji_usage varchar(20) NOT NULL,
 greeting_style varchar(40) NOT NULL, closing_style varchar(40) NOT NULL, custom_instructions text NOT NULL DEFAULT '',
 behavior_rules jsonb NOT NULL DEFAULT '[]', escalation_rules jsonb NOT NULL DEFAULT '[]', forbidden_topics jsonb NOT NULL DEFAULT '[]', fallback_behavior text NOT NULL DEFAULT '',
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(), UNIQUE(workspace_id,agent_id)
);
CREATE TABLE agent_profiles (
 id uuid PRIMARY KEY, workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE, agent_id uuid NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
 display_name varchar(100) NOT NULL, avatar_object_key text, description text NOT NULL DEFAULT '', greeting_message text NOT NULL DEFAULT '', away_message text NOT NULL DEFAULT '', fallback_message text NOT NULL DEFAULT '', language varchar(10) NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(), UNIQUE(workspace_id,agent_id)
);
CREATE TABLE business_profiles (
 id uuid PRIMARY KEY, workspace_id uuid NOT NULL UNIQUE REFERENCES workspaces(id) ON DELETE CASCADE, business_name varchar(160) NOT NULL, industry varchar(100) NOT NULL,
 business_description text NOT NULL DEFAULT '', website text, email varchar(254), phone varchar(40), address text NOT NULL DEFAULT '', business_hours jsonb NOT NULL DEFAULT '{}', timezone varchar(64) NOT NULL, brand_voice text NOT NULL DEFAULT '', company_values jsonb NOT NULL DEFAULT '[]', created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE business_policies (
 id uuid PRIMARY KEY, workspace_id uuid NOT NULL UNIQUE REFERENCES workspaces(id) ON DELETE CASCADE, shipping_policy text NOT NULL DEFAULT '', refund_policy text NOT NULL DEFAULT '', return_policy text NOT NULL DEFAULT '', warranty_policy text NOT NULL DEFAULT '', payment_policy text NOT NULL DEFAULT '', complaint_policy text NOT NULL DEFAULT '', created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX agent_personalities_workspace_idx ON agent_personalities(workspace_id,agent_id);
CREATE INDEX agent_profiles_workspace_idx ON agent_profiles(workspace_id,agent_id);
