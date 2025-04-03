package validate

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestMutualTLSValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		config        map[string]string
		expectedError error
	}{
		{
			name: "Valid configuration for agent.http",
			config: map[string]string{
				"agent.http.client_cert_file": "client.crt",
				"agent.http.cert_file":        "cert.crt",
				"agent.http.key_file":         "key.key",
			},
			expectedError: nil,
		},
		{
			name: "Missing cert_file for agent.http",
			config: map[string]string{
				"agent.http.client_cert_file": "client.crt",
				"agent.http.key_file":         "key.key",
			},
			expectedError: ErrMutualTLS,
		},
		{
			name: "Missing key_file for server.http",
			config: map[string]string{
				"server.http.client_cert_file": "client.crt",
				"server.http.cert_file":        "cert.crt",
			},
			expectedError: ErrMutualTLS,
		},
		{
			name: "Valid configuration for server.http.clients.server",
			config: map[string]string{
				"server.http.clients.server.server_cert_file": "server.crt",
				"server.http.clients.server.cert_file":        "cert.crt",
				"server.http.clients.server.key_file":         "key.key",
			},
			expectedError: nil,
		},
		{
			name: "Missing cert_file for server.http.clients.agent",
			config: map[string]string{
				"server.http.clients.agent.server_cert_file": "server.crt",
				"server.http.clients.agent.key_file":         "key.key",
			},
			expectedError: ErrMutualTLS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := viper.New()
			for key, value := range tt.config {
				v.Set(key, value)
			}

			mutualTLS := NewMutualTLS()
			err := mutualTLS.Validate(v)
			require.ErrorIs(t, err, tt.expectedError)
		})
	}
}
