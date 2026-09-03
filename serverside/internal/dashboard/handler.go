package dashboard

import (
	"errors"

	"authbackend/internal/auth"
	"authbackend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Me(c *fiber.Ctx) error {
	result, err := h.service.Me(c.UserContext(), auth.AuthenticatedUserID(c))
	if err != nil {
		return dashboardError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Current user retrieved", result)
}

func (h *Handler) Overview(c *fiber.Ctx) error {
	result, err := h.service.Overview(c.UserContext(), auth.AuthenticatedUserID(c), c.Query("workspace_id"))
	if err != nil {
		return dashboardError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Dashboard retrieved", result)
}

func dashboardError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidWorkspace):
		return response.Fail(c, fiber.StatusBadRequest, "workspace_id is required", "VALIDATION_ERROR")
	case errors.Is(err, ErrNotFound):
		return response.Fail(c, fiber.StatusNotFound, "Dashboard workspace not found", "WORKSPACE_NOT_FOUND")
	default:
		return response.Fail(c, fiber.StatusInternalServerError, "Dashboard operation failed", "INTERNAL_ERROR")
	}
}
