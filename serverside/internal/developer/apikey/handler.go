package apikey

import (
	"authbackend/internal/auth"
	"authbackend/pkg/response"
	"errors"
	"github.com/gofiber/fiber/v2"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service} }
func workspaceID(c *fiber.Ctx) (string, error) {
	v := c.Query("workspace_id")
	if v == "" {
		return "", response.Fail(c, 400, "workspace_id is required", "VALIDATION_ERROR")
	}
	return v, nil
}
func (h *Handler) List(c *fiber.Ctx) error {
	wid, e := workspaceID(c)
	if e != nil {
		return e
	}
	v, e := h.service.List(c.UserContext(), auth.AuthenticatedUserID(c), wid)
	if e != nil {
		return keyError(c, e)
	}
	return response.OK(c, 200, "API keys retrieved", v)
}
func (h *Handler) Create(c *fiber.Ctx) error {
	wid, e := workspaceID(c)
	if e != nil {
		return e
	}
	if !c.Is("json") {
		return response.Fail(c, 400, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	var b struct {
		Name   string   `json:"name"`
		Scopes []string `json:"scopes"`
	}
	if e = c.BodyParser(&b); e != nil {
		return response.Fail(c, 400, "Invalid JSON body", "INVALID_JSON")
	}
	v, e := h.service.CreateForUser(c.UserContext(), auth.AuthenticatedUserID(c), wid, b.Name, b.Scopes)
	if e != nil {
		return keyError(c, e)
	}
	return response.OK(c, 201, "API key created; secret is shown once", v)
}
func (h *Handler) Delete(c *fiber.Ctx) error {
	e := h.service.Revoke(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("id"))
	if e != nil {
		return keyError(c, e)
	}
	return response.OK(c, 200, "API key revoked", fiber.Map{"id": c.Params("id")})
}
func keyError(c *fiber.Ctx, e error) error {
	if errors.Is(e, ErrForbidden) {
		return response.Fail(c, 403, "You cannot perform this action", "FORBIDDEN")
	}
	if errors.Is(e, ErrInvalidKey) {
		return response.Fail(c, 400, "Invalid API key input", "VALIDATION_ERROR")
	}
	return response.Fail(c, 404, "API key or workspace not found", "API_KEY_NOT_FOUND")
}
