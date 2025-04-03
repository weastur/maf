package logging

import (
	"bytes"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sentryWrapper "github.com/weastur/maf/internal/utils/sentry"
)

type mockSentry struct {
	configured bool
}

func (m *mockSentry) IsConfigured() bool {
	return m.configured
}

func (m *mockSentry) Fork(_ string) *sentryWrapper.Wrapper {
	return &sentryWrapper.Wrapper{Hub: sentry.CurrentHub().Clone()}
}

func (m *mockSentry) GetHub() *sentry.Hub {
	return sentry.CurrentHub()
}

func TestInit(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		pretty        bool
		sentryEnabled bool
		expectError   bool
	}{
		{"ValidLevelPrettySentryEnabled", "debug", true, true, false},
		{"ValidLevelPrettySentryDisabled", "info", true, false, false},
		{"ValidLevelNonPrettySentryEnabled", "warn", false, true, false},
		{"ValidLevelNonPrettySentryDisabled", "error", false, false, false},
		{"InvalidLevel", "invalid", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSentry := &mockSentry{configured: tt.sentryEnabled}

			var buf bytes.Buffer
			consoleWriter := zerolog.ConsoleWriter{Out: &buf}
			log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()

			err := Init(tt.level, tt.pretty, mockSentry)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.level, zerolog.GlobalLevel().String())
			}
		})
	}
}
