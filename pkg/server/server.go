package server

import (
	"fmt"
	"slices"
	"sync"

	sentry "github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"

	"github.com/weastur/maf/pkg/server/worker/fiber"
	"github.com/weastur/maf/pkg/server/worker/raft"

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

type Server struct {
	fiberConfig *fiber.Config
	raftConfig  *raft.Config
	config      *Config
}

var (
	instance *Server
	once     sync.Once
)

func Get(
	config *Config,
	raftConfig *raft.Config,
	fiberConfig *fiber.Config,
) *Server {
	once.Do(func() {
		instance = &Server{
			config:      config,
			fiberConfig: fiberConfig,
			raftConfig:  raftConfig,
		}
	})

	return instance
}

func (s *Server) Run() error {
	if err := sentryUtils.Init(s.config.SentryDSN); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}
	defer sentryUtils.Recover(sentry.CurrentHub())

	if err := loggingUtils.Init(s.config.LogLevel, s.config.LogPretty); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	raftWorker := raft.New(s.raftConfig)
	fiberWorker := fiber.New(s.fiberConfig, raftWorker)
	workers := []Worker{raftWorker, fiberWorker}

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

		log.Trace().Msg("Waiting for all workers to stop")
		wg.Wait()
	})

	return nil
}
