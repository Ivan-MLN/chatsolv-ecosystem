package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"authbackend/pkg/response"
	"github.com/gofiber/fiber/v2"
)

func InternalHMAC(secret string, tolerance time.Duration, now func() time.Time) fiber.Handler {
	return func(c *fiber.Ctx) error {
		timestamp := c.Get("X-ChatSolv-Timestamp")
		signature := c.Get("X-ChatSolv-Signature")
		parsed, err := time.Parse(time.RFC3339, timestamp)
		if err != nil || now().Sub(parsed) > tolerance || parsed.Sub(now()) > tolerance {
			return response.Fail(c, fiber.StatusUnauthorized, "Invalid internal signature", "UNAUTHORIZED")
		}
		provided, err := hex.DecodeString(signature)
		if err != nil {
			return response.Fail(c, fiber.StatusUnauthorized, "Invalid internal signature", "UNAUTHORIZED")
		}
		mac := hmac.New(sha256.New, []byte(secret))
		_, _ = mac.Write([]byte(timestamp + "." + string(c.Body())))
		if !hmac.Equal(provided, mac.Sum(nil)) {
			return response.Fail(c, fiber.StatusUnauthorized, "Invalid internal signature", "UNAUTHORIZED")
		}
		return c.Next()
	}
}
