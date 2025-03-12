package server

import (
	"fmt"
	"sync"
	"time"

	sentry "github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	loggingUtils "github.com/weastur/maf/pkg/utils/logging"
	sentryUtils "github.com/weastur/maf/pkg/utils/sentry"

	SYS "syscall"

	DEATH "github.com/vrecan/death/v3"
)

type Server struct {
	addr             string
	certFile         string
	keyFile          string
	clientCertFile   string
	logLevel         string
	logPretty        bool
	httpReadTimeout  time.Duration
	httpWriteTimeout time.Duration
	httpIdleTimeout  time.Duration
	fiberApp         *fiber.App
	sentryDSN        string
	raftAddr         string
	raftNodeID       string
	raftDevmode      bool
	raftPeers        []string
}

var (
	instance *Server
	once     sync.Once
)

func Get(
	addr string,
	certFile string,
	keyFile string,
	clientCertFile string,
	logLevel string,
	logPretty bool,
	httpReadTimeout time.Duration,
	httpWriteTimeout time.Duration,
	httpIdleTimeout time.Duration,
	sentryDSN string,
	raftAddr string,
	raftNodeID string,
	raftDevmode bool,
	raftPeers []string,
) *Server {
	once.Do(func() {
		instance = &Server{
			addr:             addr,
			certFile:         certFile,
			keyFile:          keyFile,
			clientCertFile:   clientCertFile,
			logLevel:         logLevel,
			logPretty:        logPretty,
			httpReadTimeout:  httpReadTimeout,
			httpWriteTimeout: httpWriteTimeout,
			httpIdleTimeout:  httpIdleTimeout,
			sentryDSN:        sentryDSN,
			raftAddr:         raftAddr,
			raftNodeID:       raftNodeID,
			raftDevmode:      raftDevmode,
			raftPeers:        raftPeers,
		}
	})

	return instance
}

func (s *Server) IsLive(_ *fiber.Ctx) bool {
	log.Trace().Msg("Live check called")

	return true
}

func (s *Server) IsReady(_ *fiber.Ctx) bool {
	log.Trace().Msg("Ready check called")

	return true
}

func (s *Server) Run() error {
	if err := sentryUtils.Init(s.sentryDSN); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}
	defer sentryUtils.Recover(sentry.CurrentHub())

	if err := loggingUtils.Init(s.logLevel, s.logPretty); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	death := DEATH.NewDeath(SYS.SIGINT, SYS.SIGTERM)
	wg := sync.WaitGroup{}

	s.initFiberApp()
	s.runFiberApp(&wg)

	death.WaitForDeathWithFunc(func() {
		log.Trace().Msg("Death callback called")

		s.shutdownFiberApp()
		sentryUtils.Flush()

		log.Trace().Msg("Waiting for all goroutines to finish")
		wg.Wait()
	})

	return nil
}
