package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

type MockListener struct {
	mock.Mock
}

func (m *MockListener) Listen(addr string) error {
	args := m.Called(addr)

	return args.Error(0)
}

func (m *MockListener) ListenTLS(addr, certFile, keyFile string) error {
	args := m.Called(addr, certFile, keyFile)

	return args.Error(0)
}

func (m *MockListener) ListenMutualTLS(addr, certFile, keyFile, clientCertFile string) error {
	args := m.Called(addr, certFile, keyFile, clientCertFile)

	return args.Error(0)
}

func TestListen(t *testing.T) {
	t.Parallel()

	logger := zerolog.Nop()

	errMutualTLS := errors.New("mutual TLS error")
	errTLS := errors.New("TLS error")
	errListen := errors.New("listen error")

	tests := []struct {
		name           string
		addr           string
		certFile       string
		keyFile        string
		clientCertFile string
		setupMock      func(m *MockListener)
		expectedError  error
	}{
		{
			name:           "Listen with mutual TLS",
			addr:           "localhost:8080",
			certFile:       "cert.pem",
			keyFile:        "key.pem",
			clientCertFile: "clientCert.pem",
			setupMock: func(m *MockListener) {
				m.On("ListenMutualTLS", "localhost:8080", "cert.pem", "key.pem", "clientCert.pem").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:           "Listen with TLS",
			addr:           "localhost:8080",
			certFile:       "cert.pem",
			keyFile:        "key.pem",
			clientCertFile: "",
			setupMock: func(m *MockListener) {
				m.On("ListenTLS", "localhost:8080", "cert.pem", "key.pem").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:           "Listen without TLS",
			addr:           "localhost:8080",
			certFile:       "",
			keyFile:        "",
			clientCertFile: "",
			setupMock: func(m *MockListener) {
				m.On("Listen", "localhost:8080").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:           "Error in ListenMutualTLS",
			addr:           "localhost:8080",
			certFile:       "cert.pem",
			keyFile:        "key.pem",
			clientCertFile: "clientCert.pem",
			setupMock: func(m *MockListener) {
				m.On("ListenMutualTLS", "localhost:8080", "cert.pem", "key.pem", "clientCert.pem").Return(errMutualTLS)
			},
			expectedError: errMutualTLS,
		},
		{
			name:           "Error in ListenTLS",
			addr:           "localhost:8080",
			certFile:       "cert.pem",
			keyFile:        "key.pem",
			clientCertFile: "",
			setupMock: func(m *MockListener) {
				m.On("ListenTLS", "localhost:8080", "cert.pem", "key.pem").Return(errTLS)
			},
			expectedError: errTLS,
		},
		{
			name:           "Error in Listen",
			addr:           "localhost:8080",
			certFile:       "",
			keyFile:        "",
			clientCertFile: "",
			setupMock: func(m *MockListener) {
				m.On("Listen", "localhost:8080").Return(errListen)
			},
			expectedError: errListen,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockListener := new(MockListener)
			tt.setupMock(mockListener)

			err := Listen(mockListener, logger, tt.addr, tt.certFile, tt.keyFile, tt.clientCertFile)

			require.ErrorIs(t, err, tt.expectedError)
			mockListener.AssertExpectations(t)
		})
	}
}

func TestAPIGroup(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	assert.Equal(t, "Handled by API instance", string(body))
}
