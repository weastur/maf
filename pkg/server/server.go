package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	loggingUtils "github.com/weastur/maf/pkg/utils/logging"

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
		}
	}

	return serverInstance
}

func (s *server) IsLive(_ *fiber.Ctx) bool {
	return true
}

func (s *server) IsReady(_ *fiber.Ctx) bool {
	return true
}

func (s *server) Run() error {
	if err := loggingUtils.ConfigureLogging(s.logLevel, s.logPretty); err != nil {
		return fmt.Errorf("failed to configure logging: %w", err)
	}

	death := DEATH.NewDeath(SYS.SIGINT, SYS.SIGTERM)
	wg := sync.WaitGroup{}

	s.configureFiberApp()
	s.runFiberApp(&wg)

	death.WaitForDeathWithFunc(func() {
		s.shutdownFiberApp()

		wg.Wait()
	})

	return nil
}
