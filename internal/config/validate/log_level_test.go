package validate

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogLevel_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    map[string]string
		expectErr bool
	}{
		{
			name: "Valid log levels",
			config: map[string]string{
				"agent.log.level":  "info",
				"server.log.level": "debug",
			},
			expectErr: false,
		},
		{
			name: "Invalid agent log level",
			config: map[string]string{
				"agent.log.level":  "invalid",
				"server.log.level": "debug",
			},
			expectErr: true,
		},
		{
			name: "Invalid server log level",
			config: map[string]string{
				"agent.log.level":  "info",
				"server.log.level": "invalid",
			},
			expectErr: true,
		},
		{
			name: "Both log levels invalid",
			config: map[string]string{
				"agent.log.level":  "invalid",
				"server.log.level": "invalid",
			},
			expectErr: true,
		},
		{
			name:      "Missing log levels",
			config:    map[string]string{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()
			for key, value := range tt.config {
				v.Set(key, value)
			}

			logLevelValidator := NewLogLevel()
			err := logLevelValidator.Validate(v)

			if tt.expectErr {
				require.Error(t, err)
				assert.Equal(t, ErrLogLevel, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
