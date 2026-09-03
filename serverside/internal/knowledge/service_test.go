package knowledge

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"authbackend/internal/storage"

	"github.com/stretchr/testify/require"
)

type fakeRepository struct {
	access        WorkspaceAccess
	created       SourceRecord
	listed        []SourceRecord
	createErr     error
	atomicCreates int
	legacyCreates int
	legacyQueues  int
	updated       bool
	deleted       bool
	retried       bool
	mutationErr   error
	count         int64
	storageBytes  int64
}

func (f *fakeRepository) ResolveWorkspace(context.Context, string, string) (WorkspaceAccess, error) {
	return f.access, nil
}
func (f *fakeRepository) CreateSourceAndQueue(_ context.Context, source SourceRecord, eventType string) (SourceRecord, error) {
	f.atomicCreates++
	if f.createErr != nil {
		return SourceRecord{}, f.createErr
	}
	source.ID = "source-id"
	f.created = source
	return source, nil
}
func (f *fakeRepository) List(context.Context, string, int, int) ([]SourceRecord, error) {
	return f.listed, nil
}
func (f *fakeRepository) GetForMember(context.Context, string, string) (SourceRecord, error) {
	return SourceRecord{}, nil
}
func (f *fakeRepository) UpdateAndQueue(context.Context, string, string, string) error {
	f.updated = true
	return f.mutationErr
}
func (f *fakeRepository) DeleteAndQueue(context.Context, string, string) error {
	f.deleted = true
	return f.mutationErr
}
func (f *fakeRepository) RetryAndQueue(context.Context, string, string) error {
	f.retried = true
	return f.mutationErr
}
func (f *fakeRepository) Count(context.Context, string) (int64, error) { return f.count, nil }
func (f *fakeRepository) StorageBytes(context.Context, string) (int64, error) {
	return f.storageBytes, nil
}

type fakeStore struct {
	put       bool
	deleted   bool
	deleteErr error
}

func (f *fakeStore) Put(context.Context, string, string, io.Reader) (storage.Object, error) {
	f.put = true
	return storage.Object{Key: "wsp_a/object.txt", Checksum: "checksum", Size: 7}, nil
}
func (f *fakeStore) Get(context.Context, string, string) (io.ReadCloser, error) {
	return nil, storage.ErrNotFound
}
func (f *fakeStore) Delete(context.Context, string, string) error {
	f.deleted = true
	return f.deleteErr
}

func TestUploadDocumentCreatesSourceAndOutboxAtomically(t *testing.T) {
	repository := &fakeRepository{access: WorkspaceAccess{Role: "owner", SecondBrainID: "brain-id"}}
	objects := &fakeStore{}
	service := NewService(repository, objects, 1024, 10)

	result, err := service.UploadDocument(context.Background(), Upload{
		WorkspaceID: "wsp_a", UserID: "user-a", Title: "Document", Filename: "document.txt",
		MIMEType: "text/plain", Size: 7, Reader: bytes.NewBufferString("content"),
	})

	require.NoError(t, err)
	require.Equal(t, "source-id", result.ID)
	require.Equal(t, 1, repository.atomicCreates)
	require.Zero(t, repository.legacyCreates)
	require.Zero(t, repository.legacyQueues)
}

func TestUploadDocumentRemovesObjectWhenAtomicPersistenceFails(t *testing.T) {
	repository := &fakeRepository{access: WorkspaceAccess{Role: "owner", SecondBrainID: "brain-id"}, createErr: errors.New("outbox insert failed")}
	objects := &fakeStore{}
	service := NewService(repository, objects, 1024, 10)

	_, err := service.UploadDocument(context.Background(), Upload{
		WorkspaceID: "wsp_a", UserID: "user-a", Title: "Document", Filename: "document.txt",
		MIMEType: "text/plain", Size: 7, Reader: bytes.NewBufferString("content"),
	})

	require.Error(t, err)
	require.True(t, objects.deleted)
}

func TestKnowledgeQuotaBlocksCustomersButNotDevelopers(t *testing.T) {
	repository := &fakeRepository{
		access: WorkspaceAccess{Role: "owner", SecondBrainID: "brain-id"},
		count:  10,
	}
	service := NewService(repository, &fakeStore{}, 1024, 10)
	upload := Upload{
		WorkspaceID: "wsp_a", UserID: "user-a", Title: "Document", Filename: "document.txt",
		MIMEType: "text/plain", Size: 7, Reader: bytes.NewBufferString("content"),
	}

	_, err := service.UploadDocument(context.Background(), upload)
	require.ErrorIs(t, err, ErrDocumentLimit)

	upload.Unlimited = true
	_, err = service.UploadDocument(context.Background(), upload)
	require.NoError(t, err)
}

func TestKnowledgeStorageLimitIsEnforced(t *testing.T) {
	repository := &fakeRepository{
		access:       WorkspaceAccess{Role: "owner", SecondBrainID: "brain-id"},
		storageBytes: (2 << 30) - 3,
	}
	service := NewService(repository, &fakeStore{}, 1024, 10)

	_, err := service.UploadDocument(context.Background(), Upload{
		WorkspaceID: "wsp_a", UserID: "user-a", Title: "Document", Filename: "document.txt",
		MIMEType: "text/plain", Size: 7, Reader: bytes.NewBufferString("content"),
	})

	require.ErrorIs(t, err, ErrStorageLimit)
}

func TestKnowledgeLifecycleQueuesTenantScopedMutations(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, &fakeStore{}, 1024, 10)

	require.NoError(t, service.Update(context.Background(), "user-a", "source-a", "Updated title"))
	require.NoError(t, service.Delete(context.Background(), "user-a", "source-a"))
	require.NoError(t, service.Retry(context.Background(), "user-a", "source-a"))
	require.True(t, repository.updated)
	require.True(t, repository.deleted)
	require.True(t, repository.retried)
}

func TestKnowledgeUpdateRejectsBlankTitleBeforePersistence(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, &fakeStore{}, 1024, 10)

	err := service.Update(context.Background(), "user-a", "source-a", "   ")

	require.ErrorIs(t, err, ErrInvalidInput)
	require.False(t, repository.updated)
}
