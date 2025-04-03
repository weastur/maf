package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"resty.dev/v3"
)

type mockMarshalError struct{}

func (mockMarshalError) MarshalJSON() ([]byte, error) {
	return nil, errors.New("some error")
}

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.Nop())

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	t.Parallel()

	client := New("http://localhost", false)

	assert.Equal(t, "http://localhost", client.Host)
	assert.Equal(t, "root", client.AuthToken)
	assert.NotNil(t, client.rclient)
	assert.Equal(t, "X-Auth-Token", client.rclient.HeaderAuthorizationKey())
	assert.Empty(t, client.rclient.AuthScheme())
	assert.Equal(t, "root", client.rclient.AuthToken())
	assert.Contains(t, client.rclient.Header().Get("User-Agent"), "maf/")
	assert.Contains(t, client.rclient.ContentDecompresserKeys(), "br")
	assert.Contains(t, client.urlPrefix, "/api/v1alpha")
}

func TestNewWithTLS(t *testing.T) {
	t.Parallel()

	client := NewWithTLS("http://localhost", "server.crt", true)

	assert.Equal(t, "http://localhost", client.Host)
	assert.NotNil(t, client.rclient)
}

func TestNewWithMutualTLS(t *testing.T) {
	t.Parallel()

	client := NewWithMutualTLS("http://localhost", "client.crt", "client.key", "server.crt", true)

	assert.Equal(t, "http://localhost", client.Host)
	assert.NotNil(t, client.rclient)
}

func TestNewWithAutoTLS(t *testing.T) {
	t.Parallel()

	t.Run("NoConfig", func(t *testing.T) {
		t.Parallel()

		client := NewWithAutoTLS("http://localhost", nil, true)

		assert.Equal(t, "http://localhost", client.Host)
		assert.Equal(t, "root", client.AuthToken)
		assert.NotNil(t, client.rclient)
		assert.Contains(t, client.urlPrefix, "/api/v1alpha")
	})

	t.Run("EmptyConfig", func(t *testing.T) {
		t.Parallel()

		client := NewWithAutoTLS("http://localhost", &TLSConfig{}, true)

		assert.Equal(t, "http://localhost", client.Host)
		assert.Equal(t, "root", client.AuthToken)
		assert.NotNil(t, client.rclient)
		assert.Contains(t, client.urlPrefix, "/api/v1alpha")
	})

	t.Run("ServerCertOnly", func(t *testing.T) {
		t.Parallel()

		client := NewWithAutoTLS("http://localhost", &TLSConfig{ServerCertFile: "server.crt"}, true)

		assert.Equal(t, "http://localhost", client.Host)
		assert.NotNil(t, client.rclient)
	})

	t.Run("MutualTLS", func(t *testing.T) {
		t.Parallel()

		client := NewWithAutoTLS("http://localhost", &TLSConfig{
			CertFile:       "client.crt",
			KeyFile:        "client.key",
			ServerCertFile: "server.crt",
		}, true)

		assert.Equal(t, "http://localhost", client.Host)
		assert.NotNil(t, client.rclient)
	})
}

func TestClose(t *testing.T) {
	t.Parallel()

	client := &Client{
		rclient: resty.New(),
		logger:  zerolog.Nop(),
	}

	client.rclient.SetCloseConnection(true)
	err := client.Close()
	assert.NoError(t, err)
}

