package v1alpha

import (
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weastur/maf/internal/utils"
)

func TestAuthFilter(t *testing.T) {
	expectedVersion := utils.AppVersion()

	app := fiber.New()
	app.Use(AuthMiddleware())
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})
	app.Get("/version", VersionHandler)

	tests := []struct {
		url          string
		expectedResp string
	}{
		{url: "/version", expectedResp: `{"status":"success","error":"","data":{"version":"` + expectedVersion + `"}}`},
		{url: "/protected", expectedResp: `{"status":"error","error":"missing or malformed API Key","data":{}}`},
	}

	for _, test := range tests {
		req, _ := http.NewRequest(http.MethodGet, test.url, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.JSONEq(t, test.expectedResp, string(body))
	}
}

func TestApiKey(t *testing.T) {
	app := fiber.New()
	app.Use(AuthMiddleware())
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	tests := []struct {
		name         string
		apiKey       string
		expectedCode int
		expectedResp string
	}{
		{
			name:         "Valid API Key",
			apiKey:       "root",
			expectedCode: http.StatusOK,
			expectedResp: "",
		},
		{
			name:         "Invalid API Key",
			apiKey:       "invalid",
			expectedCode: http.StatusOK,
			expectedResp: `{"status":"error","error":"missing or malformed API Key","data":{}}`,
		},
		{
			name:         "Missing API Key",
			apiKey:       "",
			expectedCode: http.StatusOK,
			expectedResp: `{"status":"error","error":"missing or malformed API Key","data":{}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
			if test.apiKey != "" {
				req.Header.Set("X-Auth-Token", test.apiKey)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, test.expectedCode, resp.StatusCode)

			if test.expectedResp != "" {
				assert.JSONEq(t, test.expectedResp, string(body))
			}
		})
	}
}
