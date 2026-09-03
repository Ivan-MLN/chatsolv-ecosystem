package auth

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRedisResetTokenTTLAndSingleUse(t *testing.T) {
	m := miniredis.RunT(t)
	r := redis.NewClient(&redis.Options{Addr: m.Addr()})
	s := NewRedisTokenStore(r)
	ctx := context.Background()
	require.NoError(t, s.SaveReset(ctx, "hash", "user", time.Minute))
	require.Equal(t, time.Minute, m.TTL("auth:reset:hash"))
	u, e := s.ConsumeReset(ctx, "hash")
	require.NoError(t, e)
	require.Equal(t, "user", u)
	_, e = s.ConsumeReset(ctx, "hash")
	require.ErrorIs(t, e, ErrInvalidResetToken)
}

func TestRedisResetTokenExpires(t *testing.T) {
	m := miniredis.RunT(t)
	r := redis.NewClient(&redis.Options{Addr: m.Addr()})
	s := NewRedisTokenStore(r)
	ctx := context.Background()
	require.NoError(t, s.SaveReset(ctx, "hash", "user", time.Minute))
	m.FastForward(time.Minute)
	_, err := s.ConsumeReset(ctx, "hash")
	require.ErrorIs(t, err, ErrInvalidResetToken)
}

func TestRedisRefreshTokenRotationConsumesTokenOnce(t *testing.T) {
	m := miniredis.RunT(t)
	r := redis.NewClient(&redis.Options{Addr: m.Addr()})
	s := NewRedisTokenStore(r)
	ctx := context.Background()
	require.NoError(t, s.SaveRefresh(ctx, "user-1", tokenHash("refresh"), time.Hour))
	userID, err := s.ConsumeRefresh(ctx, tokenHash("refresh"))
	require.NoError(t, err)
	require.Equal(t, "user-1", userID)
	_, err = s.ConsumeRefresh(ctx, tokenHash("refresh"))
	require.ErrorIs(t, err, ErrInvalidRefreshToken)
}

func TestRedisRevokesOnlyUsersRefreshTokens(t *testing.T) {
	m := miniredis.RunT(t)
	r := redis.NewClient(&redis.Options{Addr: m.Addr()})
	s := NewRedisTokenStore(r)
	ctx := context.Background()
	require.NoError(t, s.SaveRefresh(ctx, "u1", "a", time.Hour))
	require.NoError(t, s.SaveRefresh(ctx, "u2", "b", time.Hour))
	require.NoError(t, s.RevokeUserSessions(ctx, "u1"))
	require.False(t, m.Exists("auth:refresh:a"))
	require.True(t, m.Exists("auth:refresh:b"))
}
