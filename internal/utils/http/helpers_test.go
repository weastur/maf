package http

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiUtils "github.com/weastur/maf/internal/utils/http/api"
)

type mockHealthchecker struct{}

func (m *mockHealthchecker) IsLive(_ *fiber.Ctx) bool {
	return true
}

func (m *mockHealthchecker) IsReady(_ *fiber.Ctx) bool {
	return true
}

type mockAPIInstance struct{}

func (m *mockAPIInstance) ErrorHandler(c *fiber.Ctx, _ error) error {
	return c.Status(fiber.StatusBadRequest).SendString("Handled by API instance")
}

func TestAPIGroup(t *testing.T) {
	app := fiber.New()
	apiGroup := APIGroup(app)

	apiGroup.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAPIVersionGroup(t *testing.T) {
	app := fiber.New()
	apiGroup := APIGroup(app)
	versionGroup := APIVersionGroup(apiGroup, "v1")

	versionGroup.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "v1", resp.Header.Get("X-Api-Version"))
}

func TestAttachGenericMiddlewares(t *testing.T) {
	app := fiber.New()
	logger := zerolog.Nop()
	healthchecker := &mockHealthchecker{}

	AttachGenericMiddlewares(app, logger, healthchecker)

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get("X-Request-ID"))
	assert.NotEmpty(t, resp.Header.Get("X-Ratelimit-Limit"))
	assert.NotEmpty(t, resp.Header.Get("X-Ratelimit-Remaining"))
	assert.NotEmpty(t, resp.Header.Get("X-Ratelimit-Reset"))

	reqLivez, _ := http.NewRequest(http.MethodGet, "/livez", nil)
	respLivez, err := app.Test(reqLivez)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, respLivez.StatusCode)

	reqReadyz, _ := http.NewRequest(http.MethodGet, "/readyz", nil)
	respReadyz, err := app.Test(reqReadyz)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, respReadyz.StatusCode)
}

func TestDefaultErrorHandler(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler,
	})

	app.Get("/error", func(_ *fiber.Ctx) error {
		return fiber.ErrBadRequest
	})

	req, _ := http.NewRequest(http.MethodGet, "/error", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestErrorHandlerWithAPIInstance(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler,
	})

	app.Use(func(c *fiber.Ctx) error {
		ctx := context.WithValue(t.Context(), apiUtils.APIInstanceContextKey, &mockAPIInstance{})
		c.SetUserContext(ctx)

		return c.Next()
	})

	app.Get("/error", func(_ *fiber.Ctx) error {
		return fiber.ErrBadRequest
	})

	req, _ := http.NewRequest(http.MethodGet, "/error", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "Handled by API instance")
}
