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

type Death interface {
	WaitForDeathWithFunc(f func())
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
	workers     []Worker
	death       Death
	wg          sync.WaitGroup
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
			death:       DEATH.NewDeath(SYS.SIGINT, SYS.SIGTERM),
			wg:          sync.WaitGroup{},
		}
	})

	return instance
}

func (s *Server) onDeath() {
	log.Trace().Msg("Death callback called")

	slices.Reverse(s.workers)

	for _, worker := range s.workers {
		worker.Stop()
	}

	s.sentry.Flush()

	log.Trace().Msg("Waiting for all workers to stop")
	s.wg.Wait()
}

func (s *Server) Init() error {
	var err error

	s.sentry, err = sentryWrapper.New(s.config.SentryDSN)
	if err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	if err := loggingUtils.Init(s.config.LogLevel, s.config.LogPretty, s.sentry); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	raftWorker := raft.New(s.raftConfig, s.sentry.Fork("raft"))
	fiberWorker := fiber.New(s.fiberConfig, raftWorker, s.sentry.Fork("fiber"))
	s.workers = []Worker{raftWorker, fiberWorker}

	return nil
}

func (s *Server) Run() {
	defer s.sentry.Recover()

	for _, worker := range s.workers {
		worker.Run(&s.wg)
	}

	s.death.WaitForDeathWithFunc(s.onDeath)
}
