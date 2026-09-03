package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeUsers struct {
	user                         User
	getErr, createErr, updateErr error
}

func (f *fakeUsers) Create(_ context.Context, u User) error {
	f.user = u
	return f.createErr
}
func (f *fakeUsers) GetByEmail(context.Context, string) (User, error)     { return f.user, f.getErr }
func (f *fakeUsers) UpdatePassword(context.Context, string, string) error { return f.updateErr }

type fakeTokens struct {
	resetUser    string
	refreshUser  string
	consumed     bool
	refreshUsed  bool
	savedRefresh bool
	revoked      bool
	consumeErr   error
}

func (f *fakeTokens) SaveRefresh(context.Context, string, string, time.Duration) error {
	f.savedRefresh = true
	return nil
}
func (f *fakeTokens) ConsumeRefresh(context.Context, string) (string, error) {
	if f.refreshUsed || f.refreshUser == "" {
		return "", ErrInvalidRefreshToken
	}
	f.refreshUsed = true
	return f.refreshUser, nil
}
func (f *fakeTokens) SaveReset(context.Context, string, string, time.Duration) error { return nil }
func (f *fakeTokens) ConsumeReset(context.Context, string) (string, error) {
	if f.consumeErr != nil {
		return "", f.consumeErr
	}
	if f.consumed || f.resetUser == "" {
		return "", ErrInvalidResetToken
	}
	f.consumed = true
	return f.resetUser, nil
}
func (f *fakeTokens) RevokeUserSessions(context.Context, string) error { f.revoked = true; return nil }

type fakeEmail struct{ sent bool }

func (f *fakeEmail) SendPasswordReset(context.Context, string, string) error {
	f.sent = true
	return nil
}
func service(u *fakeUsers, t *fakeTokens, e *fakeEmail) *Service {
	return NewService(u, t, e, NewArgon2Hasher(DefaultArgon2Params()), NewJWTManager([]byte("01234567890123456789012345678901"), time.Minute), time.Hour, time.Minute)
}
func TestRegisterNormalizesEmailAndHashesPassword(t *testing.T) {
	u := &fakeUsers{}
	s := service(u, &fakeTokens{}, &fakeEmail{})
	got, err := s.Register(context.Background(), RegisterInput{Name: " John ", Email: "John@Example.COM", Password: "password123"})
	require.NoError(t, err)
	require.Equal(t, "john@example.com", got.Email)
	require.NotEqual(t, "password123", u.user.PasswordHash)
}
func TestRegisterDuplicate(t *testing.T) {
	u := &fakeUsers{createErr: ErrUserExists}
	_, err := service(u, &fakeTokens{}, &fakeEmail{}).Register(context.Background(), RegisterInput{Name: "John", Email: "j@example.com", Password: "password123"})
	require.ErrorIs(t, err, ErrUserExists)
}
func TestLogin(t *testing.T) {
	h := NewArgon2Hasher(DefaultArgon2Params())
	hash, _ := h.Hash("password123")
	u := &fakeUsers{user: User{ID: "u1", Email: "j@example.com", PasswordHash: hash}}
	ts := &fakeTokens{}
	out, err := service(u, ts, &fakeEmail{}).Login(context.Background(), LoginInput{Email: "J@EXAMPLE.COM", Password: "password123"})
	require.NoError(t, err)
	require.NotEmpty(t, out.AccessToken)
	require.NotEmpty(t, out.RefreshToken)
	require.True(t, ts.savedRefresh)
}
func TestLoginInvalidCredentials(t *testing.T) {
	u := &fakeUsers{getErr: ErrUserNotFound}
	_, err := service(u, &fakeTokens{}, &fakeEmail{}).Login(context.Background(), LoginInput{Email: "x@y.com", Password: "bad"})
	require.ErrorIs(t, err, ErrInvalidCredentials)
	u = &fakeUsers{user: User{PasswordHash: "bad"}}
	_, err = service(u, &fakeTokens{}, &fakeEmail{}).Login(context.Background(), LoginInput{Email: "x@y.com", Password: "bad"})
	require.ErrorIs(t, err, ErrInvalidCredentials)
}
func TestRefreshRotatesSingleUseToken(t *testing.T) {
	tokens := &fakeTokens{refreshUser: "user-1"}
	svc := service(&fakeUsers{}, tokens, &fakeEmail{})
	result, err := svc.Refresh(context.Background(), "old-refresh-token")
	require.NoError(t, err)
	require.NotEmpty(t, result.AccessToken)
	require.NotEmpty(t, result.RefreshToken)
	require.NotEqual(t, "old-refresh-token", result.RefreshToken)
	require.True(t, tokens.savedRefresh)
	_, err = svc.Refresh(context.Background(), "old-refresh-token")
	require.ErrorIs(t, err, ErrInvalidRefreshToken)
}
func TestForgotEnumerationSafe(t *testing.T) {
	e := &fakeEmail{}
	err := service(&fakeUsers{getErr: ErrUserNotFound}, &fakeTokens{}, e).ForgotPassword(context.Background(), "none@example.com")
	require.NoError(t, err)
	require.False(t, e.sent)
}
func TestForgotSendsForKnownUser(t *testing.T) {
	e := &fakeEmail{}
	err := service(&fakeUsers{user: User{ID: "u1", Email: "x@y.com"}}, &fakeTokens{}, e).ForgotPassword(context.Background(), "x@y.com")
	require.NoError(t, err)
	require.True(t, e.sent)
}
func TestResetSingleUseAndRevokesSessions(t *testing.T) {
	ts := &fakeTokens{resetUser: "u1"}
	u := &fakeUsers{}
	s := service(u, ts, &fakeEmail{})
	require.NoError(t, s.ResetPassword(context.Background(), "token", "newpassword123"))
	require.True(t, ts.revoked)
	require.ErrorIs(t, s.ResetPassword(context.Background(), "token", "newpassword123"), ErrInvalidResetToken)
}
func TestResetStoreError(t *testing.T) {
	ts := &fakeTokens{}
	err := service(&fakeUsers{}, ts, &fakeEmail{}).ResetPassword(context.Background(), "bad", "newpassword123")
	require.True(t, errors.Is(err, ErrInvalidResetToken))
}

func TestResetInfrastructureErrorIsPropagated(t *testing.T) {
	storeErr := errors.New("redis unavailable")
	ts := &fakeTokens{consumeErr: storeErr}
	err := service(&fakeUsers{}, ts, &fakeEmail{}).ResetPassword(context.Background(), "token", "newpassword123")
	require.ErrorIs(t, err, storeErr)
}
