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

type Config struct {
	LogLevel  string
	LogPretty bool
	SentryDSN string
}

type Agent struct {
	config      *Config
	fiberConfig *fiber.Config
	sentry      Sentry
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
		}
	})

	return instance
}

func (a *Agent) Run() error {
	var err error

	a.sentry, err = sentryWrapper.New(a.config.SentryDSN)
	if err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}
	defer a.sentry.Recover()

	if err := loggingUtils.Init(a.config.LogLevel, a.config.LogPretty, a.sentry); err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}

	fiberWorker := fiber.New(a.fiberConfig, a.sentry.Fork("fiber"))
	workers := []Worker{fiberWorker}

	death := DEATH.NewDeath(SYS.SIGINT, SYS.SIGTERM)
	wg := sync.WaitGroup{}

	for _, worker := range workers {
		worker.Run(&wg)
	}

	death.WaitForDeathWithFunc(func() {
		log.Trace().Msg("Death callback called")

		slices.Reverse(workers)

		for _, worker := range workers {
			worker.Stop()
		}

		a.sentry.Flush()

		log.Trace().Msg("Waiting for all goroutines to finish")
		wg.Wait()
	})

	return nil
}
