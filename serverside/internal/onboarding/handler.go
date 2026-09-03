package onboarding

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

func (h *Handler) Get(c *fiber.Ctx) error {
	workspaceID := c.Get("X-Workspace-ID")
	if workspaceID == "" {
		workspaceID = c.Query("workspace_id")
	}
	userID := auth.AuthenticatedUserID(c)
	profile, err := h.service.Get(c.UserContext(), userID, workspaceID)
	if err != nil {
		if err == ErrForbidden {
			return response.Fail(c, fiber.StatusForbidden, "Akses onboarding ditolak", "FORBIDDEN")
		}
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), "INVALID_INPUT")
	}
	return response.OK(c, fiber.StatusOK, "Onboarding profile retrieved", profile)
}

func (h *Handler) SaveProgress(c *fiber.Ctx) error {
	workspaceID := c.Get("X-Workspace-ID")
	if workspaceID == "" {
		workspaceID = c.Query("workspace_id")
	}
	userID := auth.AuthenticatedUserID(c)
	var body struct {
		CurrentStep int            `json:"current_step"`
		Data        OnboardingData `json:"data"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Format data tidak valid", "INVALID_JSON")
	}
	profile, err := h.service.SaveProgress(c.UserContext(), userID, workspaceID, body.CurrentStep, body.Data)
	if err != nil {
		if err == ErrForbidden {
			return response.Fail(c, fiber.StatusForbidden, "Akses onboarding ditolak", "FORBIDDEN")
		}
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), "INVALID_INPUT")
	}
	return response.OK(c, fiber.StatusOK, "Progres onboarding berhasil disimpan", profile)
}

func (h *Handler) Complete(c *fiber.Ctx) error {
	workspaceID := c.Get("X-Workspace-ID")
	if workspaceID == "" {
		workspaceID = c.Query("workspace_id")
	}
	userID := auth.AuthenticatedUserID(c)
	var body struct {
		Data OnboardingData `json:"data"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Format data tidak valid", "INVALID_JSON")
	}
	profile, err := h.service.Complete(c.UserContext(), userID, workspaceID, body.Data)
	if err != nil {
		if err == ErrForbidden {
			return response.Fail(c, fiber.StatusForbidden, "Akses onboarding ditolak", "FORBIDDEN")
		}
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), "INVALID_INPUT")
	}
	return response.OK(c, fiber.StatusOK, "Onboarding berhasil diselesaikan! ChatSolv siap digunakan.", profile)
}
