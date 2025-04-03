package agent

import (
	"fmt"
	"slices"
	"sync"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"

	"github.com/weastur/maf/internal/agent/worker/fiber"

	loggingUtils "github.com/weastur/maf/internal/utils/logging"
	sentryWrapper "github.com/weastur/maf/internal/utils/sentry"

	SYS "syscall"

	DEATH "github.com/vrecan/death/v3"
)

type Worker interface {
	Run(wg *sync.WaitGroup)
	Stop()
}

type Sentry interface {
	Flush()
	Recover()
	IsConfigured() bool
	GetHub() *sentry.Hub
	Fork(scopeTag string) *sentryWrapper.Wrapper
}

type Death interface {
	WaitForDeathWithFunc(f func())
}

type Config struct {
	LogLevel  string
	LogPretty bool
	SentryDSN string
}

type Agent struct {
	config      *Config
	fiberConfig *fiber.Config
	sentry      Sentry
	workers     []Worker
	death       Death
	wg          sync.WaitGroup
}

var (
	instance *Agent
	once     sync.Once
)

func Get(
	config *Config,
	fiberConfig *fiber.Config,
) *Agent {
	once.Do(func() {
		instance = &Agent{
			config:      config,
			fiberConfig: fiberConfig,
			death:       DEATH.NewDeath(SYS.SIGINT, SYS.SIGTERM),
			wg:          sync.WaitGroup{},
		}
	})

	return instance
}

func (a *Agent) onDeath() {
	log.Trace().Msg("Death callback called")

	slices.Reverse(a.workers)

	for _, worker := range a.workers {
		worker.Stop()
	}

	a.sentry.Flush()

	log.Trace().Msg("Waiting for all workers to stop")
	a.wg.Wait()
}

func (a *Agent) Init() error {
	var err error

	a.sentry, err = sentryWrapper.New(a.config.SentryDSN)
	if err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}

	if err := loggingUtils.Init(a.config.LogLevel, a.config.LogPretty, a.sentry); err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}

	fiberWorker := fiber.New(a.fiberConfig, a.sentry.Fork("fiber"))
	a.workers = []Worker{fiberWorker}

	return nil
}

func (a *Agent) Run() {
	defer a.sentry.Recover()

	for _, worker := range a.workers {
		worker.Run(&a.wg)
	}

	a.death.WaitForDeathWithFunc(a.onDeath)
}
