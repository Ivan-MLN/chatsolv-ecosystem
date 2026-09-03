CREATE TABLE knowledge_sources (
 id uuid PRIMARY KEY, workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE, second_brain_id uuid NOT NULL REFERENCES second_brains(id) ON DELETE CASCADE,
 type varchar(20) NOT NULL CHECK(type IN('document','text','faq','product','structured')), title varchar(240) NOT NULL,
 original_filename text, mime_type varchar(150), file_size bigint CHECK(file_size IS NULL OR file_size>=0), original_object_key text, checksum char(64),
 status varchar(20) NOT NULL CHECK(status IN('queued','processing','converting','writing','syncing','ready','failed','deleting','deleted')), error_code varchar(80),
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(), deleted_at timestamptz,
 UNIQUE(workspace_id,checksum)
);
CREATE TABLE knowledge_notes (
 id uuid PRIMARY KEY, workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE, second_brain_id uuid NOT NULL REFERENCES second_brains(id) ON DELETE CASCADE,
 source_id uuid NOT NULL REFERENCES knowledge_sources(id) ON DELETE CASCADE, title varchar(240) NOT NULL, category varchar(80) NOT NULL,
 relative_path text NOT NULL, content_hash char(64) NOT NULL, version bigint NOT NULL DEFAULT 1 CHECK(version>0),
 status varchar(20) NOT NULL CHECK(status IN('active','inactive','deleting','deleted')), created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(), deleted_at timestamptz,
 UNIQUE(workspace_id,relative_path,version)
);
CREATE INDEX knowledge_sources_workspace_idx ON knowledge_sources(workspace_id,status,created_at DESC);
CREATE INDEX knowledge_notes_workspace_source_idx ON knowledge_notes(workspace_id,source_id);
CREATE TABLE object_metadata (
 object_key text PRIMARY KEY, workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE, size bigint NOT NULL CHECK(size>=0), checksum char(64) NOT NULL, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX object_metadata_workspace_idx ON object_metadata(workspace_id,object_key);
