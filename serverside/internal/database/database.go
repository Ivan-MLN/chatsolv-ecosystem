package database

import (
	"authbackend/internal/config"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func Postgres(ctx context.Context, c config.Config) (*pgxpool.Pool, error) {
	p, e := pgxpool.ParseConfig(c.DatabaseURL)
	if e != nil {
		return nil, e
	}
	p.MaxConns = c.DatabaseMax
	p.MinConns = c.DatabaseMin
	p.MaxConnLifetime = c.DatabaseLifetime
	p.MaxConnIdleTime = c.DatabaseIdle
	return pgxpool.NewWithConfig(ctx, p)
}
func Redis(c config.Config) (*redis.Client, error) {
	o, e := redis.ParseURL(c.RedisURL)
	if e != nil {
		return nil, e
	}
	return redis.NewClient(o), nil
}
