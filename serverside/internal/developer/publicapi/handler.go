package publicapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"authbackend/internal/conversation"
	"authbackend/internal/developer/apikey"
	"authbackend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) CreateSession(c *fiber.Ctx) error {
	apiKey, err := bearerToken(c)
	if err != nil {
		return err
	}
	if !c.Is("json") {
		return response.Fail(c, fiber.StatusBadRequest, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
	}
	var input CreateSessionInput
	if err = c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid JSON body", "INVALID_JSON")
	}
	created, err := h.service.CreateSession(c.UserContext(), apiKey, input)
	if err != nil {
		return publicError(c, err)
	}
	return response.OK(c, fiber.StatusCreated, "Agent session created", created)
}

func (h *Handler) SendMessage(c *fiber.Ctx) error {
	clientToken, err := bearerToken(c)
	if err != nil {
		return err
	}
	var body struct {
		Message string `json:"message"`
	}
	if !c.Is("json") || c.BodyParser(&body) != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid JSON body", "INVALID_JSON")
	}
	result, err := h.service.SendMessage(c.UserContext(), c.Params("id"), clientToken, body.Message)
	if err != nil {
		return publicError(c, err)
	}
	return response.OK(c, fiber.StatusOK, "Message processed", result)
}

func (h *Handler) StreamMessage(c *fiber.Ctx) error {
	clientToken, err := bearerToken(c)
	if err != nil {
		return err
	}
	var body struct {
		Message string `json:"message"`
	}
	if !c.Is("json") || c.BodyParser(&body) != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid JSON body", "INVALID_JSON")
	}
	result, err := h.service.SendMessage(c.UserContext(), c.Params("id"), clientToken, body.Message)
	if err != nil {
		return publicError(c, err)
	}
	payload, _ := json.Marshal(result)
	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Set(fiber.HeaderCacheControl, "no-cache")
	return c.SendString("event: message.start\ndata: {}\n\n" + "event: message.completed\ndata: " + string(payload) + "\n\n")
}

func bearerToken(c *fiber.Ctx) (string, error) {
	parts := strings.Fields(c.Get(fiber.HeaderAuthorization))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", response.Fail(c, fiber.StatusUnauthorized, "Bearer token is required", "UNAUTHORIZED")
	}
	return parts[1], nil
}

func publicError(c *fiber.Ctx, err error) error {
	if errors.Is(err, ErrInvalidInput) {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request", "VALIDATION_ERROR")
	}
	if errors.Is(err, ErrInvalidToken) || errors.Is(err, apikey.ErrInvalidKey) || errors.Is(err, apikey.ErrForbidden) {
		return response.Fail(c, fiber.StatusUnauthorized, "Invalid or expired token", "UNAUTHORIZED")
	}
	if errors.Is(err, conversation.ErrMessageLimitReached) {
		return response.Fail(c, fiber.StatusTooManyRequests, "Monthly message limit reached", "MESSAGE_LIMIT_REACHED")
	}
	return response.Fail(c, fiber.StatusServiceUnavailable, fmt.Sprintf("Agent is unavailable"), "AGENT_NOT_READY")
}
