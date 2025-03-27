package v1alpha

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weastur/maf/internal/utils"
)

func TestVersionHandler(t *testing.T) {
	expectedVersion := utils.AppVersion()

	app := fiber.New()
	app.Get("/version", VersionHandler)

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.JSONEq(t, `{"status":"success","error":"","data":{"version":"`+expectedVersion+`"}}`, string(body))
}

func TestErrorHandler(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler,
	})

	app.Get("/error", func(_ *fiber.Ctx) error {
		return errors.New("test error")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.JSONEq(t, `{"status":"error","error":"test error","data":{}}`, string(body))
}
