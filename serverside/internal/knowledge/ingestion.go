package knowledge

import (
	"authbackend/generated/sqlc"
	"authbackend/internal/brain/obsidian"
	"authbackend/internal/storage"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Ingestion struct {
	pool    *pgxpool.Pool
	objects storage.Store
	brain   obsidian.SecondBrain
}

func NewIngestion(pool *pgxpool.Pool, objects storage.Store, brain obsidian.SecondBrain) *Ingestion {
	return &Ingestion{pool, objects, brain}
}
func (i *Ingestion) Ingest(ctx context.Context, sourceID string) error {
	id, err := knowledgeUUID(sourceID)
	if err != nil {
		return err
	}
	q := sqlc.New(i.pool)
	row, err := q.GetKnowledgeSourceForIngestion(ctx, id)
	if err != nil {
		return err
	}
	if row.Status == "ready" {
		return nil
	}
	if !row.VaultKey.Valid {
		return errors.New("second brain not ready")
	}
	_ = q.UpdateKnowledgeStatus(ctx, sqlc.UpdateKnowledgeStatusParams{ID: id, Status: "processing"})
	var reader io.ReadCloser
	if strings.HasPrefix(row.OriginalObjectKey.String, "inline:") {
		reader = io.NopCloser(strings.NewReader(strings.TrimPrefix(row.OriginalObjectKey.String, "inline:")))
	} else {
		reader, err = i.objects.Get(ctx, knowledgeID(row.WorkspaceID), row.OriginalObjectKey.String)
		if err != nil {
			return i.fail(ctx, id, "OBJECT_DOWNLOAD_FAILED", err)
		}
	}
	defer reader.Close()
	mime := row.MimeType.String
	notes, err := Convert(Source{ID: sourceID, WorkspaceID: knowledgeID(row.WorkspaceID), Type: row.Type, Title: row.Title}, reader, mime)
	if err != nil {
		return i.fail(ctx, id, "DOCUMENT_PARSE_FAILED", err)
	}
	_ = q.UpdateKnowledgeStatus(ctx, sqlc.UpdateKnowledgeStatusParams{ID: id, Status: "writing"})
	for _, note := range notes {
		if err = i.brain.WriteNote(ctx, row.VaultKey.String, obsidian.Note{Path: note.Path, Content: note.Content}); err != nil {
			return i.fail(ctx, id, "KNOWLEDGE_WRITE_FAILED", err)
		}
		sum := sha256.Sum256([]byte(note.Content))
		nid, _ := knowledgeUUID(uuid.NewString())
		if err = q.CreateKnowledgeNote(ctx, sqlc.CreateKnowledgeNoteParams{ID: nid, WorkspaceID: row.WorkspaceID, SecondBrainID: row.SecondBrainID, SourceID: id, Title: note.Title, Category: note.Category, RelativePath: note.Path, ContentHash: hex.EncodeToString(sum[:])}); err != nil {
			return i.fail(ctx, id, "KNOWLEDGE_WRITE_FAILED", err)
		}
	}
	return q.UpdateKnowledgeStatus(ctx, sqlc.UpdateKnowledgeStatusParams{ID: id, Status: "ready"})
}
func (i *Ingestion) Delete(ctx context.Context, sourceID string) error {
	id, err := knowledgeUUID(sourceID)
	if err != nil {
		return err
	}
	q := sqlc.New(i.pool)
	row, err := q.GetKnowledgeSourceForIngestion(ctx, id)
	if err != nil {
		return err
	}
	paths, err := q.ListKnowledgeNotePaths(ctx, id)
	if err != nil {
		return err
	}
	for _, path := range paths {
		if err = i.brain.DeleteNote(ctx, row.VaultKey.String, path); err != nil {
			return err
		}
	}
	if row.OriginalObjectKey.Valid && !strings.HasPrefix(row.OriginalObjectKey.String, "inline:") {
		if err = i.objects.Delete(ctx, knowledgeID(row.WorkspaceID), row.OriginalObjectKey.String); err != nil && !errors.Is(err, storage.ErrNotFound) {
			return err
		}
	}
	if err = q.DeleteKnowledgeNotesBySource(ctx, id); err != nil {
		return err
	}
	return q.MarkKnowledgeDeleted(ctx, id)
}
func (i *Ingestion) fail(ctx context.Context, id pgtype.UUID, code string, cause error) error {
	_ = sqlc.New(i.pool).UpdateKnowledgeStatus(ctx, sqlc.UpdateKnowledgeStatusParams{ID: id, Status: "failed", ErrorCode: pgtype.Text{String: code, Valid: true}})
	return cause
}
