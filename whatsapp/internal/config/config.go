package config

import (
	"errors"
	"os"
	"strings"
	"time"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	// Server
	Port string
	Env  string

	// HMAC secret shared with the ChatSolv backend.
	// Must be identical to INTERNAL_SERVICE_SECRET in the backend .env.
	InternalServiceSecret string

	// Base URL of the ChatSolv backend (used for HMAC-signed callbacks).
	BackendURL string

	// Directory where per-channel SQLite databases are stored.
	// One file per channel: <DBRoot>/<channel_id>.db
	DBRoot string

	// Timeout for outbound HTTP callbacks to the backend.
	CallbackTimeout time.Duration

	// Graceful shutdown timeout.
	ShutdownTimeout time.Duration

	// Log level: debug | info | warn | error
	LogLevel string
}

// Load reads config from environment variables and validates required fields.
func Load() (Config, error) {
	var c Config
	c.Port = env("PORT", "4010")
	c.Env = env("APP_ENV", "development")
	c.InternalServiceSecret = env("INTERNAL_SERVICE_SECRET", "")
	c.BackendURL = env("BACKEND_URL", "http://localhost:3000")
	c.DBRoot = env("DB_ROOT", "./data/sessions")
	c.LogLevel = env("LOG_LEVEL", "info")

	var err error
	if c.CallbackTimeout, err = parseDuration("CALLBACK_TIMEOUT", "10s"); err != nil {
		return c, err
	}
	if c.ShutdownTimeout, err = parseDuration("SHUTDOWN_TIMEOUT", "10s"); err != nil {
		return c, err
	}

	if len(c.InternalServiceSecret) < 32 {
		return c, errors.New("INTERNAL_SERVICE_SECRET must be at least 32 bytes")
	}
	if c.BackendURL == "" {
		return c, errors.New("BACKEND_URL is required")
	}
	return c, nil
}

func env(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// splitOrigins parses a comma-separated ALLOWED_ORIGINS env var.
func SplitOrigins(raw string) []string {
	return strings.Split(raw, ",")
}

func parseDuration(key, def string) (time.Duration, error) {
	d, err := time.ParseDuration(env(key, def))
	if err != nil {
		return 0, err
	}
	if d <= 0 {
		return 0, errors.New(key + " must be positive")
	}
	return d, nil
}