func TestParseRaftKVGetResponse(t *testing.T) {
	t.Parallel()

	client := &Client{
		logger: zerolog.Nop(),
	}

	t.Run("ValidResponse", func(t *testing.T) {
		t.Parallel()

		input := map[string]any{
			"Key":   "test-key",
			"Value": "test-value",
			"Exist": true,
		}

		expected := &raftKVGetResponse{
			Key:   "test-key",
			Value: "test-value",
			Exist: true,
		}

		result, err := client.parseRaftKVGetResponse(input)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("InvalidResponseFormat", func(t *testing.T) {
		t.Parallel()

		input := "[]"

		result, err := client.parseRaftKVGetResponse(input)
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("EmptyResponse", func(t *testing.T) {
		t.Parallel()

		input := map[string]any{}

		result, err := client.parseRaftKVGetResponse(input)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Key)
		assert.Empty(t, result.Value)
		assert.False(t, result.Exist)
	})

	t.Run("NotAJSON", func(t *testing.T) {
		t.Parallel()

		input := []byte("{invalid-json")

		result, err := client.parseRaftKVGetResponse(input)
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("MarhalError", func(t *testing.T) {
		t.Parallel()

		input := mockMarshalError{}

		result, err := client.parseRaftKVGetResponse(input)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestMakeURL(t *testing.T) {
	t.Parallel()

	client := &Client{
		urlPrefix: "http://localhost/api/v1alpha",
		logger:    zerolog.Nop(),
	}

	t.Run("SingleElement", func(t *testing.T) {
		t.Parallel()

		result := client.makeURL("raft/join")
		expected := "http://localhost/api/v1alpha/raft/join"
		assert.Equal(t, expected, result)
	})

	t.Run("MultipleElements", func(t *testing.T) {
		t.Parallel()

		result := client.makeURL("raft", "kv", "key1")
		expected := "http://localhost/api/v1alpha/raft/kv/key1"
		assert.Equal(t, expected, result)
	})

	t.Run("EmptyElements", func(t *testing.T) {
		t.Parallel()

		result := client.makeURL()
		expected := "http://localhost/api/v1alpha"
		assert.Equal(t, expected, result)
	})

	t.Run("TrailingSlash", func(t *testing.T) {
		t.Parallel()

		clientWithSlash := &Client{
			urlPrefix: "http://localhost/api/v1alpha/",
			logger:    zerolog.Nop(),
		}
		result := clientWithSlash.makeURL("raft", "info")
		expected := "http://localhost/api/v1alpha/raft/info"
		assert.Equal(t, expected, result)
	})

	t.Run("InvalidURLPrefix", func(t *testing.T) {
		t.Parallel()

		clientInvalid := &Client{
			urlPrefix: "http://[::1]:invalid",
			logger:    zerolog.Nop(),
		}

		assert.Panics(t, func() {
			clientInvalid.makeURL("raft", "info")
		})
	})
}

func TestRaftJoin(t *testing.T) {
	t.Parallel()

	t.Run("SuccessfulJoin", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1alpha/raft/join", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)

			var req raftJoinRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.NoError(t, err)
			assert.Equal(t, "server-1", req.ServerID)
			assert.Equal(t, "127.0.0.1:8080", req.Addr)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response{Status: "success"})
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftJoin("server-1", "127.0.0.1:8080")
		require.NoError(t, err)
	})

	t.Run("APIError", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response{Status: "error", Error: "internal error"})
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftJoin("server-1", "127.0.0.1:8080")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "internal error")
	})

	t.Run("ServerError", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftJoin("server-1", "127.0.0.1:8080")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bad status code")
	})

	t.Run("InvalidResponseFormat", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("[]"))
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftJoin("server-1", "127.0.0.1:8080")
		require.Error(t, err)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{invalid-json"))
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftJoin("server-1", "127.0.0.1:8080")
		require.Error(t, err)
	})

	t.Run("RequestFailure", func(t *testing.T) {
		t.Parallel()

		client := New("http://invalid-url", false)
		err := client.RaftJoin("server-1", "127.0.0.1:8080")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to perform join request")
	})
}

func TestRaftForget(t *testing.T) {
	t.Parallel()

	t.Run("SuccessfulForget", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1alpha/raft/forget", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)

			var req raftForgetRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.NoError(t, err)
			assert.Equal(t, "server-1", req.ServerID)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response{Status: "success"})
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftForget("server-1")
		require.NoError(t, err)
	})

	t.Run("APIError", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response{Status: "error", Error: "internal error"})
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftForget("server-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "internal error")
	})

	t.Run("ServerError", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftForget("server-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bad status code")
	})

	t.Run("InvalidResponseFormat", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("[]"))
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftForget("server-1")
		require.Error(t, err)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{invalid-json"))
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftForget("server-1")
		require.Error(t, err)
	})

	t.Run("RequestFailure", func(t *testing.T) {
		t.Parallel()

		client := New("http://invalid-url", false)
		err := client.RaftForget("server-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to perform forget request")
	})
}

func TestRaftKVGet(t *testing.T) {
	t.Parallel()

	t.Run("SuccessfulGet", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1alpha/raft/kv/test-key", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response{
				Status: "success",
				Data: map[string]any{
					"Key":   "test-key",
					"Value": "test-value",
					"Exist": true,
				},
			})
		}))
		defer server.Close()

		client := New(server.URL, false)
		value, exist, err := client.RaftKVGet("test-key")
		require.NoError(t, err)
		assert.True(t, exist)
		assert.Equal(t, "test-value", value)
	})

	t.Run("KeyDoesNotExist", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response{
				Status: "success",
				Data: map[string]any{
					"Key":   "test-key",
					"Value": "",
					"Exist": false,
				},
			})
		}))
		defer server.Close()

		client := New(server.URL, false)
		value, exist, err := client.RaftKVGet("test-key")
		require.NoError(t, err)
		assert.False(t, exist)
		assert.Empty(t, value)
	})

	t.Run("APIError", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response{
				Status: "error",
				Error:  "internal error",
			})
		}))
		defer server.Close()

		client := New(server.URL, false)
		value, exist, err := client.RaftKVGet("test-key")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "internal error")
		assert.False(t, exist)
		assert.Empty(t, value)
	})

	t.Run("ServerError", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := New(server.URL, false)
		value, exist, err := client.RaftKVGet("test-key")
		require.Error(t, err)
		assert.False(t, exist)
		assert.Empty(t, value)
	})

	t.Run("InvalidResponseFormat", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("[]"))
		}))
		defer server.Close()

		client := New(server.URL, false)
		value, exist, err := client.RaftKVGet("test-key")
		require.Error(t, err)
		assert.False(t, exist)
		assert.Empty(t, value)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{invalid-json"))
		}))
		defer server.Close()

		client := New(server.URL, false)
		value, exist, err := client.RaftKVGet("test-key")
		require.Error(t, err)
		assert.False(t, exist)
		assert.Empty(t, value)
	})

	t.Run("RequestFailure", func(t *testing.T) {
		t.Parallel()

		client := New("http://invalid-url", false)
		value, exist, err := client.RaftKVGet("test-key")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to perform KV get request")
		assert.False(t, exist)
		assert.Empty(t, value)
	})
}

