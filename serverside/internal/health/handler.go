package health

import (
	"context"
	"time"

	"authbackend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Check func(context.Context) error
type Handler struct {
	postgres Check
	redis    Check
}

func NewHandler(postgres, redis Check) *Handler { return &Handler{postgres: postgres, redis: redis} }
func (h *Handler) Live(c *fiber.Ctx) error      { return c.JSON(fiber.Map{"status": "ok"}) }
func (h *Handler) Ready(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), time.Second)
	defer cancel()
	if h.postgres(ctx) != nil || h.redis(ctx) != nil {
		return response.Fail(c, fiber.StatusServiceUnavailable, "Service unavailable", "NOT_READY")
	}
	return c.JSON(fiber.Map{"status": "ready"})
}
