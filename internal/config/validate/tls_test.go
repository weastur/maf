package validate

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestTLSValidate(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]interface{}
		expectedError error
	}{
		{
			name: "Valid configuration for agent.http and server.http",
			config: map[string]interface{}{
				"agent.http.cert_file":  "cert.pem",
				"agent.http.key_file":   "key.pem",
				"server.http.cert_file": "cert.pem",
				"server.http.key_file":  "key.pem",
			},
			expectedError: nil,
		},
		{
			name: "Missing key_file for agent.http",
			config: map[string]interface{}{
				"agent.http.cert_file":  "cert.pem",
				"server.http.cert_file": "cert.pem",
				"server.http.key_file":  "key.pem",
			},
			expectedError: ErrTLS,
		},
		{
			name: "Missing cert_file for server.http",
			config: map[string]interface{}{
				"agent.http.cert_file": "cert.pem",
				"agent.http.key_file":  "key.pem",
				"server.http.key_file": "key.pem",
			},
			expectedError: ErrTLS,
		},
		{
			name: "Valid configuration for server.http.clients",
			config: map[string]interface{}{
				"server.http.clients.server.cert_file": "cert.pem",
				"server.http.clients.server.key_file":  "key.pem",
				"server.http.clients.agent.cert_file":  "cert.pem",
				"server.http.clients.agent.key_file":   "key.pem",
			},
			expectedError: nil,
		},
		{
			name: "Missing key_file for server.http.clients.server",
			config: map[string]interface{}{
				"server.http.clients.server.cert_file": "cert.pem",
				"server.http.clients.agent.cert_file":  "cert.pem",
				"server.http.clients.agent.key_file":   "key.pem",
			},
			expectedError: ErrTLS,
		},
		{
			name: "Missing cert_file for server.http.clients.agent",
			config: map[string]interface{}{
				"server.http.clients.server.cert_file": "cert.pem",
				"server.http.clients.server.key_file":  "key.pem",
				"server.http.clients.agent.key_file":   "key.pem",
			},
			expectedError: ErrTLS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()
			for key, value := range tt.config {
				v.Set(key, value)
			}

			tlsValidator := NewTLS()
			err := tlsValidator.Validate(v)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