func TestRaftKVSet(t *testing.T) {
	t.Parallel()

	t.Run("SuccessfulSet", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1alpha/raft/kv", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)

			var req raftKVSetRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.NoError(t, err)
			assert.Equal(t, "test-key", req.Key)
			assert.Equal(t, "test-value", req.Value)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response{Status: "success"})
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftKVSet("test-key", "test-value")
		assert.NoError(t, err)
	})

	t.Run("APIError", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response{Status: "error", Error: "internal error"})
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftKVSet("test-key", "test-value")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "internal error")
	})

	t.Run("ServerError", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftKVSet("test-key", "test-value")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bad status code")
	})

	t.Run("InvalidResponseFormat", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("[]"))
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftKVSet("test-key", "test-value")
		assert.Error(t, err)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{invalid-json"))
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftKVSet("test-key", "test-value")
		assert.Error(t, err)
	})

	t.Run("RequestFailure", func(t *testing.T) {
		t.Parallel()

		client := New("http://invalid-url", false)
		err := client.RaftKVSet("test-key", "test-value")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to perform KV set request")
	})
}

func TestRaftKVDelete(t *testing.T) {
	t.Parallel()

	t.Run("SuccessfulDelete", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1alpha/raft/kv/test-key", r.URL.Path)
			assert.Equal(t, http.MethodDelete, r.Method)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response{Status: "success"})
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftKVDelete("test-key")
		assert.NoError(t, err)
	})

	t.Run("APIError", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response{Status: "error", Error: "internal error"})
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftKVDelete("test-key")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "internal error")
	})

	t.Run("ServerError", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftKVDelete("test-key")
		require.Error(t, err)
	})

	t.Run("InvalidResponseFormat", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("[]"))
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftKVDelete("test-key")
		require.Error(t, err)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{invalid-json"))
		}))
		defer server.Close()

		client := New(server.URL, false)
		err := client.RaftKVDelete("test-key")
		require.Error(t, err)
	})

	t.Run("RequestFailure", func(t *testing.T) {
		t.Parallel()

		client := New("http://invalid-url", false)
		err := client.RaftKVDelete("test-key")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to perform KV delete request")
	})
}

func TestRaftInfo(t *testing.T) {
	t.Parallel()

	t.Run("SuccessfulInfo", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1alpha/raft/info", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "true", r.URL.Query().Get("include_stats"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response{
				Status: "success",
				Data: map[string]any{
					"ClusterSize": 3,
					"Leader":      "node-1",
				},
			})
		}))
		defer server.Close()

		client := New(server.URL, false)
		data, err := client.RaftInfo(true)
		require.NoError(t, err)
		assert.NotNil(t, data)

		info, ok := data.(map[string]any)
		require.True(t, ok)

		clusterSize, ok := info["ClusterSize"].(float64)
		require.True(t, ok)
		assert.Equal(t, 3, int(clusterSize))
		assert.Equal(t, "node-1", info["Leader"])
	})

	t.Run("APIError", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response{
				Status: "error",
				Error:  "internal error",
			})
		}))
		defer server.Close()

		client := New(server.URL, false)
		data, err := client.RaftInfo(false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "internal error")
		assert.Nil(t, data)
	})

	t.Run("ServerError", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := New(server.URL, false)
		data, err := client.RaftInfo(false)
		require.Error(t, err)
		assert.Nil(t, data)
	})

	t.Run("InvalidResponseFormat", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("[]"))
		}))
		defer server.Close()

		client := New(server.URL, false)
		data, err := client.RaftInfo(false)
		require.Error(t, err)
		assert.Nil(t, data)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{invalid-json"))
		}))
		defer server.Close()

		client := New(server.URL, false)
		data, err := client.RaftInfo(false)
		require.Error(t, err)
		assert.Nil(t, data)
	})

	t.Run("RequestFailure", func(t *testing.T) {
		t.Parallel()

		client := New("http://invalid-url", false)
		data, err := client.RaftInfo(false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to perform raft info request")
		assert.Nil(t, data)
	})
}
