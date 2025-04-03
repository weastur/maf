package sentry

import (
	"os"
	"testing"

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

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("Empty DSN", func(t *testing.T) {
		t.Parallel()

		wrapper, err := New("")
		require.NoError(t, err)
		assert.NotNil(t, wrapper)
		assert.False(t, wrapper.IsConfigured())
	})

	t.Run("Invalid DSN", func(t *testing.T) {
		t.Parallel()

		wrapper, err := New("invalid-dsn")
		require.Error(t, err)
		assert.Nil(t, wrapper)
	})

	t.Run("Valid DSN", func(t *testing.T) {
		t.Parallel()

		wrapper, err := New("https://examplePublicKey@o0.ingest.sentry.io/0")
		require.NoError(t, err)
		assert.NotNil(t, wrapper)
		assert.True(t, wrapper.IsConfigured())
	})
}

func TestWrapper_GetHub(t *testing.T) {
	t.Parallel()

	wrapper, err := New("")
	require.NoError(t, err)
	assert.Nil(t, wrapper.GetHub())

	wrapper, err = New("https://examplePublicKey@o0.ingest.sentry.io/0")
	require.NoError(t, err)
	assert.NotNil(t, wrapper.GetHub())
}

func TestWrapper_Flush(t *testing.T) {
	t.Parallel()

	wrapper, err := New("")
	require.NoError(t, err)
	wrapper.Flush() // Should not panic

	wrapper, err = New("https://examplePublicKey@o0.ingest.sentry.io/0")
	require.NoError(t, err)
	wrapper.Flush() // Should not panic
}

func TestWrapper_Recover(t *testing.T) {
	t.Parallel()

	wrapper, err := New("")
	require.NoError(t, err)
	assert.Panics(t, func() {
		defer wrapper.Recover()
		panic("test panic")
	})

	wrapper, err = New("https://examplePublicKey@o0.ingest.sentry.io/0")
	require.NoError(t, err)
	assert.Panics(t, func() {
		defer wrapper.Recover()
		panic("test panic")
	})
}

func TestWrapper_Fork(t *testing.T) {
	t.Parallel()

	wrapper, err := New("")
	require.NoError(t, err)

	forked := wrapper.Fork("test-scope")
	assert.Equal(t, wrapper, forked)

	wrapper, err = New("https://examplePublicKey@o0.ingest.sentry.io/0")
	require.NoError(t, err)

	forked = wrapper.Fork("test-scope")
	assert.NotEqual(t, wrapper, forked)
	assert.True(t, forked.IsConfigured())
}

func TestWrapper_IsConfigured(t *testing.T) {
	t.Parallel()

	wrapper, err := New("")
	require.NoError(t, err)
	assert.False(t, wrapper.IsConfigured())

	wrapper, err = New("https://examplePublicKey@o0.ingest.sentry.io/0")
	require.NoError(t, err)
	assert.True(t, wrapper.IsConfigured())
}
