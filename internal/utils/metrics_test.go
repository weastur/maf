package utils

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureMetrics(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	ConfigureMetrics(app)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected status code 200")
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/plain", "expected content type to be text/plain")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "maf_version")
}
