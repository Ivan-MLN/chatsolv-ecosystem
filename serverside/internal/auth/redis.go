package auth

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisTokenStore struct{ r *redis.Client }

func NewRedisTokenStore(r *redis.Client) *RedisTokenStore { return &RedisTokenStore{r} }
func (s *RedisTokenStore) SaveRefresh(ctx context.Context, uid, hash string, ttl time.Duration) error {
	return s.r.Set(ctx, "auth:refresh:"+hash, uid, ttl).Err()
}
func (s *RedisTokenStore) SaveReset(ctx context.Context, hash, uid string, ttl time.Duration) error {
	return s.r.Set(ctx, "auth:reset:"+hash, uid, ttl).Err()
}

var consumeReset = redis.NewScript(`local v=redis.call('GET',KEYS[1]); if not v then return false end; redis.call('DEL',KEYS[1]); return v`)

func (s *RedisTokenStore) ConsumeRefresh(ctx context.Context, hash string) (string, error) {
	v, e := consumeReset.Run(ctx, s.r, []string{"auth:refresh:" + hash}).Text()
	if errors.Is(e, redis.Nil) {
		return "", ErrInvalidRefreshToken
	}
	return v, e
}

func (s *RedisTokenStore) ConsumeReset(ctx context.Context, hash string) (string, error) {
	v, e := consumeReset.Run(ctx, s.r, []string{"auth:reset:" + hash}).Text()
	if errors.Is(e, redis.Nil) {
		return "", ErrInvalidResetToken
	}
	return v, e
}
func (s *RedisTokenStore) RevokeUserSessions(ctx context.Context, uid string) error {
	var cur uint64
	for {
		keys, next, e := s.r.Scan(ctx, cur, "auth:refresh:*", 100).Result()
		if e != nil {
			return e
		}
		owned := make([]string, 0, len(keys))
		for _, key := range keys {
			value, getErr := s.r.Get(ctx, key).Result()
			if getErr != nil && !errors.Is(getErr, redis.Nil) {
				return getErr
			}
			if value == uid {
				owned = append(owned, key)
			}
		}
		if len(owned) > 0 {
			if e = s.r.Del(ctx, owned...).Err(); e != nil {
				return e
			}
		}
		cur = next
		if cur == 0 {
			return nil
		}
	}
}
