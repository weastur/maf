package sentry

import (
	"os"
	"testing"

	sentrygo "github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	log.Logger = log.Output(zerolog.Nop())

	os.Exit(m.Run())
}

func TestInit(t *testing.T) {
	t.Run("Empty DSN", func(t *testing.T) {
		sentrygo.CurrentHub().BindClient(nil) // Reset the hub

		err := Init("")
		require.NoError(t, err, "Init should not return an error for empty DSN")
	})

	t.Run("Invalid DSN", func(t *testing.T) {
		sentrygo.CurrentHub().BindClient(nil) // Reset the hub

		err := Init("https://not-a-valid-url")
		require.Error(t, err, "Init should return error for invalid DSN")
		assert.False(t, IsConfigured(), "Sentry should be configured after Init")
	})

	t.Run("Valid DSN", func(t *testing.T) {
		sentrygo.CurrentHub().BindClient(nil) // Reset the hub

		err := Init("https://examplePublicKey@o0.ingest.sentry.io/0")
		require.NoError(t, err, "Init should not return an error for valid DSN")
		assert.True(t, IsConfigured(), "Sentry should be configured after Init")
	})
}

func TestFlush(t *testing.T) {
	t.Run("Not Configured", func(_ *testing.T) {
		sentrygo.CurrentHub().BindClient(nil) // Reset the hub
		Flush()                               // Should not panic or cause issues
	})

	t.Run("Configured", func(_ *testing.T) {
		sentrygo.CurrentHub().BindClient(nil) // Reset the hub

		_ = Init("https://examplePublicKey@o0.ingest.sentry.io/0")

		Flush() // Should not panic or cause issues
	})
}

func TestRecover(t *testing.T) {
	t.Run("Not Configured", func(t *testing.T) {
		sentrygo.CurrentHub().BindClient(nil) // Reset the hub

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Recover should not panic when Sentry is not configured")
			}
		}()
		Recover(sentrygo.CurrentHub())
	})

	t.Run("Configured", func(t *testing.T) {
		sentrygo.CurrentHub().BindClient(nil) // Reset the hub

		_ = Init("https://examplePublicKey@o0.ingest.sentry.io/0")

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Recover should re-panic after handling the error")
			}
		}()
		defer Recover(sentrygo.CurrentHub())
		panic("test panic")
	})
}

func TestFork(t *testing.T) {
	t.Run("Not Configured", func(t *testing.T) {
		sentrygo.CurrentHub().BindClient(nil) // Reset the hub

		hub := Fork("test-scope")
		assert.NotNil(t, hub, "Fork should return a hub even if Sentry is not configured")
	})

	t.Run("Configured", func(t *testing.T) {
		sentrygo.CurrentHub().BindClient(nil) // Reset the hub

		_ = Init("https://examplePublicKey@o0.ingest.sentry.io/0")
		hub := Fork("test-scope")
		assert.NotNil(t, hub, "Fork should return a valid hub when Sentry is configured")
	})
}

func TestIsConfigured(t *testing.T) {
	t.Run("Not Configured", func(t *testing.T) {
		sentrygo.CurrentHub().BindClient(nil) // Reset the hub
		assert.False(t, IsConfigured(), "IsConfigured should return false when Sentry is not initialized")
	})

	t.Run("Configured", func(t *testing.T) {
		sentrygo.CurrentHub().BindClient(nil) // Reset the hub

		_ = Init("https://examplePublicKey@o0.ingest.sentry.io/0")

		assert.True(t, IsConfigured(), "IsConfigured should return true when Sentry is initialized")
	})
}
