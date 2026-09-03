package template

import (
	"authbackend/internal/auth"
	"authbackend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *fiber.Ctx) error {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		workspaceID = c.Get("X-Workspace-ID")
	}
	userID := auth.AuthenticatedUserID(c)
	templates, err := h.service.List(c.UserContext(), userID, workspaceID)
	if err != nil {
		if err == ErrForbidden {
			return response.Fail(c, fiber.StatusForbidden, "Akses template ditolak", "FORBIDDEN")
		}
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), "INVALID_INPUT")
	}
	return response.OK(c, fiber.StatusOK, "Daftar template berhasil diambil", templates)
}

func (h *Handler) Apply(c *fiber.Ctx) error {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		workspaceID = c.Get("X-Workspace-ID")
	}
	templateID := c.Params("id")
	userID := auth.AuthenticatedUserID(c)
	tmpl, err := h.service.ApplyTemplate(c.UserContext(), userID, workspaceID, templateID)
	if err != nil {
		if err == ErrForbidden {
			return response.Fail(c, fiber.StatusForbidden, "Akses penerapan template ditolak", "FORBIDDEN")
		}
		if err == ErrTemplateNotFound {
			return response.Fail(c, fiber.StatusNotFound, "Template tidak ditemukan", "NOT_FOUND")
		}
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), "INVALID_INPUT")
	}
	return response.OK(c, fiber.StatusOK, "Template berhasil diterapkan ke AI Agent Anda", tmpl)
}
