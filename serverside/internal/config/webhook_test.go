package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadRejectsInvalidWebhookEncryptionKey(t *testing.T) {
	validEnvironment(t)
	t.Setenv("WEBHOOK_ENCRYPTION_KEY", "too-short")
	_, err := Load()
	require.Error(t, err)
}
