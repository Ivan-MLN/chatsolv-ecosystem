package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestArgon2HashAndVerify(t *testing.T) {
	h := NewArgon2Hasher(DefaultArgon2Params())
	s, e := h.Hash("secret123")
	require.NoError(t, e)
	require.NotContains(t, s, "secret123")
	ok, e := h.Verify("secret123", s)
	require.NoError(t, e)
	require.True(t, ok)
	ok, _ = h.Verify("wrong", s)
	require.False(t, ok)
}
func TestJWTClaims(t *testing.T) {
	m := NewJWTManager([]byte("01234567890123456789012345678901"), time.Minute)
	s, _, e := m.Generate("u1")
	require.NoError(t, e)
	tok, e := jwt.Parse(s, func(*jwt.Token) (any, error) { return m.secret, nil }, jwt.WithValidMethods([]string{"HS256"}))
	require.NoError(t, e)
	c := tok.Claims.(jwt.MapClaims)
	require.Equal(t, "u1", c["sub"])
	require.NotEmpty(t, c["jti"])
}
