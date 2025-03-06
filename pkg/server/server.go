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

type Server interface {
	Run() error
	IsLive(c *fiber.Ctx) bool
	IsReady(c *fiber.Ctx) bool
}

type server struct {
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
}

var serverInstance Server

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
) Server {
	if serverInstance == nil {
		serverInstance = &server{
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
		}
	}

	return serverInstance
}

func (s *server) IsLive(_ *fiber.Ctx) bool {
	log.Trace().Msg("Live check called")

	return true
}

func (s *server) IsReady(_ *fiber.Ctx) bool {
	log.Trace().Msg("Ready check called")

	return true
}

func (s *server) Run() error {
	if err := loggingUtils.ConfigureLogging(s.logLevel, s.logPretty); err != nil {
		return fmt.Errorf("failed to configure logging: %w", err)
	}

	death := DEATH.NewDeath(SYS.SIGINT, SYS.SIGTERM)
	wg := sync.WaitGroup{}

	sentryUtils.ConfigureSentry(s.sentryDSN)
	defer sentryUtils.RecoverForSentry(sentry.CurrentHub())
	s.configureFiberApp()
	s.runFiberApp(&wg)

	death.WaitForDeathWithFunc(func() {
		log.Trace().Msg("Death callback called")

		s.shutdownFiberApp()
		sentryUtils.FlushSentry()

		log.Trace().Msg("Waiting for all goroutines to finish")
		wg.Wait()
	})

	return nil
}
