package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func validEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://localhost/db")
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
}

func TestLoadRejectsInvalidInteger(t *testing.T) {
	validEnvironment(t)
	t.Setenv("DATABASE_MAX_CONNS", "many")
	_, err := Load()
	require.Error(t, err)
}

func TestLoadRejectsInvalidPoolBounds(t *testing.T) {
	validEnvironment(t)
	t.Setenv("DATABASE_MAX_CONNS", "2")
	t.Setenv("DATABASE_MIN_CONNS", "5")
	_, err := Load()
	require.Error(t, err)
}

func TestLoadRejectsProductionWithDevelopmentEmailSender(t *testing.T) {
	validEnvironment(t)
	t.Setenv("APP_ENV", "production")
	_, err := Load()
	require.Error(t, err)
}

func TestLoadRejectsNonPositiveDurations(t *testing.T) {
	validEnvironment(t)
	t.Setenv("PASSWORD_RESET_TTL", "0s")
	_, err := Load()
	require.Error(t, err)
}
