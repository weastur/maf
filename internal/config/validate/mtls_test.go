package validate

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestMutualTLS_Validate(t *testing.T) {
	tests := []struct {
		name          string
		configSetup   func(*viper.Viper)
		expectedError error
	}{
		{
			name: "Valid configuration for agent.http",
			configSetup: func(v *viper.Viper) {
				v.Set("agent.http.client_cert_file", "client.crt")
				v.Set("agent.http.cert_file", "cert.crt")
				v.Set("agent.http.key_file", "key.key")
			},
			expectedError: nil,
		},
		{
			name: "Missing cert_file for agent.http",
			configSetup: func(v *viper.Viper) {
				v.Set("agent.http.client_cert_file", "client.crt")
				v.Set("agent.http.key_file", "key.key")
			},
			expectedError: ErrMutualTLS,
		},
		{
			name: "Missing key_file for server.http",
			configSetup: func(v *viper.Viper) {
				v.Set("server.http.client_cert_file", "client.crt")
				v.Set("server.http.cert_file", "cert.crt")
			},
			expectedError: ErrMutualTLS,
		},
		{
			name: "Valid configuration for server.http.clients.server",
			configSetup: func(v *viper.Viper) {
				v.Set("server.http.clients.server.server_cert_file", "server.crt")
				v.Set("server.http.clients.server.cert_file", "cert.crt")
				v.Set("server.http.clients.server.key_file", "key.key")
			},
			expectedError: nil,
		},
		{
			name: "Missing cert_file for server.http.clients.agent",
			configSetup: func(v *viper.Viper) {
				v.Set("server.http.clients.agent.server_cert_file", "server.crt")
				v.Set("server.http.clients.agent.key_file", "key.key")
			},
			expectedError: ErrMutualTLS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()
			tt.configSetup(v)

			mutualTLS := NewMutualTLS()
			err := mutualTLS.Validate(v)

			assert.ErrorIs(t, tt.expectedError, err)
		})
	}
}
