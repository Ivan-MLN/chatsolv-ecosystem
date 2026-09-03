ALTER TABLE workspaces DROP CONSTRAINT IF EXISTS workspaces_owner_membership_deferred_note;
DROP TABLE IF EXISTS failed_jobs;
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS second_brains;
DROP TABLE IF EXISTS agents;
