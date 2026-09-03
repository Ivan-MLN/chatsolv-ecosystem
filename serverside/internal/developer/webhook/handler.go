package webhook

import (
	"authbackend/internal/auth"
	"authbackend/pkg/response"
	"errors"
	"github.com/gofiber/fiber/v2"
)

type Handler struct{ service *Service }

func NewHandler(s *Service) *Handler { return &Handler{s} }
func webhookWorkspaceID(c *fiber.Ctx) (string, error) {
	v := c.Query("workspace_id")
	if v == "" {
		return "", response.Fail(c, 400, "workspace_id is required", "VALIDATION_ERROR")
	}
	return v, nil
}
func (h *Handler) List(c *fiber.Ctx) error {
	wid, e := webhookWorkspaceID(c)
	if e != nil {
		return e
	}
	v, e := h.service.List(c.UserContext(), auth.AuthenticatedUserID(c), wid)
	if e != nil {
		return webhookError(c, e)
	}
	return response.OK(c, 200, "Webhooks retrieved", v)
}
func (h *Handler) Create(c *fiber.Ctx) error {
	wid, e := webhookWorkspaceID(c)
	if e != nil {
		return e
	}
	if !c.Is("json") {
		return response.Fail(c, 400, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	var in CreateInput
	if e = c.BodyParser(&in); e != nil {
		return response.Fail(c, 400, "Invalid JSON body", "INVALID_JSON")
	}
	v, e := h.service.Create(c.UserContext(), auth.AuthenticatedUserID(c), wid, in)
	if e != nil {
		return webhookError(c, e)
	}
	return response.OK(c, 201, "Webhook created; signing secret is shown once", v)
}
func (h *Handler) Update(c *fiber.Ctx) error {
	if !c.Is("json") {
		return response.Fail(c, 400, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	var in UpdateInput
	if e := c.BodyParser(&in); e != nil {
		return response.Fail(c, 400, "Invalid JSON body", "INVALID_JSON")
	}
	v, e := h.service.Update(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("id"), in)
	if e != nil {
		return webhookError(c, e)
	}
	return response.OK(c, 200, "Webhook updated", v)
}
func (h *Handler) Delete(c *fiber.Ctx) error {
	if e := h.service.Delete(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("id")); e != nil {
		return webhookError(c, e)
	}
	return response.OK(c, 200, "Webhook deleted", fiber.Map{"id": c.Params("id")})
}
func webhookError(c *fiber.Ctx, e error) error {
	if errors.Is(e, ErrForbidden) {
		return response.Fail(c, 403, "You cannot perform this action", "FORBIDDEN")
	}
	if errors.Is(e, ErrEntitlementRequired) {
		return response.Fail(c, 403, "Webhook entitlement is required", "SUBSCRIPTION_REQUIRED")
	}
	if errors.Is(e, ErrInvalidInput) {
		return response.Fail(c, 400, "Invalid webhook input", "VALIDATION_ERROR")
	}
	return response.Fail(c, 404, "Webhook or workspace not found", "WEBHOOK_NOT_FOUND")
}
