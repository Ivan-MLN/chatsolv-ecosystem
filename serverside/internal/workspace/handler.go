package workspace

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

func NewHandler(service *Service) *Handler {
	return &Handler{service: service, validator: validator.New()}
}

type createRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=120"`
	Slug     string `json:"slug" validate:"required,min=2,max=63"`
	Timezone string `json:"timezone" validate:"required,max=64"`
}

type updateRequest struct {
	Name     string `json:"name" validate:"omitempty,min=2,max=120"`
	Timezone string `json:"timezone" validate:"omitempty,max=64"`
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var request createRequest
	if err := h.bind(c, &request); err != nil {
		return err
	}
	result, err := h.service.Create(c.UserContext(), auth.AuthenticatedUserID(c), CreateInput{Name: request.Name, Slug: request.Slug, Timezone: request.Timezone})
	if err != nil {
		return mapError(c, err)
	}
	return response.OK(c, fiber.StatusAccepted, "Workspace provisioning started", result)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	result, err := h.service.Get(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("workspaceID"))
	if err != nil {
		return mapError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Workspace retrieved", result)
}

func (h *Handler) CanonicalGet(c *fiber.Ctx) error {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		return response.Fail(c, fiber.StatusBadRequest, "workspace_id is required", "VALIDATION_ERROR")
	}
	result, err := h.service.Get(c.UserContext(), auth.AuthenticatedUserID(c), workspaceID)
	if err != nil {
		return mapError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Workspace retrieved", result)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	var request updateRequest
	if err := h.bind(c, &request); err != nil {
		return err
	}
	result, err := h.service.Update(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("workspaceID"), UpdateInput{Name: request.Name, Timezone: request.Timezone})
	if err != nil {
		return mapError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Workspace updated", result)
}

func (h *Handler) CanonicalUpdate(c *fiber.Ctx) error {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		return response.Fail(c, fiber.StatusBadRequest, "workspace_id is required", "VALIDATION_ERROR")
	}
	var request updateRequest
	if err := h.bind(c, &request); err != nil {
		return err
	}
	result, err := h.service.Update(c.UserContext(), auth.AuthenticatedUserID(c), workspaceID, UpdateInput{Name: request.Name, Timezone: request.Timezone})
	if err != nil {
		return mapError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Workspace updated", result)
}

func (h *Handler) Subscription(c *fiber.Ctx) error {
	result, err := h.service.Subscription(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("workspaceID"))
	if err != nil {
		return mapError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Subscription retrieved", result)
}

func (h *Handler) bind(c *fiber.Ctx, value any) error {
	mediaType, _, err := mime.ParseMediaType(c.Get(fiber.HeaderContentType))
	if err != nil || mediaType != fiber.MIMEApplicationJSON {
		return response.Fail(c, fiber.StatusBadRequest, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	if err := c.BodyParser(value); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid JSON body", "INVALID_JSON")
	}
	if err := h.validator.Struct(value); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Validation failed", "VALIDATION_ERROR")
	}
	return nil
}

func mapError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return response.Fail(c, fiber.StatusBadRequest, "Invalid workspace input", "VALIDATION_ERROR")
	case errors.Is(err, ErrForbidden):
		return response.Fail(c, fiber.StatusForbidden, "You cannot perform this action", "FORBIDDEN")
	case errors.Is(err, ErrNotFound):
		return response.Fail(c, fiber.StatusNotFound, "Workspace not found", "WORKSPACE_NOT_FOUND")
	case errors.Is(err, ErrSlugExists):
		return response.Fail(c, fiber.StatusConflict, "Workspace slug already exists", "WORKSPACE_SLUG_EXISTS")
	default:
		return response.Fail(c, fiber.StatusInternalServerError, "Something went wrong", "INTERNAL_ERROR")
	}
}
