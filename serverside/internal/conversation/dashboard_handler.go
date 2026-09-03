package conversation

import (
	"errors"
	"strconv"
	"time"

	"authbackend/internal/auth"
	"authbackend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct{ service *DashboardService }

func NewDashboardHandler(service *DashboardService) *DashboardHandler {
	return &DashboardHandler{service}
}
func (h *DashboardHandler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit"))
	items, err := h.service.List(c.UserContext(), auth.AuthenticatedUserID(c), c.Query("workspace_id"), ListFilter{Status: c.Query("status"), Mode: c.Query("mode"), Limit: limit})
	if err != nil {
		return dashboardError(c, err)
	}
	return response.OK(c, 200, "Conversations retrieved", items)
}
func (h *DashboardHandler) Get(c *fiber.Ctx) error {
	item, err := h.service.Get(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("id"))
	if err != nil {
		return dashboardError(c, err)
	}
	return response.OK(c, 200, "Conversation retrieved", item)
}
func (h *DashboardHandler) Messages(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit"))
	var cursor *time.Time
	if raw := c.Query("cursor"); raw != "" {
		value, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return response.Fail(c, 400, "Invalid cursor", "VALIDATION_ERROR")
		}
		cursor = &value
	}
	items, err := h.service.Messages(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("id"), MessageFilter{Cursor: cursor, Limit: limit})
	if err != nil {
		return dashboardError(c, err)
	}
	return response.OK(c, 200, "Messages retrieved", items)
}
func (h *DashboardHandler) SetMode(c *fiber.Ctx) error {
	if !c.Is("json") {
		return response.Fail(c, 400, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	var body struct {
		Mode Mode `json:"mode"`
	}
	if c.BodyParser(&body) != nil {
		return response.Fail(c, 400, "Invalid JSON body", "INVALID_JSON")
	}
	if err := h.service.SetMode(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("id"), body.Mode); err != nil {
		return dashboardError(c, err)
	}
	return response.OK(c, 200, "Conversation mode updated", fiber.Map{"id": c.Params("id"), "mode": body.Mode})
}
func dashboardError(c *fiber.Ctx, err error) error {
	if errors.Is(err, ErrInvalidDashboardInput) {
		return response.Fail(c, 400, "Invalid conversation input", "VALIDATION_ERROR")
	}
	if errors.Is(err, ErrDashboardForbidden) {
		return response.Fail(c, 403, "You cannot perform this action", "FORBIDDEN")
	}
	return response.Fail(c, 404, "Conversation not found", "CONVERSATION_NOT_FOUND")
}
