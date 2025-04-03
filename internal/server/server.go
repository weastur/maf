package server

import (
	"fmt"
	"slices"
	"sync"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"

	"github.com/weastur/maf/internal/server/worker/fiber"
	"github.com/weastur/maf/internal/server/worker/raft"

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

type Server struct {
	fiberConfig *fiber.Config
	raftConfig  *raft.Config
	config      *Config
	sentry      Sentry
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
	var err error

	s.sentry, err = sentryWrapper.New(s.config.SentryDSN)
	if err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}
	defer s.sentry.Recover()

	if err := loggingUtils.Init(s.config.LogLevel, s.config.LogPretty, s.sentry); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	raftWorker := raft.New(s.raftConfig, s.sentry.Fork("raft"))
	fiberWorker := fiber.New(s.fiberConfig, raftWorker, s.sentry.Fork("fiber"))
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

		s.sentry.Flush()

		log.Trace().Msg("Waiting for all workers to stop")
		wg.Wait()
	})

	return nil
}
