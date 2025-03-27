package logging

import (
	"testing"

	sentrygo "github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sentryUtils "github.com/weastur/maf/internal/utils/sentry"
)

func TestInit_WithValidLevelAndPretty(t *testing.T) {
	sentrygo.CurrentHub().BindClient(nil) // Reset the hub

	err := Init("debug", true)
	assert.NoError(t, err, "Init should not return an error for valid level and pretty")
}

func TestInit_WithInvalidLevel(t *testing.T) {
	sentrygo.CurrentHub().BindClient(nil) // Reset the hub

	err := Init("unknown", true)
	require.Error(t, err, "Init should return error for invalid level")
}

func TestInit_WithSentry(t *testing.T) {
	err := sentryUtils.Init("https://examplePublicKey@o0.ingest.sentry.io/0")
	require.NoError(t, err, "Sentry should be configured for valid DSN")

	for _, pretty := range []bool{true, false} {
		err := Init("debug", pretty)

		require.NoError(t, err,
			"Init should not return an error for valid level, pretty equal to %t and sentry configured",
			pretty)
	}
}
