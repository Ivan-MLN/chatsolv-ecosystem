package knowledge

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"authbackend/internal/storage"
)

var (
	ErrForbidden         = errors.New("knowledge forbidden")
	ErrInvalidInput      = errors.New("invalid knowledge input")
	ErrDocumentTooLarge  = errors.New("document too large")
	ErrDocumentDuplicate = errors.New("document already exists")
	ErrDocumentLimit     = errors.New("document limit reached")
	ErrStorageLimit      = errors.New("knowledge storage limit reached")
)

type SourceRecord struct {
	ID, WorkspaceID, SecondBrainID, Type, Title, OriginalFilename, MIMEType, ObjectKey, Checksum, Status, ErrorCode string
	FileSize                                                                                                        int64
	CreatedAt, UpdatedAt                                                                                            time.Time
}
type WorkspaceAccess struct{ Role, SecondBrainID, VaultKey string }
type Repository interface {
	ResolveWorkspace(context.Context, string, string) (WorkspaceAccess, error)
	CreateSourceAndQueue(context.Context, SourceRecord, string) (SourceRecord, error)
	List(context.Context, string, int, int) ([]SourceRecord, error)
	GetForMember(context.Context, string, string) (SourceRecord, error)
	UpdateAndQueue(context.Context, string, string, string) error
	DeleteAndQueue(context.Context, string, string) error
	RetryAndQueue(context.Context, string, string) error
	Count(context.Context, string) (int64, error)
	StorageBytes(context.Context, string) (int64, error)
}
type Upload struct {
	WorkspaceID, UserID, Title, Filename, MIMEType string
	Size                                           int64
	Reader                                         io.Reader
	Unlimited                                      bool
}
type Service struct {
	repository   Repository
	objects      storage.Store
	maxBytes     int64
	maxDocuments int64
}

func NewService(repository Repository, objects storage.Store, maxBytes, maxDocuments int64) *Service {
	return &Service{repository, objects, maxBytes, maxDocuments}
}
func (s *Service) UploadDocument(ctx context.Context, input Upload) (SourceRecord, error) {
	access, err := s.repository.ResolveWorkspace(ctx, input.UserID, input.WorkspaceID)
	if err != nil {
		return SourceRecord{}, err
	}
	if access.Role != "owner" && access.Role != "admin" && access.Role != "member" {
		return SourceRecord{}, ErrForbidden
	}
	if strings.TrimSpace(input.Title) == "" || input.Reader == nil || input.Size <= 0 {
		return SourceRecord{}, ErrInvalidInput
	}
	if input.Size > s.maxBytes {
		return SourceRecord{}, ErrDocumentTooLarge
	}
	if !supportedMIME(input.MIMEType) {
		return SourceRecord{}, ErrUnsupportedType
	}
	if !input.Unlimited {
		count, err := s.repository.Count(ctx, input.WorkspaceID)
		if err != nil {
			return SourceRecord{}, err
		}
		if count >= s.maxDocuments {
			return SourceRecord{}, ErrDocumentLimit
		}
		used, err := s.repository.StorageBytes(ctx, input.WorkspaceID)
		if err != nil {
			return SourceRecord{}, err
		}
		if used+input.Size > 2<<30 {
			return SourceRecord{}, ErrStorageLimit
		}
	}
	object, err := s.objects.Put(ctx, input.WorkspaceID, input.Filename, io.LimitReader(input.Reader, s.maxBytes+1))
	if err != nil {
		return SourceRecord{}, err
	}
	record := SourceRecord{WorkspaceID: input.WorkspaceID, SecondBrainID: access.SecondBrainID, Type: "document", Title: strings.TrimSpace(input.Title), OriginalFilename: input.Filename, MIMEType: input.MIMEType, ObjectKey: object.Key, Checksum: object.Checksum, FileSize: object.Size, Status: "queued"}
	record, err = s.repository.CreateSourceAndQueue(ctx, record, "knowledge.ingest")
	if err != nil {
		_ = s.objects.Delete(ctx, input.WorkspaceID, object.Key)
		return SourceRecord{}, err
	}
	return record, nil
}
func (s *Service) CreateText(ctx context.Context, userID, workspaceID, title, body, kind string, unlimited bool) (SourceRecord, error) {
	access, err := s.repository.ResolveWorkspace(ctx, userID, workspaceID)
	if err != nil {
		return SourceRecord{}, err
	}
	if access.Role == "viewer" {
		return SourceRecord{}, ErrForbidden
	}
	if !oneOfKind(kind) || strings.TrimSpace(title) == "" || strings.TrimSpace(body) == "" || len(body) > 2<<20 {
		return SourceRecord{}, ErrInvalidInput
	}
	if !unlimited {
		count, err := s.repository.Count(ctx, workspaceID)
		if err != nil {
			return SourceRecord{}, err
		}
		if count >= s.maxDocuments {
			return SourceRecord{}, ErrDocumentLimit
		}
	}
	record := SourceRecord{WorkspaceID: workspaceID, SecondBrainID: access.SecondBrainID, Type: kind, Title: strings.TrimSpace(title), MIMEType: "text/plain", Status: "queued", ObjectKey: "inline:" + body}
	record, err = s.repository.CreateSourceAndQueue(ctx, record, "knowledge.ingest")
	if err != nil {
		return SourceRecord{}, err
	}
	return record, nil
}
func (s *Service) Update(ctx context.Context, userID, sourceID, title string) error {
	title = strings.TrimSpace(title)
	if title == "" || len(title) > 240 {
		return ErrInvalidInput
	}
	return s.repository.UpdateAndQueue(ctx, userID, sourceID, title)
}
func (s *Service) Delete(ctx context.Context, userID, sourceID string) error {
	return s.repository.DeleteAndQueue(ctx, userID, sourceID)
}
func (s *Service) Retry(ctx context.Context, userID, sourceID string) error {
	return s.repository.RetryAndQueue(ctx, userID, sourceID)
}
func supportedMIME(value string) bool {
	switch value {
	case "application/pdf", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "text/plain", "text/csv", "application/csv", "text/markdown":
		return true
	}
	return false
}
func oneOfKind(value string) bool { return value == "text" || value == "faq" }
