package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env, Port, DatabaseURL, RedisURL                                                 string
	VaultRoot, HermesBinary, HermesRoot, HermesTemplateProfile                       string
	DatabaseMax, DatabaseMin                                                         int32
	DatabaseLifetime, DatabaseIdle, ShutdownTimeout, AccessTTL, RefreshTTL, ResetTTL time.Duration
	JWTSecret                                                                        string
	CORSOrigins                                                                      []string
	LogLevel                                                                         string
	BodyLimit                                                                        int
	RateLimit                                                                        int64
	RateWindow                                                                       time.Duration
	JobPollInterval                                                                  time.Duration
	JobMaxAttempts                                                                   int
	ObjectStorageEndpoint, ObjectStorageBucket                                       string
	ObjectStorageAccessKey, ObjectStorageSecretKey                                   string
	ObjectStorageUseSSL                                                              bool
	WhatsAppServiceURL, InternalServiceSecret                                        string
	WebhookEncryptionKey                                                             string
}

func duration(k, d string) (time.Duration, error) { return time.ParseDuration(env(k, d)) }
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func integer(k, d string) (int, error) { return strconv.Atoi(env(k, d)) }
func Load() (Config, error) {
	var c Config
	c.Env = env("APP_ENV", "development")
	c.Port = env("APP_PORT", "3000")
	c.DatabaseURL = os.Getenv("DATABASE_URL")
	c.RedisURL = env("REDIS_URL", "redis://localhost:6379")
	c.JWTSecret = os.Getenv("JWT_SECRET")
	rawOrigins := strings.Split(env("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173,http://127.0.0.1:3000,http://127.0.0.1:5173,https://cs.naeladtya.my.id"), ",")
	var origins []string
	for _, o := range rawOrigins {
		trimmed := strings.TrimRight(strings.TrimSpace(o), "/")
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	c.CORSOrigins = origins
	c.LogLevel = env("LOG_LEVEL", "info")
	c.VaultRoot = env("VAULT_ROOT", "/data/chatsolv/vaults")
	c.HermesBinary = env("HERMES_BINARY", "hermes")
	c.HermesRoot = env("HERMES_ROOT", "/data/chatsolv/hermes")
	c.HermesTemplateProfile = env("HERMES_TEMPLATE_PROFILE", "default")
	var e error
	c.ObjectStorageEndpoint = env("OBJECT_STORAGE_ENDPOINT", "localhost:9000")
	c.ObjectStorageBucket = env("OBJECT_STORAGE_BUCKET", "chatsolv-originals")
	c.ObjectStorageAccessKey = env("OBJECT_STORAGE_ACCESS_KEY", "chatsolv")
	c.ObjectStorageSecretKey = env("OBJECT_STORAGE_SECRET_KEY", "chatsolv-development-secret")
	c.WhatsAppServiceURL = env("WHATSAPP_SERVICE_URL", "http://localhost:4010")
	c.InternalServiceSecret = env("INTERNAL_SERVICE_SECRET", "replace-with-at-least-32-random-bytes")
	c.WebhookEncryptionKey = env("WEBHOOK_ENCRYPTION_KEY", "change-me-32-byte-encryption-key")
	c.ObjectStorageUseSSL, e = strconv.ParseBool(env("OBJECT_STORAGE_USE_SSL", "false"))
	if e != nil {
		return c, e
	}
	mx, e := integer("DATABASE_MAX_CONNS", "20")
	if e != nil {
		return c, e
	}
	mn, e := integer("DATABASE_MIN_CONNS", "5")
	if e != nil {
		return c, e
	}
	c.DatabaseMax = int32(mx)
	c.DatabaseMin = int32(mn)
	c.BodyLimit, e = integer("REQUEST_BODY_LIMIT", "16384")
	if e != nil {
		return c, e
	}
	rl, e := integer("RATE_LIMIT_MAX", "10")
	if e != nil {
		return c, e
	}
	c.RateLimit = int64(rl)
	c.JobMaxAttempts, e = integer("JOB_MAX_ATTEMPTS", "5")
	if e != nil {
		return c, e
	}
	for _, x := range []struct {
		k, d string
		p    *time.Duration
	}{{"DATABASE_MAX_CONN_LIFETIME", "1h", &c.DatabaseLifetime}, {"DATABASE_MAX_CONN_IDLE_TIME", "30m", &c.DatabaseIdle}, {"SHUTDOWN_TIMEOUT", "10s", &c.ShutdownTimeout}, {"JWT_ACCESS_TTL", "15m", &c.AccessTTL}, {"JWT_REFRESH_TTL", "720h", &c.RefreshTTL}, {"PASSWORD_RESET_TTL", "15m", &c.ResetTTL}, {"RATE_LIMIT_WINDOW", "1m", &c.RateWindow}, {"JOB_POLL_INTERVAL", "2s", &c.JobPollInterval}} {
		if *x.p, e = duration(x.k, x.d); e != nil {
			return c, e
		}
	}
	if c.DatabaseURL == "" || len(c.JWTSecret) < 32 {
		return c, errors.New("DATABASE_URL and JWT_SECRET (minimum 32 bytes) are required")
	}
	if c.Env != "development" && c.Env != "test" {
		return c, errors.New("production EmailSender is not configured; APP_ENV must remain development or test")
	}
	if c.DatabaseMin < 0 || c.DatabaseMax < 1 || c.DatabaseMin > c.DatabaseMax || c.BodyLimit < 1024 || c.RateLimit < 1 || c.JobMaxAttempts < 1 || c.VaultRoot == "" || c.HermesRoot == "" || c.HermesBinary == "" || c.HermesTemplateProfile == "" || c.ObjectStorageEndpoint == "" || c.ObjectStorageBucket == "" || c.ObjectStorageAccessKey == "" || c.ObjectStorageSecretKey == "" || c.WhatsAppServiceURL == "" || len(c.InternalServiceSecret) < 32 || len(c.WebhookEncryptionKey) != 32 {
		return c, errors.New("invalid database pool, request body, or rate limit configuration")
	}
	if c.DatabaseLifetime <= 0 || c.DatabaseIdle <= 0 || c.ShutdownTimeout <= 0 || c.AccessTTL <= 0 || c.RefreshTTL <= 0 || c.ResetTTL <= 0 || c.RateWindow <= 0 || c.JobPollInterval <= 0 {
		return c, errors.New("all configured durations must be positive")
	}
	return c, nil
}
