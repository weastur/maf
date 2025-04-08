package v1alpha

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/weastur/maf/internal/server/worker/raft"
	httpUtils "github.com/weastur/maf/internal/utils/http"
	apiUtils "github.com/weastur/maf/internal/utils/http/api"
)

const requestID = "test-request-id"

func getTestFiberApp() (*fiber.App, *MockConsensus) {
	app := fiber.New(fiber.Config{
		ErrorHandler: httpUtils.ErrorHandler,
	})
	mockConsensus := new(MockConsensus)
	mockValidator := new(MockValidator)
	mockAPI := &APIV1Alpha{
		version:   "v1alpha",
		prefix:    "/v1alpha",
		validator: mockValidator,
	}

	mockValidator.On("Validate", mock.Anything).Return(nil).Once()
	app.Use(func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		ctx = context.WithValue(ctx, consensusInstanceContextKey, mockConsensus)
		ctx = context.WithValue(ctx, apiUtils.APIInstanceContextKey, mockAPI)
		ctx = context.WithValue(ctx, apiUtils.RequestIDContextKey, requestID)
		c.SetUserContext(ctx)

		return c.Next()
	})

	return app, mockConsensus
}

func TestRaftJoinHandler(t *testing.T) {
	t.Parallel()

	t.Run("successful join", func(t *testing.T) {
		t.Parallel()

		app, mockConsensus := getTestFiberApp()
		app.Post("/test", raftJoinHandler)

		defer app.Shutdown()
		mockConsensus.On("Join", "server-1", "127.0.0.1").Return(nil).Once()

		body := `{"serverId": "server-1", "addr": "127.0.0.1"}`
		req, _ := http.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		mockConsensus.AssertExpectations(t)
	})

	t.Run("error on join", func(t *testing.T) {
		t.Parallel()

		app, mockConsensus := getTestFiberApp()
		app.Post("/test", raftJoinHandler)

		defer app.Shutdown()
		mockConsensus.On("Join", "server-1", "127.0.0.1").Return(errors.New("join error")).Once()

		reqBody := `{"serverId": "server-1", "addr": "127.0.0.1"}`
		req, _ := http.NewRequest(http.MethodPost, "/test", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		respBody, _ := io.ReadAll(resp.Body)

		var response map[string]any
		err = json.Unmarshal(respBody, &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
		assert.Equal(t, "join error", response["error"])

		mockConsensus.AssertExpectations(t)
	})

	t.Run("invalid body", func(t *testing.T) {
		t.Parallel()

		app, _ := getTestFiberApp()
		app.Post("/test", raftJoinHandler)

		defer app.Shutdown()

		reqBody := `{"serverId": 123, "addr": 127.0.0.1}` // Invalid body: serverId should be a string
		req, _ := http.NewRequest(http.MethodPost, "/test", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		respBody, _ := io.ReadAll(resp.Body)

		var response map[string]any
		err = json.Unmarshal(respBody, &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
		assert.NotEmpty(t, response["error"])
	})
}

func TestRaftForgetHandler(t *testing.T) {
	t.Parallel()

	t.Run("successful forget", func(t *testing.T) {
		t.Parallel()

		app, mockConsensus := getTestFiberApp()
		app.Post("/test", raftForgetHandler)

		defer app.Shutdown()
		mockConsensus.On("Forget", "server-1").Return(nil).Once()

		body := `{"serverId": "server-1"}`
		req, _ := http.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		mockConsensus.AssertExpectations(t)
	})

	t.Run("error on forget", func(t *testing.T) {
		t.Parallel()

		app, mockConsensus := getTestFiberApp()
		app.Post("/test", raftForgetHandler)

		defer app.Shutdown()
		mockConsensus.On("Forget", "server-1").Return(errors.New("forget error")).Once()

		body := `{"serverId": "server-1"}`
		req, _ := http.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		respBody, _ := io.ReadAll(resp.Body)

		var response map[string]any
		err = json.Unmarshal(respBody, &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
		assert.Equal(t, "forget error", response["error"])

		mockConsensus.AssertExpectations(t)
	})

	t.Run("invalid body", func(t *testing.T) {
		t.Parallel()

		app, _ := getTestFiberApp()
		app.Post("/test", raftForgetHandler)

		defer app.Shutdown()

		body := `{"serverId": 123}` // Invalid body: serverId should be a string
		req, _ := http.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		respBody, _ := io.ReadAll(resp.Body)

		var response map[string]any
		err = json.Unmarshal(respBody, &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
		assert.NotEmpty(t, response["error"])
	})
}

func TestRaftInfoHandler(t *testing.T) {
	t.Parallel()

	t.Run("successful info retrieval", func(t *testing.T) {
		t.Parallel()

		app, mockConsensus := getTestFiberApp()
		app.Get("/test", raftInfoHandler)

		defer app.Shutdown()
		mockConsensus.On("GetInfo", true).Return(&raft.Info{
			State: "leader",
			Stats: map[string]string{"uptime": "100s"},
		}, nil).Once()

		req, _ := http.NewRequest(http.MethodGet, "/test?include_stats=true", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)

		var response map[string]any
		err = json.Unmarshal(respBody, &response)
		require.NoError(t, err)

		assert.Equal(t, "success", response["status"])
		assert.Contains(t, response, "data")
		data := response["data"].(map[string]any)
		assert.Equal(t, "leader", data["state"])
		assert.Contains(t, data, "stats")
		stats := data["stats"].(map[string]any)
		assert.Equal(t, "100s", stats["uptime"])

		mockConsensus.AssertExpectations(t)
	})

	t.Run("error on GetInfo call", func(t *testing.T) {
		t.Parallel()

		app, mockConsensus := getTestFiberApp()
		app.Get("/test", raftInfoHandler)

		defer app.Shutdown()
		mockConsensus.On("GetInfo", true).Return(&raft.Info{}, errors.New("get info error")).Once()

		req, _ := http.NewRequest(http.MethodGet, "/test?include_stats=true", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		respBody, _ := io.ReadAll(resp.Body)

		var response map[string]any
		err = json.Unmarshal(respBody, &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
		assert.Equal(t, "get info error", response["error"])

		mockConsensus.AssertExpectations(t)
	})
}

func TestRaftKVGetHandler(t *testing.T) {
	t.Parallel()

	t.Run("successful key retrieval", func(t *testing.T) {
		t.Parallel()

		app, mockConsensus := getTestFiberApp()
		app.Get("/test/:key", raftKVGetHandler)

		defer app.Shutdown()
		mockConsensus.On("Get", "test-key").Return("test-value", true).Once()

		req, _ := http.NewRequest(http.MethodGet, "/test/test-key", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)

		var response map[string]any
		err = json.Unmarshal(respBody, &response)
		require.NoError(t, err)

		assert.Equal(t, "success", response["status"])
		assert.Contains(t, response, "data")
		data := response["data"].(map[string]any)
		assert.Equal(t, "test-key", data["key"])
		assert.Equal(t, "test-value", data["value"])
		assert.True(t, data["exist"].(bool))

		mockConsensus.AssertExpectations(t)
	})
}

func TestRaftKVSetHandler(t *testing.T) {
	t.Parallel()

	t.Run("successful key set", func(t *testing.T) {
		t.Parallel()

		app, mockConsensus := getTestFiberApp()
		app.Post("/test", raftKVSetHandler)

		defer app.Shutdown()
		mockConsensus.On("Set", "test-key", "test-value").Return(nil).Once()

		body := `{"key": "test-key", "value": "test-value"}`
		req, _ := http.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		mockConsensus.AssertExpectations(t)
	})

	t.Run("error on key set", func(t *testing.T) {
		t.Parallel()

		app, mockConsensus := getTestFiberApp()
		app.Post("/test", raftKVSetHandler)

		defer app.Shutdown()
		mockConsensus.On("Set", "test-key", "test-value").Return(errors.New("set error")).Once()

		body := `{"key": "test-key", "value": "test-value"}`
		req, _ := http.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		respBody, _ := io.ReadAll(resp.Body)

		var response map[string]any
		err = json.Unmarshal(respBody, &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
		assert.Equal(t, "set error", response["error"])

		mockConsensus.AssertExpectations(t)
	})

	t.Run("invalid body", func(t *testing.T) {
		t.Parallel()

		app, _ := getTestFiberApp()
		app.Post("/test", raftKVSetHandler)

		defer app.Shutdown()

		body := `{"key": 123, "value": "test-value"}` // Invalid body: key should be a string
		req, _ := http.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		respBody, _ := io.ReadAll(resp.Body)

		var response map[string]any
		err = json.Unmarshal(respBody, &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
		assert.NotEmpty(t, response["error"])
	})
}

func TestRaftKVDeleteHandler(t *testing.T) {
	t.Parallel()

	t.Run("successful key deletion", func(t *testing.T) {
		t.Parallel()

		app, mockConsensus := getTestFiberApp()
		app.Delete("/test/:key", raftKVDeleteHandler)

		defer app.Shutdown()
		mockConsensus.On("Delete", "test-key").Return(nil).Once()

		req, _ := http.NewRequest(http.MethodDelete, "/test/test-key", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		mockConsensus.AssertExpectations(t)
	})

	t.Run("error on key deletion", func(t *testing.T) {
		t.Parallel()

		app, mockConsensus := getTestFiberApp()
		app.Delete("/test/:key", raftKVDeleteHandler)

		defer app.Shutdown()
		mockConsensus.On("Delete", "test-key").Return(errors.New("delete error")).Once()

		req, _ := http.NewRequest(http.MethodDelete, "/test/test-key", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		respBody, _ := io.ReadAll(resp.Body)

		var response map[string]any
		err = json.Unmarshal(respBody, &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
		assert.Equal(t, "delete error", response["error"])

		mockConsensus.AssertExpectations(t)
	})
}
