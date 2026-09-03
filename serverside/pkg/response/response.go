package response

import "github.com/gofiber/fiber/v2"

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type Envelope struct {
	Data      any        `json:"data,omitempty"`
	Meta      any        `json:"meta,omitempty"`
	Error     *ErrorBody `json:"error,omitempty"`
	RequestID string     `json:"request_id,omitempty"`
}

func OK(c *fiber.Ctx, status int, msg string, data any) error {
	return c.Status(status).JSON(Envelope{Data: data, Meta: fiber.Map{"message": msg}, RequestID: requestID(c)})
}
func Fail(c *fiber.Ctx, status int, msg, code string) error {
	return c.Status(status).JSON(Envelope{Error: &ErrorBody{Code: code, Message: msg}, RequestID: requestID(c)})
}

func requestID(c *fiber.Ctx) string {
	value, _ := c.Locals("request_id").(string)
	return value
}
