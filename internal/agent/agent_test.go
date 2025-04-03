package agent

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
	"github.com/weastur/maf/internal/agent/worker/fiber"
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
	fiberConfig := &fiber.Config{}

	agentInstance := Get(config, fiberConfig)

	assert.NotNil(t, agentInstance)

	assert.Equal(t, config, agentInstance.config)
	assert.Equal(t, fiberConfig, agentInstance.fiberConfig)

	secondInstance := Get(nil, nil)

	assert.Equal(t, agentInstance, secondInstance)
}

func TestOnDeath(t *testing.T) {
	t.Parallel()

	mockSentry := &MockSentry{}
	mockWorker1 := &MockWorker{}
	mockWorker2 := &MockWorker{}

	agent := &Agent{
		sentry:  mockSentry,
		workers: []Worker{mockWorker1, mockWorker2},
		wg:      sync.WaitGroup{},
	}

	mockWorker1.On("Stop").Return()
	mockWorker2.On("Stop").Return()
	mockSentry.On("Flush").Return()

	agent.onDeath()

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

	agent := &Agent{
		sentry:  mockSentry,
		workers: []Worker{mockWorker1, mockWorker2},
		death:   mockDeath,
		wg:      sync.WaitGroup{},
	}

	mockWorker1.On("Run", &agent.wg).Return()
	mockWorker2.On("Run", &agent.wg).Return()
	mockSentry.On("Recover").Return()
	mockDeath.On("WaitForDeathWithFunc", mock.Anything).Return()

	agent.Run()

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
	fiberConfig := &fiber.Config{}

	agent := &Agent{
		config:      config,
		fiberConfig: fiberConfig,
	}

	err := agent.Init()

	require.NoError(t, err)
	assert.Len(t, agent.workers, 1)
}
