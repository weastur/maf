package v1alpha

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/weastur/maf/internal/server/worker/raft"
	apiUtils "github.com/weastur/maf/internal/utils/http/api"
)

type MockConsensus struct {
	mock.Mock
}

func (m *MockConsensus) IsLeader() bool {
	args := m.Called()

	return args.Bool(0)
}

func (m *MockConsensus) Join(serverID, addr string) error {
	args := m.Called(serverID, addr)

	return args.Error(0)
}

func (m *MockConsensus) Forget(serverID string) error {
	args := m.Called(serverID)

	return args.Error(0)
}

func (m *MockConsensus) GetInfo(verbose bool) (*raft.Info, error) {
	args := m.Called(verbose)

	return args.Get(0).(*raft.Info), args.Error(1)
}

func (m *MockConsensus) Get(key string) (string, bool) {
	args := m.Called(key)

	return args.String(0), args.Bool(1)
}

func (m *MockConsensus) Set(key, value string) error {
	args := m.Called(key, value)

	return args.Error(0)
}

func (m *MockConsensus) Delete(key string) error {
	args := m.Called(key)

	return args.Error(0)
}

type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) Validate(data any) error {
	args := m.Called(data)

	return args.Error(0)
}

func TestUnpackCtx(t *testing.T) {
	t.Parallel()

	logger := zerolog.Nop()
	mockConsensus := new(MockConsensus)
	mockAPI := new(APIV1Alpha)
	requestID := "test-request-id"

	tests := []struct {
		name           string
		setupContext   func(c *fiber.Ctx)
		expectedLogger zerolog.Logger
		expectedCo     Consensus
		expectedAPI    *APIV1Alpha
		expectedRID    string
	}{
		{
			name: "Valid context with all values",
			setupContext: func(c *fiber.Ctx) {
				ctx := c.UserContext()
				ctx = context.WithValue(ctx, consensusInstanceContextKey, mockConsensus)
				ctx = context.WithValue(ctx, apiUtils.APIInstanceContextKey, mockAPI)
				ctx = context.WithValue(ctx, apiUtils.RequestIDContextKey, requestID)
				c.SetUserContext(ctx)
			},
			expectedLogger: logger.With().Str(requestIDLogField, requestID).Logger(),
			expectedCo:     mockConsensus,
			expectedAPI:    mockAPI,
			expectedRID:    requestID,
		},
		{
			name: "Context missing consensus",
			setupContext: func(c *fiber.Ctx) {
				ctx := c.UserContext()
				ctx = context.WithValue(ctx, apiUtils.APIInstanceContextKey, mockAPI)
				ctx = context.WithValue(ctx, apiUtils.RequestIDContextKey, requestID)
				c.SetUserContext(ctx)
			},
			expectedLogger: logger.With().Str(requestIDLogField, requestID).Logger(),
			expectedCo:     nil,
			expectedAPI:    mockAPI,
			expectedRID:    requestID,
		},
		{
			name: "Context missing API instance",
			setupContext: func(c *fiber.Ctx) {
				ctx := c.UserContext()
				ctx = context.WithValue(ctx, consensusInstanceContextKey, mockConsensus)
				ctx = context.WithValue(ctx, apiUtils.RequestIDContextKey, requestID)
				c.SetUserContext(ctx)
			},
			expectedLogger: logger.With().Str(requestIDLogField, requestID).Logger(),
			expectedCo:     mockConsensus,
			expectedAPI:    nil,
			expectedRID:    requestID,
		},
		{
			name: "Context missing request ID",
			setupContext: func(c *fiber.Ctx) {
				ctx := c.UserContext()
				ctx = context.WithValue(ctx, consensusInstanceContextKey, mockConsensus)
				ctx = context.WithValue(ctx, apiUtils.APIInstanceContextKey, mockAPI)
				c.SetUserContext(ctx)
			},
			expectedLogger: logger.With().Str(requestIDLogField, "").Logger(),
			expectedCo:     mockConsensus,
			expectedAPI:    mockAPI,
			expectedRID:    "",
		},
		{
			name: "Empty context",
			setupContext: func(_ *fiber.Ctx) {
				// No values set in context
			},
			expectedLogger: logger.With().Str(requestIDLogField, "").Logger(),
			expectedCo:     nil,
			expectedAPI:    nil,
			expectedRID:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				tt.setupContext(c)
				uCtx := unpackCtx(c)

				assert.Equal(t, tt.expectedLogger, uCtx.logger)
				assert.Equal(t, tt.expectedCo, uCtx.co)
				assert.Equal(t, tt.expectedAPI, uCtx.api)
				assert.Equal(t, tt.expectedRID, uCtx.rid)

				return c.SendStatus(http.StatusOK)
			})

			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestParseAndValidate(t *testing.T) {
	t.Parallel()

	type mockRequest struct {
		Field string `validate:"required"`
	}

	requestID := "test-request-id"

	tests := []struct {
		name        string
		body        []byte
		validJSON   bool
		expectError bool
	}{
		{
			name:        "Valid request",
			body:        []byte(`{"Field":"value"}`),
			validJSON:   true,
			expectError: false,
		},
		{
			name:        "Invalid JSON body",
			body:        []byte(`{"Field":}`),
			validJSON:   false,
			expectError: true,
		},
		{
			name:        "Validation error",
			body:        []byte(`{"Field":""}`),
			validJSON:   true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				mockValidator := new(MockValidator)
				mockAPI := &APIV1Alpha{validator: mockValidator}
				mockConsensus := new(MockConsensus)

				if tt.validJSON {
					if tt.expectError {
						mockValidator.On("Validate", mock.Anything).Return(errors.New("validation error"))
					} else {
						mockValidator.On("Validate", mock.Anything).Return(nil)
					}
				}

				ctx := c.UserContext()
				ctx = context.WithValue(ctx, consensusInstanceContextKey, mockConsensus)
				ctx = context.WithValue(ctx, apiUtils.APIInstanceContextKey, mockAPI)
				ctx = context.WithValue(ctx, apiUtils.RequestIDContextKey, requestID)
				c.SetUserContext(ctx)

				req := &mockRequest{}
				err := parseAndValidate(c, req)

				if tt.expectError {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}

				if tt.validJSON {
					mockValidator.AssertExpectations(t)
				}

				return c.SendStatus(http.StatusOK)
			})

			req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(tt.body))
			req.Header.Set("Content-Type", "application/json")
			_, _ = app.Test(req)
		})
	}
}
