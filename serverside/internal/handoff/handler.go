package handoff

import (
	"authbackend/internal/auth"
	"authbackend/pkg/response"
	"strconv"

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
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	list, err := h.service.List(c.UserContext(), userID, workspaceID, limit)
	if err != nil {
		if err == ErrForbidden {
			return response.Fail(c, fiber.StatusForbidden, "Akses handoff ditolak", "FORBIDDEN")
		}
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), "INVALID_INPUT")
	}
	return response.OK(c, fiber.StatusOK, "Daftar handoff berhasil diambil", list)
}

func (h *Handler) ListEvents(c *fiber.Ctx) error {
	workspaceID := c.Get("X-Workspace-ID")
	if workspaceID == "" {
		workspaceID = c.Query("workspace_id")
	}
	conversationID := c.Params("id")
	userID := auth.AuthenticatedUserID(c)
	events, err := h.service.ListEvents(c.UserContext(), userID, workspaceID, conversationID)
	if err != nil {
		if err == ErrForbidden {
			return response.Fail(c, fiber.StatusForbidden, "Akses riwayat aktivitas ditolak", "FORBIDDEN")
		}
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), "INVALID_INPUT")
	}
	return response.OK(c, fiber.StatusOK, "Riwayat event percakapan berhasil diambil", events)
}

func (h *Handler) Takeover(c *fiber.Ctx) error {
	workspaceID := c.Get("X-Workspace-ID")
	if workspaceID == "" {
		workspaceID = c.Query("workspace_id")
	}
	conversationID := c.Params("id")
	userID := auth.AuthenticatedUserID(c)
	if err := h.service.ManualDashboardTakeover(c.UserContext(), userID, workspaceID, conversationID); err != nil {
		if err == ErrForbidden {
			return response.Fail(c, fiber.StatusForbidden, "Akses ambil alih ditolak", "FORBIDDEN")
		}
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), "INVALID_INPUT")
	}
	return response.OK(c, fiber.StatusOK, "Percakapan berhasil diambil alih oleh Anda", fiber.Map{
		"conversation_id": conversationID,
		"mode":            "human",
	})
}

func (h *Handler) ReturnToAI(c *fiber.Ctx) error {
	workspaceID := c.Get("X-Workspace-ID")
	if workspaceID == "" {
		workspaceID = c.Query("workspace_id")
	}
	conversationID := c.Params("id")
	userID := auth.AuthenticatedUserID(c)
	if err := h.service.ReturnToAI(c.UserContext(), userID, workspaceID, conversationID); err != nil {
		if err == ErrForbidden {
			return response.Fail(c, fiber.StatusForbidden, "Akses pengembalian ke AI ditolak", "FORBIDDEN")
		}
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), "INVALID_INPUT")
	}
	return response.OK(c, fiber.StatusOK, "Percakapan telah dikembalikan ke ChatSolv", fiber.Map{
		"conversation_id": conversationID,
		"mode":            "agent",
	})
}
