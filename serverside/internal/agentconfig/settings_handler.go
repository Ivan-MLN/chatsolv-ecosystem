package agentconfig

import (
	"authbackend/internal/auth"
	"authbackend/pkg/response"
	"errors"
	"mime"

	"github.com/gofiber/fiber/v2"
)

type SettingsHandler struct{ service *SettingsService }

func NewSettingsHandler(service *SettingsService) *SettingsHandler {
	return &SettingsHandler{service: service}
}

func (h *SettingsHandler) GetAgentProfile(c *fiber.Ctx) error {
	userID, agentID := auth.AuthenticatedUserID(c), c.Params("agentID")
	if _, err := h.service.repository.Authorize(c.UserContext(), userID, agentID); err != nil {
		return configError(c, err)
	}
	value, err := h.service.repository.GetAgentProfile(c.UserContext(), agentID)
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Agent profile retrieved", value)
}
func (h *SettingsHandler) UpdateAgentProfile(c *fiber.Ctx) error {
	if !isJSON(c) {
		return response.Fail(c, fiber.StatusBadRequest, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	var value AgentProfile
	if err := c.BodyParser(&value); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid JSON body", "INVALID_JSON")
	}
	version, err := h.service.UpdateAgentProfile(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("agentID"), value)
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Agent profile sync queued", fiber.Map{"config_version": version, "status": "syncing"})
}
func (h *SettingsHandler) GetBusiness(c *fiber.Ctx) error {
	userID, workspaceID := auth.AuthenticatedUserID(c), c.Params("workspaceID")
	if _, err := h.service.repository.AuthorizeWorkspace(c.UserContext(), userID, workspaceID); err != nil {
		return configError(c, err)
	}
	value, err := h.service.repository.GetBusinessProfile(c.UserContext(), workspaceID)
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Business profile retrieved", value)
}
func (h *SettingsHandler) UpdateBusiness(c *fiber.Ctx) error {
	if !isJSON(c) {
		return response.Fail(c, fiber.StatusBadRequest, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	var value BusinessProfile
	if err := c.BodyParser(&value); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid JSON body", "INVALID_JSON")
	}
	version, err := h.service.UpdateBusinessProfile(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("workspaceID"), value)
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Business profile sync queued", fiber.Map{"config_version": version, "status": "syncing"})
}
func (h *SettingsHandler) GetPolicies(c *fiber.Ctx) error {
	userID, workspaceID := auth.AuthenticatedUserID(c), c.Params("workspaceID")
	if _, err := h.service.repository.AuthorizeWorkspace(c.UserContext(), userID, workspaceID); err != nil {
		return configError(c, err)
	}
	value, err := h.service.repository.GetBusinessPolicies(c.UserContext(), workspaceID)
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Business policies retrieved", value)
}
func (h *SettingsHandler) UpdatePolicies(c *fiber.Ctx) error {
	if !isJSON(c) {
		return response.Fail(c, fiber.StatusBadRequest, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	var value BusinessPolicies
	if err := c.BodyParser(&value); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid JSON body", "INVALID_JSON")
	}
	version, err := h.service.UpdateBusinessPolicies(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("workspaceID"), value)
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Business policy sync queued", fiber.Map{"config_version": version, "status": "syncing"})
}
func isJSON(c *fiber.Ctx) bool {
	media, _, err := mime.ParseMediaType(c.Get(fiber.HeaderContentType))
	return err == nil && media == fiber.MIMEApplicationJSON
}

var _ = errors.Is
