package webhook

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

type fakeRepo struct {
	role    string
	enabled bool
	saved   Endpoint
}

func (f *fakeRepo) Authorize(context.Context, string, string) (string, bool, error) {
	return f.role, f.enabled, nil
}
func (f *fakeRepo) Create(_ context.Context, e Endpoint) error { f.saved = e; return nil }
func (f *fakeRepo) List(context.Context, string, string) ([]Endpoint, error) {
	return []Endpoint{f.saved}, nil
}
func (f *fakeRepo) Update(context.Context, string, string, UpdateInput) (Endpoint, error) {
	return f.saved, nil
}
func (f *fakeRepo) Delete(context.Context, string, string) error { return nil }
func TestCreateEncryptsWebhookSecretAndReturnsItOnce(t *testing.T) {
	r := &fakeRepo{role: "owner", enabled: true}
	s := NewService(r, []byte("01234567890123456789012345678901"))
	v, e := s.Create(context.Background(), "u", "w", CreateInput{URL: "https://example.com/hook", Events: []string{"message.created"}})
	require.NoError(t, e)
	require.NotEmpty(t, v.Secret)
	require.NotContains(t, string(r.saved.SecretCiphertext), v.Secret)
}
func TestCreateRejectsNonHTTPSWebhook(t *testing.T) {
	s := NewService(&fakeRepo{role: "owner", enabled: true}, []byte("01234567890123456789012345678901"))
	_, e := s.Create(context.Background(), "u", "w", CreateInput{URL: "http://example.com/hook", Events: []string{"message.created"}})
	require.ErrorIs(t, e, ErrInvalidInput)
}
func TestCreateRequiresWebhookEntitlement(t *testing.T) {
	s := NewService(&fakeRepo{role: "owner"}, []byte("01234567890123456789012345678901"))
	_, e := s.Create(context.Background(), "u", "w", CreateInput{URL: "https://example.com/hook", Events: []string{"message.created"}})
	require.ErrorIs(t, e, ErrEntitlementRequired)
}
func TestCreateRequiresOwnerOrAdmin(t *testing.T) {
	s := NewService(&fakeRepo{role: "member", enabled: true}, []byte("01234567890123456789012345678901"))
	_, e := s.Create(context.Background(), "u", "w", CreateInput{URL: "https://example.com/hook", Events: []string{"message.created"}})
	require.ErrorIs(t, e, ErrForbidden)
}
