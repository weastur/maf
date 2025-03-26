package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureMetrics(t *testing.T) {
	app := fiber.New()
	ConfigureMetrics(app)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "maf_version")
}
