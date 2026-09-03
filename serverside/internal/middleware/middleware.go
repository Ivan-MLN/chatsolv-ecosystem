package middleware

import (
	"authbackend/pkg/response"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"strings"
	"time"
)

func RequestID(c *fiber.Ctx) error {
	id := c.Get("X-Request-ID")
	if _, e := uuid.Parse(id); e != nil {
		id = uuid.NewString()
	}
	c.Set("X-Request-ID", id)
	c.Locals("request_id", id)
	return c.Next()
}
func Logger(l *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		e := c.Next()
		l.Info("request", "request_id", c.Locals("request_id"), "method", c.Method(), "path", c.Path(), "status", c.Response().StatusCode(), "latency", time.Since(start))
		return e
	}
}
func SecurityHeaders(c *fiber.Ctx) error {
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("X-Frame-Options", "DENY")
	c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
	c.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
	if strings.HasPrefix(c.Path(), "/api/") || strings.HasPrefix(c.Path(), "/internal/") {
		c.Set("Cache-Control", "no-store")
	}
	return c.Next()
}

var rateScript = redis.NewScript(`local n=redis.call('INCR',KEYS[1]);if n==1 then redis.call('PEXPIRE',KEYS[1],ARGV[1]) end;return n`)

func RateLimit(r *redis.Client, max int64, w time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := "rate:" + c.Path() + ":" + c.IP()
		n, e := rateScript.Run(c.UserContext(), r, []string{key}, w.Milliseconds()).Int64()
		if e != nil {
			return response.Fail(c, 500, "Something went wrong", "INTERNAL_ERROR")
		}
		if n > max {
			return response.Fail(c, 429, "Too many requests", "RATE_LIMITED")
		}
		return c.Next()
	}
}
