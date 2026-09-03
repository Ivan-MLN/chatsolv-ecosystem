package adminmgmt

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
	workspaceID := c.Get("X-Workspace-ID")
	if workspaceID == "" {
		workspaceID = c.Query("workspace_id")
	}
	userID := auth.AuthenticatedUserID(c)
	admins, err := h.service.List(c.UserContext(), userID, workspaceID)
	if err != nil {
		if err == ErrForbidden {
			return response.Fail(c, fiber.StatusForbidden, "Akses tim admin ditolak", "FORBIDDEN")
		}
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), "INVALID_INPUT")
	}
	return response.OK(c, fiber.StatusOK, "Daftar admin berhasil diambil", admins)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	workspaceID := c.Get("X-Workspace-ID")
	if workspaceID == "" {
		workspaceID = c.Query("workspace_id")
	}
	userID := auth.AuthenticatedUserID(c)
	var body WorkspaceAdmin
	if err := c.BodyParser(&body); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Format data tidak valid", "INVALID_JSON")
	}
	created, err := h.service.Create(c.UserContext(), userID, workspaceID, body)
	if err != nil {
		if err == ErrForbidden {
			return response.Fail(c, fiber.StatusForbidden, "Akses tim admin ditolak", "FORBIDDEN")
		}
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), "INVALID_INPUT")
	}
	return response.OK(c, fiber.StatusCreated, "Admin berhasil ditambahkan", created)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	workspaceID := c.Get("X-Workspace-ID")
	if workspaceID == "" {
		workspaceID = c.Query("workspace_id")
	}
	adminID := c.Params("id")
	userID := auth.AuthenticatedUserID(c)
	var body WorkspaceAdmin
	if err := c.BodyParser(&body); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Format data tidak valid", "INVALID_JSON")
	}
	updated, err := h.service.Update(c.UserContext(), userID, workspaceID, adminID, body)
	if err != nil {
		if err == ErrForbidden {
			return response.Fail(c, fiber.StatusForbidden, "Akses tim admin ditolak", "FORBIDDEN")
		}
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), "INVALID_INPUT")
	}
	return response.OK(c, fiber.StatusOK, "Data admin berhasil diperbarui", updated)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	workspaceID := c.Get("X-Workspace-ID")
	if workspaceID == "" {
		workspaceID = c.Query("workspace_id")
	}
	adminID := c.Params("id")
	userID := auth.AuthenticatedUserID(c)
	if err := h.service.Delete(c.UserContext(), userID, workspaceID, adminID); err != nil {
		if err == ErrForbidden {
			return response.Fail(c, fiber.StatusForbidden, "Akses tim admin ditolak", "FORBIDDEN")
		}
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), "INVALID_INPUT")
	}
	return response.OK(c, fiber.StatusOK, "Admin berhasil dihapus", fiber.Map{"id": adminID})
}
