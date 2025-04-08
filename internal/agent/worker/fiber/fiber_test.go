package fiber

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSentry struct {
	mock.Mock
}

func (m *MockSentry) Recover() {
	m.Called()
}

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	log.Logger = log.Output(zerolog.Nop())

	os.Exit(m.Run())
}

func TestFiber_IsLive(t *testing.T) {
	t.Parallel()

	f := &Fiber{
		logger: log.With().Logger(),
	}

	ctx := new(fiber.Ctx)
	isLive := f.IsLive(ctx)

	assert.True(t, isLive)
}

func TestFiber_IsReady(t *testing.T) {
	t.Parallel()

	f := &Fiber{
		logger: log.With().Logger(),
	}

	ctx := new(fiber.Ctx)
	isReady := f.IsReady(ctx)

	assert.True(t, isReady)
}

func TestFiber_Stop(t *testing.T) {
	t.Parallel()

	t.Run("successful shutdown", func(t *testing.T) {
		t.Parallel()

		shutdownCalled := false
		app := fiber.New()
		app.Hooks().OnShutdown(func() error {
			shutdownCalled = true

			return nil
		})

		mockLogger := log.With().Logger()

		f := &Fiber{
			app:    app,
			logger: mockLogger,
			config: &Config{
				ShutdownTimeout: 5 * time.Second,
			},
		}

		f.Stop()
		assert.True(t, shutdownCalled)
	})

	t.Run("shutdown with error", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		mockLogger := log.With().Logger()

		f := &Fiber{
			app:    app,
			logger: mockLogger,
			config: &Config{
				ShutdownTimeout: 5 * time.Second,
			},
		}

		app.Hooks().OnShutdown(func() error {
			return errors.New("shutdown error")
		})

		assert.NotPanics(t, func() {
			f.Stop()
		}, "Stop should not panic even if shutdown returns an error")
	})
}

func TestFiber_Run(t *testing.T) {
	t.Parallel()

	t.Run("error during Listen", func(t *testing.T) {
		t.Parallel()

		mockSentry := new(MockSentry)
		mockLogger := log.With().Logger()

		wg := &sync.WaitGroup{}
		done := make(chan struct{})
		close(done)

		mockSentry.On("Recover").Return()

		app := fiber.New()

		f := &Fiber{
			config: &Config{
				Addr: "invalid_address", // Use an invalid address to trigger an error
			},
			app:    app,
			logger: mockLogger,
			sentry: mockSentry,
		}

		assert.NotPanics(t, func() {
			f.Run(wg)
			wg.Wait()
		}, "Run should not panic even if Listen returns an error")

		mockSentry.AssertCalled(t, "Recover")
	})
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("successful initialization", func(t *testing.T) {
		t.Parallel()

		mockSentry := new(MockSentry)

		mockSentry.On("Recover").Return()

		config := &Config{
			Addr:            "127.0.0.1:8080",
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     10 * time.Second,
			ShutdownTimeout: 5 * time.Second,
		}

		var f *Fiber

		assert.NotPanics(t, func() {
			f = New(config, mockSentry)
		}, "New should not panic when initializing Fiber")

		assert.NotNil(t, f)
		assert.Equal(t, config, f.config)
		assert.Equal(t, mockSentry, f.sentry)
		assert.NotNil(t, f.app)
	})
}
