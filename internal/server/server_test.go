package server

import (
	"os"
	"sync"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/weastur/maf/internal/server/worker/fiber"
	"github.com/weastur/maf/internal/server/worker/raft"
	sentryWrapper "github.com/weastur/maf/internal/utils/sentry"
)

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	log.Logger = log.Output(zerolog.Nop())

	os.Exit(m.Run())
}

type MockWorker struct {
	mock.Mock
}

func (m *MockWorker) Run(wg *sync.WaitGroup) {
	m.Called(wg)
}

func (m *MockWorker) Stop() {
	m.Called()
}

type MockSentry struct {
	mock.Mock
}

func (m *MockSentry) Flush() {
	m.Called()
}

func (m *MockSentry) Recover() {
	m.Called()
}

func (m *MockSentry) IsConfigured() bool {
	args := m.Called()

	return args.Bool(0)
}

func (m *MockSentry) GetHub() *sentry.Hub {
	args := m.Called()

	return args.Get(0).(*sentry.Hub)
}

func (m *MockSentry) Fork(scopeTag string) *sentryWrapper.Wrapper {
	args := m.Called(scopeTag)

	return args.Get(0).(*sentryWrapper.Wrapper)
}

type MockDeath struct {
	mock.Mock
}

func (m *MockDeath) WaitForDeathWithFunc(f func()) {
	m.Called(f)
}

func TestGet(t *testing.T) {
	t.Parallel()

	config := &Config{
		LogLevel:  "debug",
		LogPretty: true,
		SentryDSN: "",
	}
	raftConfig := &raft.Config{
		NodeID: "test-node",
	}
	fiberConfig := &fiber.Config{}

	serverInstance := Get(config, raftConfig, fiberConfig)

	assert.NotNil(t, serverInstance)

	assert.Equal(t, config, serverInstance.config)
	assert.Equal(t, raftConfig, serverInstance.raftConfig)
	assert.Equal(t, fiberConfig, serverInstance.fiberConfig)

	secondInstance := Get(nil, nil, nil)

	assert.Equal(t, serverInstance, secondInstance)
}

func TestOnDeath(t *testing.T) {
	t.Parallel()

	mockSentry := &MockSentry{}
	mockWorker1 := &MockWorker{}
	mockWorker2 := &MockWorker{}

	server := &Server{
		sentry:  mockSentry,
		workers: []Worker{mockWorker1, mockWorker2},
		wg:      sync.WaitGroup{},
	}

	mockWorker1.On("Stop").Return()
	mockWorker2.On("Stop").Return()
	mockSentry.On("Flush").Return()

	server.onDeath()

	mockSentry.AssertExpectations(t)
	mockWorker1.AssertExpectations(t)
	mockWorker2.AssertExpectations(t)
}

func TestRun(t *testing.T) {
	t.Parallel()

	mockSentry := &MockSentry{}
	mockWorker1 := &MockWorker{}
	mockWorker2 := &MockWorker{}
	mockDeath := &MockDeath{}

	server := &Server{
		sentry:  mockSentry,
		workers: []Worker{mockWorker1, mockWorker2},
		death:   mockDeath,
		wg:      sync.WaitGroup{},
	}

	mockWorker1.On("Run", &server.wg).Return()
	mockWorker2.On("Run", &server.wg).Return()
	mockSentry.On("Recover").Return()
	mockDeath.On("WaitForDeathWithFunc", mock.Anything).Return()

	server.Run()

	mockSentry.AssertExpectations(t)
	mockWorker1.AssertExpectations(t)
	mockWorker2.AssertExpectations(t)
	mockDeath.AssertExpectations(t)
}

func TestInit(t *testing.T) {
	config := &Config{
		LogLevel:  "debug",
		LogPretty: true,
		SentryDSN: "https://examplePublicKey@o0.ingest.sentry.io/0",
	}
	raftConfig := &raft.Config{
		NodeID: "test-node",
	}
	fiberConfig := &fiber.Config{}

	server := &Server{
		config:      config,
		fiberConfig: fiberConfig,
		raftConfig:  raftConfig,
	}

	err := server.Init()

	require.NoError(t, err)
	assert.Len(t, server.workers, 2)
}
