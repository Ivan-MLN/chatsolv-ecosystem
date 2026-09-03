package conversation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrConversationBusy = errors.New("conversation is busy")

const releaseLockScript = `if redis.call("GET", KEYS[1]) == ARGV[1] then return redis.call("DEL", KEYS[1]) else return 0 end`

type RedisLocker struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisLocker(client *redis.Client, ttl time.Duration) *RedisLocker {
	return &RedisLocker{client: client, ttl: ttl}
}

func (l *RedisLocker) WithLock(ctx context.Context, conversationID string, fn func(context.Context) error) error {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return err
	}
	token := hex.EncodeToString(random)
	key := "conversation:" + conversationID + ":agent-lock"
	acquired, err := l.client.SetNX(ctx, key, token, l.ttl).Result()
	if err != nil {
		return err
	}
	if !acquired {
		return ErrConversationBusy
	}
	defer l.client.Eval(context.Background(), releaseLockScript, []string{key}, token)
	return fn(ctx)
}
