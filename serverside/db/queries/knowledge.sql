-- name: ResolveKnowledgeWorkspace :one
SELECT wm.role, sb.id AS second_brain_id, COALESCE(sb.vault_key, '') AS vault_key FROM workspace_members wm JOIN second_brains sb ON sb.workspace_id=wm.workspace_id AND sb.deleted_at IS NULL WHERE wm.workspace_id=$1 AND wm.user_id=$2 LIMIT 1;

-- name: CreateKnowledgeSource :one
INSERT INTO knowledge_sources(id,workspace_id,second_brain_id,type,title,original_filename,mime_type,file_size,original_object_key,checksum,status)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'queued')
RETURNING id,workspace_id,second_brain_id,type,title,original_filename,mime_type,file_size,original_object_key,checksum,status,error_code,created_at,updated_at,deleted_at;

-- name: ListKnowledgeSources :many
SELECT id,workspace_id,second_brain_id,type,title,original_filename,mime_type,file_size,original_object_key,checksum,status,error_code,created_at,updated_at,deleted_at FROM knowledge_sources WHERE workspace_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: GetKnowledgeSourceForMember :one
SELECT ks.id,ks.workspace_id,ks.second_brain_id,ks.type,ks.title,ks.original_filename,ks.mime_type,ks.file_size,ks.original_object_key,ks.checksum,ks.status,ks.error_code,ks.created_at,ks.updated_at,ks.deleted_at FROM knowledge_sources ks JOIN workspace_members wm ON wm.workspace_id=ks.workspace_id WHERE ks.id=$1 AND wm.user_id=$2 AND ks.deleted_at IS NULL;

-- name: GetKnowledgeSourceForIngestion :one
SELECT ks.id,ks.workspace_id,ks.second_brain_id,ks.type,ks.title,ks.original_filename,ks.mime_type,ks.file_size,ks.original_object_key,ks.checksum,ks.status,ks.error_code,ks.created_at,ks.updated_at,ks.deleted_at,sb.vault_key FROM knowledge_sources ks JOIN second_brains sb ON sb.id=ks.second_brain_id WHERE ks.id=$1 AND ks.deleted_at IS NULL;

-- name: UpdateKnowledgeStatus :exec
UPDATE knowledge_sources SET status=$2,error_code=$3,updated_at=now() WHERE id=$1;

-- name: CreateKnowledgeNote :exec
INSERT INTO knowledge_notes(id,workspace_id,second_brain_id,source_id,title,category,relative_path,content_hash,version,status)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,1,'active')
ON CONFLICT(workspace_id,relative_path,version) DO UPDATE SET title=EXCLUDED.title,category=EXCLUDED.category,content_hash=EXCLUDED.content_hash,status='active',updated_at=now();

-- name: DeleteKnowledgeNotesBySource :exec
UPDATE knowledge_notes SET status='deleted',deleted_at=now(),updated_at=now() WHERE source_id=$1 AND deleted_at IS NULL;

-- name: ListKnowledgeNotePaths :many
SELECT relative_path FROM knowledge_notes WHERE source_id=$1 AND deleted_at IS NULL AND status='active';

-- name: QueueKnowledgeEvent :exec
INSERT INTO outbox_events(id,workspace_id,event_type,aggregate_type,aggregate_id,payload) VALUES($1,$2,$3,'knowledge_source',$4,$5);

-- name: UpdateKnowledgeTitleForMember :one
UPDATE knowledge_sources ks
SET title=$3,updated_at=now()
WHERE ks.id=$1 AND ks.deleted_at IS NULL AND EXISTS(
  SELECT 1 FROM workspace_members wm
  WHERE wm.workspace_id=ks.workspace_id AND wm.user_id=$2 AND wm.role IN('owner','admin','member')
)
RETURNING ks.workspace_id;

-- name: MarkKnowledgeDeletingForMember :one
UPDATE knowledge_sources ks
SET status='deleting',updated_at=now()
WHERE ks.id=$1 AND ks.deleted_at IS NULL AND ks.status NOT IN('deleting','deleted') AND EXISTS(
  SELECT 1 FROM workspace_members wm
  WHERE wm.workspace_id=ks.workspace_id AND wm.user_id=$2 AND wm.role IN('owner','admin','member')
)
RETURNING ks.workspace_id;

-- name: RetryKnowledgeSourceForMember :one
UPDATE knowledge_sources ks
SET status='queued',error_code=NULL,updated_at=now()
WHERE ks.id=$1 AND ks.status='failed' AND ks.deleted_at IS NULL AND EXISTS(
  SELECT 1 FROM workspace_members wm
  WHERE wm.workspace_id=ks.workspace_id AND wm.user_id=$2 AND wm.role IN('owner','admin','member')
)
RETURNING ks.workspace_id;

-- name: MarkKnowledgeDeleting :exec
UPDATE knowledge_sources SET status='deleting',updated_at=now() WHERE id=$1 AND workspace_id=$2 AND deleted_at IS NULL;

-- name: MarkKnowledgeDeleted :exec
UPDATE knowledge_sources SET status='deleted',deleted_at=now(),updated_at=now() WHERE id=$1;

-- name: RetryKnowledgeSource :exec
UPDATE knowledge_sources SET status='queued',error_code=NULL,updated_at=now() WHERE id=$1 AND workspace_id=$2 AND status='failed' AND deleted_at IS NULL;

-- name: CountKnowledgeDocuments :one
SELECT count(*) FROM knowledge_sources WHERE workspace_id=$1 AND deleted_at IS NULL AND status NOT IN('deleting','deleted');

-- name: CreateObjectMetadata :exec
INSERT INTO object_metadata(object_key,workspace_id,size,checksum) VALUES($1,$2,$3,$4);

-- name: DeleteObjectMetadata :exec
DELETE FROM object_metadata WHERE object_key=$1 AND workspace_id=$2;
