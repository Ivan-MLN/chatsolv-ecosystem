package internalapi

import (
	"errors"

	"authbackend/internal/conversation"
	"authbackend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service} }

func (h *Handler) SetHandoffs(handoffs HandoffCommandService) {
	h.service.SetHandoffs(handoffs)
}
func (h *Handler) ChannelStatus(c *fiber.Ctx) error {
	var in ChannelStatusInput
	if !c.Is("json") || c.BodyParser(&in) != nil {
		return response.Fail(c, 400, "Invalid JSON body", "INVALID_JSON")
	}
	if err := h.service.ChannelStatus(c.UserContext(), in); err != nil {
		return internalError(c, err)
	}
	return response.OK(c, 200, "Channel status updated", fiber.Map{"channel_id": in.ChannelID, "status": in.Status})
}
func (h *Handler) ChannelEvent(c *fiber.Ctx) error {
	var in ChannelEventInput
	if !c.Is("json") || c.BodyParser(&in) != nil {
		return response.Fail(c, 400, "Invalid JSON body", "INVALID_JSON")
	}
	if err := h.service.ChannelEvent(c.UserContext(), in); err != nil {
		return internalError(c, err)
	}
	return response.OK(c, 200, "Channel event processed", fiber.Map{"channel_id": in.ChannelID, "event": in.Event})
}
func (h *Handler) Incoming(c *fiber.Ctx) error {
	var in IncomingMessage
	if !c.Is("json") || c.BodyParser(&in) != nil {
		return response.Fail(c, 400, "Invalid JSON body", "INVALID_JSON")
	}
	result, err := h.service.Incoming(c.UserContext(), in)
	if err != nil {
		return internalError(c, err)
	}
	return response.OK(c, 200, "Message processed", result)
}
func (h *Handler) Respond(c *fiber.Ctx) error {
	var in RespondInput
	if !c.Is("json") || c.BodyParser(&in) != nil {
		return response.Fail(c, 400, "Invalid JSON body", "INVALID_JSON")
	}
	result, err := h.service.Respond(c.UserContext(), c.Params("agentID"), in)
	if err != nil {
		return internalError(c, err)
	}
	return response.OK(c, 200, "Agent response generated", result)
}
func (h *Handler) Health(c *fiber.Ctx) error {
	health, err := h.service.Health(c.UserContext(), c.Params("agentID"))
	if err != nil {
		return internalError(c, err)
	}
	return response.OK(c, 200, "Agent health retrieved", health)
}
func internalError(c *fiber.Ctx, err error) error {
	if errors.Is(err, ErrInvalidInput) {
		return response.Fail(c, 400, "Invalid internal request", "VALIDATION_ERROR")
	}
	if errors.Is(err, conversation.ErrMessageLimitReached) {
		return response.Fail(c, fiber.StatusTooManyRequests, "Monthly message limit reached", "MESSAGE_LIMIT_REACHED")
	}
	return response.Fail(c, 404, "Internal resource not found", "NOT_FOUND")
}
