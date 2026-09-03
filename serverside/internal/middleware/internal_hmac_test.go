package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func sign(secret, timestamp, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "." + body))
	return hex.EncodeToString(mac.Sum(nil))
}
func TestInternalHMACAcceptsValidSignature(t *testing.T) {
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	app := fiber.New()
	app.Post("/", InternalHMAC("01234567890123456789012345678901", 5*time.Minute, func() time.Time { return now }), func(c *fiber.Ctx) error { return c.SendStatus(204) })
	body := `{"event":"connected"}`
	timestamp := now.Format(time.RFC3339)
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("X-ChatSolv-Timestamp", timestamp)
	req.Header.Set("X-ChatSolv-Signature", sign("01234567890123456789012345678901", timestamp, body))
	res, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, 204, res.StatusCode)
}
func TestInternalHMACRejectsReplayAndTampering(t *testing.T) {
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	app := fiber.New()
	app.Post("/", InternalHMAC("01234567890123456789012345678901", 5*time.Minute, func() time.Time { return now }), func(c *fiber.Ctx) error { return c.SendStatus(204) })
	for _, timestamp := range []string{now.Add(-10 * time.Minute).Format(time.RFC3339), now.Format(time.RFC3339)} {
		body := "tampered"
		req := httptest.NewRequest("POST", "/", strings.NewReader(body))
		req.Header.Set("X-ChatSolv-Timestamp", timestamp)
		req.Header.Set("X-ChatSolv-Signature", sign("01234567890123456789012345678901", timestamp, "different"))
		res, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 401, res.StatusCode)
	}
}
