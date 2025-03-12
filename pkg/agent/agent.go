package agent

import (
	"fmt"
	"slices"
	"sync"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"

	"github.com/weastur/maf/pkg/agent/worker/fiber"

	loggingUtils "github.com/weastur/maf/pkg/utils/logging"
	sentryUtils "github.com/weastur/maf/pkg/utils/sentry"

	SYS "syscall"

	DEATH "github.com/vrecan/death/v3"
)

type Worker interface {
	Run(wg *sync.WaitGroup)
	Stop()
}

type Config struct {
	LogLevel  string
	LogPretty bool
	SentryDSN string
}

type Agent struct {
	config      *Config
	fiberConfig *fiber.Config
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
	if err := sentryUtils.Init(a.config.SentryDSN); err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}
	defer sentryUtils.Recover(sentry.CurrentHub())

	if err := loggingUtils.Init(a.config.LogLevel, a.config.LogPretty); err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}

	fiberWorker := fiber.New(a.fiberConfig)
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

		sentryUtils.Flush()

		log.Trace().Msg("Waiting for all goroutines to finish")
		wg.Wait()
	})

	return nil
}
