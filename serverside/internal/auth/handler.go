package auth

import (
	"authbackend/pkg/response"
	"errors"
	"mime"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	s *Service
	v *validator.Validate
}

func NewHandler(s *Service) *Handler { return &Handler{s, validator.New()} }

var errResponseWritten = errors.New("response already written")

type registerRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=254"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,max=128"`
}

type forgotRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type resetRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (h *Handler) bind(c *fiber.Ctx, v any) error {
	mediaType, _, err := mime.ParseMediaType(c.Get(fiber.HeaderContentType))
	if err != nil || mediaType != "application/json" {
		_ = response.Fail(c, 400, "Content-Type must be application/json", "INVALID_CONTENT_TYPE")
		return errResponseWritten
	}
	if e := c.BodyParser(v); e != nil {
		_ = response.Fail(c, 400, "Invalid JSON body", "INVALID_JSON")
		return errResponseWritten
	}
	if e := h.v.Struct(v); e != nil {
		_ = response.Fail(c, 400, "Validation failed", "VALIDATION_ERROR")
		return errResponseWritten
	}
	return nil
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var r registerRequest
	if e := h.bind(c, &r); e != nil {
		return nil
	}
	// Normal registration is ALWAYS assigned platform_role = "user" by backend
	u, e := h.s.Register(c.UserContext(), RegisterInput{r.Name, r.Email, r.Password})
	if e != nil {
		return mapError(c, e)
	}
	return response.OK(c, 201, "Registration successful", fiber.Map{
		"id":            u.ID,
		"name":          u.Name,
		"email":         u.Email,
		"platform_role": "user",
	})
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var r loginRequest
	if e := h.bind(c, &r); e != nil {
		return nil
	}
	v, e := h.s.Login(c.UserContext(), LoginInput{r.Email, r.Password})
	if e != nil {
		return mapError(c, e)
	}
	return response.OK(c, 200, "Login successful", fiber.Map{
		"access_token":  v.AccessToken,
		"refresh_token": v.RefreshToken,
		"token_type":    v.TokenType,
		"expires_in":    v.ExpiresIn,
	})
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	var r refreshRequest
	if e := h.bind(c, &r); e != nil {
		return nil
	}
	v, e := h.s.Refresh(c.UserContext(), r.RefreshToken)
	if e != nil {
		return mapError(c, e)
	}
	return response.OK(c, 200, "Token refreshed", fiber.Map{
		"access_token":  v.AccessToken,
		"refresh_token": v.RefreshToken,
		"token_type":    v.TokenType,
		"expires_in":    v.ExpiresIn,
	})
}

func (h *Handler) Forgot(c *fiber.Ctx) error {
	var r forgotRequest
	if e := h.bind(c, &r); e != nil {
		return nil
	}
	if e := h.s.ForgotPassword(c.UserContext(), r.Email); e != nil {
		return mapError(c, e)
	}
	return response.OK(c, 200, "If the account exists, password reset instructions have been sent", fiber.Map{})
}

func (h *Handler) Reset(c *fiber.Ctx) error {
	var r resetRequest
	if e := h.bind(c, &r); e != nil {
		return nil
	}
	if e := h.s.ResetPassword(c.UserContext(), r.Token, r.NewPassword); e != nil {
		return mapError(c, e)
	}
	return response.OK(c, 200, "Password reset successful", fiber.Map{})
}

func mapError(c *fiber.Ctx, e error) error {
	switch {
	case errors.Is(e, ErrUserExists):
		return response.Fail(c, 409, "Email already registered", "USER_ALREADY_EXISTS")
	case errors.Is(e, ErrInvalidCredentials):
		return response.Fail(c, 401, "Invalid credentials", "INVALID_CREDENTIALS")
	case errors.Is(e, ErrInvalidRefreshToken):
		return response.Fail(c, 401, "Invalid or expired refresh token", "INVALID_REFRESH_TOKEN")
	case errors.Is(e, ErrInvalidResetToken):
		return response.Fail(c, 400, "Invalid or expired reset token", "INVALID_RESET_TOKEN")
	default:
		return response.Fail(c, 500, "Internal server error", "INTERNAL_ERROR")
	}
}
