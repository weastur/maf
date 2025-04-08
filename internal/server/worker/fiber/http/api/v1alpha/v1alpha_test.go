package v1alpha

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIV1Alpha_Init(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	logger := zerolog.Nop()
	mockConsensus := new(MockConsensus)

	api := &APIV1Alpha{
		version:   "v1alpha",
		prefix:    "/v1alpha",
		validator: new(MockValidator),
	}
	api.Init(app.Group("/api"), logger, mockConsensus)

	t.Run("Swagger Docs Endpoint", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequest(fiber.MethodGet, "/api/v1alpha/docs", nil)
		resp, _ := app.Test(req, -1)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "MySQL auto failover server API")
	})

	t.Run("Version Endpoint", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequest(fiber.MethodGet, "/api/v1alpha/version", nil)
		resp, _ := app.Test(req, -1)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), `"version"`)
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequest(fiber.MethodGet, "/api/v1alpha/protected", nil)
		resp, _ := app.Test(req, -1)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)

		var response map[string]any
		err := json.Unmarshal(body, &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
		assert.NotEmpty(t, response["error"])
	})

	t.Run("Invalid Endpoint", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequest(fiber.MethodGet, "/api/v1alpha/invalid", nil)
		req.Header.Set("X-Auth-Token", "root")
		resp, _ := app.Test(req, -1)

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Cannot GET /api/v1alpha/invalid")
	})
}

func TestAPIV1Alpha_Get(t *testing.T) {
	t.Parallel()

	t.Run("Singleton Instance", func(t *testing.T) {
		t.Parallel()

		instance1 := Get()
		instance2 := Get()

		assert.NotNil(t, instance1)
		assert.NotNil(t, instance2)
		assert.Equal(t, instance1, instance2, "Get() should return the same instance")
	})

	t.Run("Instance Properties", func(t *testing.T) {
		t.Parallel()

		instance := Get()

		assert.Equal(t, "v1alpha", instance.version, "Version should be 'v1alpha'")
		assert.Equal(t, "/v1alpha", instance.prefix, "Prefix should be '/v1alpha'")
	})
}

func TestAPIV1Alpha_ErrorHandler(t *testing.T) {
	t.Parallel()

	t.Run("Custom Error Handling", func(t *testing.T) {
		t.Parallel()

		api := &APIV1Alpha{
			version: "v1alpha",
			prefix:  "/v1alpha",
		}

		app := fiber.New(fiber.Config{
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				return api.ErrorHandler(c, err)
			},
		})

		app.Get("/error", func(_ *fiber.Ctx) error {
			return errors.New("custom error message")
		})

		req, _ := http.NewRequest(fiber.MethodGet, "/error", nil)
		resp, _ := app.Test(req, -1)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)

		var response map[string]any
		err := json.Unmarshal(body, &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
		assert.Equal(t, "custom error message", response["error"])
	})
}
