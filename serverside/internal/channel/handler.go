package channel

import (
	"errors"

	"authbackend/internal/access"
	"authbackend/internal/auth"
	"authbackend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }
func (h *Handler) List(c *fiber.Ctx) error {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		return response.Fail(c, fiber.StatusBadRequest, "workspace_id is required", "VALIDATION_ERROR")
	}
	channels, err := h.service.List(c.UserContext(), auth.AuthenticatedUserID(c), workspaceID)
	if err != nil {
		return channelError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Channels retrieved", channels)
}
func (h *Handler) ConnectWhatsApp(c *fiber.Ctx) error {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		return response.Fail(c, fiber.StatusBadRequest, "workspace_id is required", "VALIDATION_ERROR")
	}
	if !isChannelJSON(c) {
		return response.Fail(c, fiber.StatusBadRequest, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	var body struct {
		DisplayName string `json:"display_name"`
		PhoneNumber string `json:"phone_number"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid JSON body", "INVALID_JSON")
	}
	resolved, _ := access.FromLocals(c)
	created, pairing, err := h.service.ConnectWhatsAppWithBypass(c.UserContext(), auth.AuthenticatedUserID(c), workspaceID, body.DisplayName, body.PhoneNumber, resolved.Entitlements.IsUnlimited)
	if err != nil {
		return channelError(c, err)
	}
	return response.OK(c, fiber.StatusAccepted, "WhatsApp pairing started", fiber.Map{
		"channel": created,
		"pairing": fiber.Map{
			"session_id":   pairing.SessionID,
			"status":       pairing.Status,
			"qr":           pairing.QR,
			"pairing_code": pairing.PairingCode,
		},
	})
}
func (h *Handler) GetProfile(c *fiber.Ctx) error {
	profile, err := h.service.GetProfile(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("id"))
	if err != nil {
		return channelError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "WhatsApp profile retrieved", profile)
}

func (h *Handler) Restart(c *fiber.Ctx) error {
	ch, err := h.service.Restart(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("id"))
	if err != nil {
		return channelError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "WhatsApp connection restarted", ch)
}

func (h *Handler) ToggleStatus(c *fiber.Ctx) error {
	var body struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil || (body.Status != "connected" && body.Status != "suspended") {
		return response.Fail(c, fiber.StatusBadRequest, "status must be connected or suspended", "VALIDATION_ERROR")
	}
	ch, err := h.service.SetStatus(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("id"), body.Status)
	if err != nil {
		return channelError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Channel status updated", ch)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	if err := h.service.Delete(c.UserContext(), auth.AuthenticatedUserID(c), c.Params("id")); err != nil {
		return channelError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Channel disconnected", fiber.Map{"id": c.Params("id"), "status": "disconnected"})
}
func channelError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return response.Fail(c, fiber.StatusBadRequest, "Invalid channel input", "VALIDATION_ERROR")
	case errors.Is(err, ErrForbidden):
		return response.Fail(c, fiber.StatusForbidden, "You cannot perform this action", "FORBIDDEN")
	case errors.Is(err, ErrNotFound):
		return response.Fail(c, fiber.StatusNotFound, "Channel or workspace not found", "CHANNEL_NOT_FOUND")
	case errors.Is(err, ErrQuotaExceeded):
		return response.Fail(c, fiber.StatusConflict, "Channel quota exceeded", "CHANNEL_QUOTA_EXCEEDED")
	case errors.Is(err, ErrConnectionFailed):
		return response.Fail(c, fiber.StatusBadGateway, "WhatsApp service could not start pairing", "WHATSAPP_CONNECTION_FAILED")
	default:
		return response.Fail(c, fiber.StatusInternalServerError, "An unexpected channel error occurred", "INTERNAL_SERVER_ERROR")
	}
}
func isChannelJSON(c *fiber.Ctx) bool { return c.Is("json") }
