package agentconfig

import (
	"authbackend/internal/auth"
	"authbackend/pkg/response"
	"errors"
	"mime"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service   *Service
	validator *validator.Validate
}

func NewHandler(service *Service) *Handler { return &Handler{service, validator.New()} }
func (h *Handler) GetPersonality(c *fiber.Ctx) error {
	if _, err := h.service.repository.Authorize(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("agentID")); err != nil {
		return configError(c, err)
	}
	p, err := h.service.repository.GetPersonality(c.UserContext(), c.Params("agentID"))
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, 200, "Personality retrieved", p)
}
func (h *Handler) UpdatePersonality(c *fiber.Ctx) error {
	media, _, err := mime.ParseMediaType(c.Get(fiber.HeaderContentType))
	if err != nil || media != fiber.MIMEApplicationJSON {
		return response.Fail(c, 400, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	var p Personality
	if err = c.BodyParser(&p); err != nil {
		return response.Fail(c, 400, "Invalid JSON body", "INVALID_JSON")
	}
	version, err := h.service.UpdatePersonality(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("agentID"), p)
	if err != nil {
		return configError(c, err)
	}
	return response.OK(c, 200, "Personality sync queued", fiber.Map{"config_version": version, "status": "syncing"})
}
func configError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return response.Fail(c, 400, "Invalid personality configuration", "VALIDATION_ERROR")
	case errors.Is(err, ErrForbidden):
		return response.Fail(c, 403, "You cannot perform this action", "FORBIDDEN")
	default:
		return response.Fail(c, 404, "Agent not found", "AGENT_NOT_FOUND")
	}
}
