package apikey

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

type fakeRepo struct{ stored Record }

func (f *fakeRepo) Create(_ context.Context, r Record) error             { f.stored = r; return nil }
func (f *fakeRepo) FindByPrefix(context.Context, string) (Record, error) { return f.stored, nil }
func (f *fakeRepo) AuthorizeWorkspace(context.Context, string, string) (string, error) {
	return "owner", nil
}
func (f *fakeRepo) List(context.Context, string, string) ([]Record, error) {
	return []Record{f.stored}, nil
}
func (f *fakeRepo) Revoke(context.Context, string, string) error { return nil }
func TestCreateStoresHashAndReturnsFullKeyOnce(t *testing.T) {
	repo := &fakeRepo{}
	result, err := NewService(repo).Create(context.Background(), "wsp", "production", []string{"agent:invoke"})
	require.NoError(t, err)
	require.Contains(t, result.Secret, "cs_live_")
	require.NotContains(t, repo.stored.Hash, result.Secret)
	require.NotEmpty(t, repo.stored.Hash)
	require.Equal(t, result.Secret[len(result.Secret)-4:], repo.stored.LastFour)
}
func TestAuthenticateChecksScopeAndHash(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	created, err := svc.Create(context.Background(), "wsp", "site", []string{"agent:invoke"})
	require.NoError(t, err)
	record, err := svc.Authenticate(context.Background(), created.Secret, "agent:invoke")
	require.NoError(t, err)
	require.Equal(t, "wsp", record.WorkspaceID)
	_, err = svc.Authenticate(context.Background(), created.Secret, "knowledge:write")
	require.ErrorIs(t, err, ErrForbidden)
}
