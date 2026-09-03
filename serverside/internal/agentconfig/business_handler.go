package agentconfig

import (
	"authbackend/internal/auth"
	"authbackend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type CanonicalBusinessHandler struct{ service *SettingsService }

func NewCanonicalBusinessHandler(service *SettingsService) *CanonicalBusinessHandler {
	return &CanonicalBusinessHandler{service: service}
}
func (h *CanonicalBusinessHandler) Get(c *fiber.Ctx) error {
	workspaceID, err := canonicalWorkspaceID(c)
	if err != nil {
		return err
	}
	if _, err = h.service.repository.AuthorizeWorkspace(c.UserContext(), auth.AuthenticatedUserID(c), workspaceID); err != nil {
		return configError(c, err)
	}
	value, err := h.service.repository.GetBusinessProfile(c.UserContext(), workspaceID)
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Business profile retrieved", value)
}
func (h *CanonicalBusinessHandler) Update(c *fiber.Ctx) error {
	workspaceID, err := canonicalWorkspaceID(c)
	if err != nil {
		return err
	}
	if !isJSON(c) {
		return response.Fail(c, fiber.StatusBadRequest, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	var value BusinessProfile
	if err = c.BodyParser(&value); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid JSON body", "INVALID_JSON")
	}
	version, err := h.service.UpdateBusinessProfile(c.UserContext(), auth.AuthenticatedUserID(c), workspaceID, value)
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Business profile sync queued", fiber.Map{"config_version": version, "status": "syncing"})
}
