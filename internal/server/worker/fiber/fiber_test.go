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
	"github.com/weastur/maf/internal/server/worker/raft"
)

type MockConsensus struct {
	mock.Mock
}

func (m *MockConsensus) IsReady() bool {
	args := m.Called()

	return args.Bool(0)
}

func (m *MockConsensus) IsLive() bool {
	args := m.Called()

	return args.Bool(0)
}

func (m *MockConsensus) Set(key, value string) error {
	args := m.Called(key, value)

	return args.Error(0)
}

func (m *MockConsensus) SubscribeOnLeadershipChanges(ch raft.LeadershipChangesCh) {
	m.Called(ch)
}

func (m *MockConsensus) IsLeader() bool {
	args := m.Called()

	return args.Bool(0)
}

func (m *MockConsensus) Join(serverID, addr string) error {
	args := m.Called(serverID, addr)

	return args.Error(0)
}

func (m *MockConsensus) Forget(serverID string) error {
	args := m.Called(serverID)

	return args.Error(0)
}

func (m *MockConsensus) GetInfo(verbose bool) (*raft.Info, error) {
	args := m.Called(verbose)

	return args.Get(0).(*raft.Info), args.Error(1)
}

func (m *MockConsensus) Get(key string) (string, bool) {
	args := m.Called(key)

	return args.String(0), args.Bool(1)
}

func (m *MockConsensus) Delete(key string) error {
	args := m.Called(key)

	return args.Error(0)
}

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

	mockConsensus := new(MockConsensus)
	mockConsensus.On("IsLive").Return(true)

	f := &Fiber{
		co:     mockConsensus,
		logger: log.With().Logger(),
	}

	ctx := new(fiber.Ctx)
	isLive := f.IsLive(ctx)

	assert.True(t, isLive)
	mockConsensus.AssertCalled(t, "IsLive")
}

func TestFiber_IsReady(t *testing.T) {
	t.Parallel()

	mockConsensus := new(MockConsensus)
	mockConsensus.On("IsReady").Return(true)

	f := &Fiber{
		co:     mockConsensus,
		logger: log.With().Logger(),
	}

	ctx := new(fiber.Ctx)
	isReady := f.IsReady(ctx)

	assert.True(t, isReady)
	mockConsensus.AssertCalled(t, "IsReady")
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

		mockConsensus := new(MockConsensus)
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
			co:     mockConsensus,
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

func TestFiber_WatchLeadershipChanges(t *testing.T) {
	t.Parallel()

	t.Run("shuts down when done channel is closed", func(t *testing.T) {
		t.Parallel()

		mockConsensus := new(MockConsensus)
		mockLogger := log.With().Logger()

		f := &Fiber{
			co:                  mockConsensus,
			logger:              mockLogger,
			leadershipChangesCh: make(raft.LeadershipChangesCh, 1),
		}

		done := make(chan struct{})
		close(done)

		assert.NotPanics(t, func() {
			f.WatchLeadershipChanges(done)
		}, "WatchLeadershipChanges should exit gracefully when done channel is closed")
	})

	t.Run("handles leadership changes", func(t *testing.T) {
		t.Parallel()

		mockConsensus := new(MockConsensus)
		mockConsensus.On("Set", LeaderAPIAddrKey, "advertise_address").Return(nil)

		mockLogger := log.With().Logger()

		f := &Fiber{
			co:                  mockConsensus,
			logger:              mockLogger,
			leadershipChangesCh: make(raft.LeadershipChangesCh, 1),
			config: &Config{
				Advertise: "advertise_address",
			},
		}

		done := make(chan struct{})
		go func() {
			time.Sleep(500 * time.Millisecond)
			close(done)
		}()

		go func() {
			f.leadershipChangesCh <- true
		}()

		assert.NotPanics(t, func() {
			f.WatchLeadershipChanges(done)
		}, "WatchLeadershipChanges should handle leadership changes without panicking")

		mockConsensus.AssertCalled(t, "Set", LeaderAPIAddrKey, "advertise_address")
	})

	t.Run("logs fatal error on Set failure", func(t *testing.T) {
		t.Parallel()

		mockConsensus := new(MockConsensus)
		mockConsensus.On("Set", LeaderAPIAddrKey, "advertise_address").Return(errors.New("set error"))

		mockLogger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})

		f := &Fiber{
			co:                  mockConsensus,
			logger:              mockLogger,
			leadershipChangesCh: make(raft.LeadershipChangesCh, 1),
			config: &Config{
				Advertise: "advertise_address",
			},
		}

		done := make(chan struct{})
		go func() {
			time.Sleep(500 * time.Millisecond)
			close(done)
		}()

		go func() {
			f.leadershipChangesCh <- true
		}()

		assert.Panics(t, func() {
			f.WatchLeadershipChanges(done)
		}, "WatchLeadershipChanges should not panic even if Set fails")

		mockConsensus.AssertCalled(t, "Set", LeaderAPIAddrKey, "advertise_address")
	})
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("successful initialization", func(t *testing.T) {
		t.Parallel()

		mockConsensus := new(MockConsensus)
		mockSentry := new(MockSentry)

		mockConsensus.On("SubscribeOnLeadershipChanges", mock.Anything).Return()
		mockSentry.On("Recover").Return()

		config := &Config{
			Addr:            "127.0.0.1:8080",
			Advertise:       "127.0.0.1:8080",
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     10 * time.Second,
			ShutdownTimeout: 5 * time.Second,
		}

		var f *Fiber

		assert.NotPanics(t, func() {
			f = New(config, mockConsensus, mockSentry)
		}, "New should not panic when initializing Fiber")

		assert.NotNil(t, f)
		assert.Equal(t, config, f.config)
		assert.Equal(t, mockConsensus, f.co)
		assert.Equal(t, mockSentry, f.sentry)
		assert.NotNil(t, f.app)
		assert.NotNil(t, f.leadershipChangesCh)

		mockConsensus.AssertCalled(t, "SubscribeOnLeadershipChanges", mock.Anything)
	})
}
