package agentconfig

import (
	"authbackend/internal/auth"
	"authbackend/pkg/response"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type CanonicalHandler struct{ service *CanonicalService }

func NewCanonicalHandler(service *CanonicalService) *CanonicalHandler {
	return &CanonicalHandler{service: service}
}

func canonicalWorkspaceID(c *fiber.Ctx) (string, error) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		return "", response.Fail(c, fiber.StatusBadRequest, "workspace_id is required", "VALIDATION_ERROR")
	}
	return workspaceID, nil
}
func (h *CanonicalHandler) Get(c *fiber.Ctx) error {
	workspaceID, err := canonicalWorkspaceID(c)
	if err != nil {
		return err
	}
	value, err := h.service.Get(c.UserContext(), auth.AuthenticatedUserID(c), workspaceID)
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Agent retrieved", value)
}
func (h *CanonicalHandler) Update(c *fiber.Ctx) error {
	workspaceID, err := canonicalWorkspaceID(c)
	if err != nil {
		return err
	}
	if !isJSON(c) {
		return response.Fail(c, fiber.StatusBadRequest, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	var body struct {
		Name string `json:"name"`
	}
	if err = c.BodyParser(&body); err != nil {
		return response.Fail(c, 400, "Invalid JSON body", "INVALID_JSON")
	}
	value, err := h.service.Update(c.UserContext(), auth.AuthenticatedUserID(c), workspaceID, body.Name)
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Agent updated", value)
}
func (h *CanonicalHandler) GetProfile(c *fiber.Ctx) error {
	workspaceID, err := canonicalWorkspaceID(c)
	if err != nil {
		return err
	}
	value, err := h.service.GetProfile(c.UserContext(), auth.AuthenticatedUserID(c), workspaceID)
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, 200, "Agent profile retrieved", value)
}
func (h *CanonicalHandler) UpdateProfile(c *fiber.Ctx) error {
	workspaceID, err := canonicalWorkspaceID(c)
	if err != nil {
		return err
	}
	if !isJSON(c) {
		return response.Fail(c, 400, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	var body AgentProfile
	if err = c.BodyParser(&body); err != nil {
		return response.Fail(c, 400, "Invalid JSON body", "INVALID_JSON")
	}
	version, err := h.service.UpdateProfile(c.UserContext(), auth.AuthenticatedUserID(c), workspaceID, body)
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, 200, "Agent profile sync queued", fiber.Map{"config_version": version, "status": "syncing"})
}
func (h *CanonicalHandler) GetPersonality(c *fiber.Ctx) error {
	workspaceID, err := canonicalWorkspaceID(c)
	if err != nil {
		return err
	}
	value, err := h.service.GetPersonality(c.UserContext(), auth.AuthenticatedUserID(c), workspaceID)
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, 200, "Personality retrieved", value)
}
func (h *CanonicalHandler) UpdatePersonality(c *fiber.Ctx) error {
	workspaceID, err := canonicalWorkspaceID(c)
	if err != nil {
		return err
	}
	if !isJSON(c) {
		return response.Fail(c, 400, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	var body Personality
	if err = c.BodyParser(&body); err != nil {
		return response.Fail(c, 400, "Invalid JSON body", "INVALID_JSON")
	}
	version, err := h.service.UpdatePersonality(c.UserContext(), auth.AuthenticatedUserID(c), workspaceID, body)
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, 200, "Personality sync queued", fiber.Map{"config_version": version, "status": "syncing"})
}
func (h *CanonicalHandler) Test(c *fiber.Ctx) error {
	workspaceID, err := canonicalWorkspaceID(c)
	if err != nil {
		return err
	}
	if !isJSON(c) {
		return response.Fail(c, 400, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	var body struct {
		Message        string `json:"message"`
		ConversationID string `json:"conversation_id"`
		Reset          bool   `json:"reset"`
	}
	if err = c.BodyParser(&body); err != nil {
		return response.Fail(c, 400, "Invalid JSON body", "INVALID_JSON")
	}
	value, err := h.service.Test(c.UserContext(), auth.AuthenticatedUserID(c), workspaceID, body.Message, body.ConversationID, body.Reset)
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, 200, "Agent test completed", value)
}
func (h *CanonicalHandler) GenerateSetup(c *fiber.Ctx) error {
	workspaceID, err := canonicalWorkspaceID(c)
	if err != nil {
		return err
	}
	if !isJSON(c) {
		return response.Fail(c, 400, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	var body struct {
		Description string `json:"description"`
	}
	if err = c.BodyParser(&body); err != nil || strings.TrimSpace(body.Description) == "" {
		return response.Fail(c, 400, "Description is required", "VALIDATION_ERROR")
	}
	result, err := h.service.GenerateSetup(c.UserContext(), auth.AuthenticatedUserID(c), workspaceID, body.Description)
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, 200, "AI Agent configuration generated", result)
}
