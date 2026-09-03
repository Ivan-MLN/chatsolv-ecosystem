package dashboard

import (
	"net/http/httptest"
	"testing"
	"time"

	"authbackend/internal/auth"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type handlerRepository struct {
	fakeRepository
}

func TestDashboardHandlerRequiresWorkspaceID(t *testing.T) {
	app := fiber.New()
	service := NewService(&handlerRepository{})
	app.Get("/dashboard", setUser("user-a"), NewHandler(service).Overview)

	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/dashboard", nil))

	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, response.StatusCode)
}

func TestMeHandlerUsesAuthenticatedUser(t *testing.T) {
	repository := &handlerRepository{fakeRepository: fakeRepository{me: Me{User: User{ID: "user-a"}}}}
	app := fiber.New()
	app.Get("/me", setUser("user-a"), NewHandler(NewService(repository)).Me)

	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/me", nil))

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, response.StatusCode)
	require.Equal(t, "user-a", repository.gotUserID)
}

func setUser(userID string) fiber.Handler {
	manager := auth.NewJWTManager([]byte("01234567890123456789012345678901"), time.Minute)
	token, _, err := manager.Generate(userID)
	if err != nil {
		panic(err)
	}
	middleware := auth.RequireAccessToken(manager)
	return func(c *fiber.Ctx) error {
		c.Request().Header.Set(fiber.HeaderAuthorization, "Bearer "+token)
		return middleware(c)
	}
}
