package agent

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	loggingUtils "github.com/weastur/maf/pkg/utils/logging"

	SYS "syscall"

	DEATH "github.com/vrecan/death/v3"
)

type Agent interface {
	Run() error
	IsLive(c *fiber.Ctx) bool
	IsReady(c *fiber.Ctx) bool
}

type agent struct {
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

var agentInstance Agent

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
) Agent {
	if agentInstance == nil {
		agentInstance = &agent{
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

	return agentInstance
}

func (a *agent) IsLive(_ *fiber.Ctx) bool {
	log.Trace().Msg("Live check called")

	return true
}

func (a *agent) IsReady(_ *fiber.Ctx) bool {
	log.Trace().Msg("Ready check called")

	return true
}

func (a *agent) Run() error {
	if err := loggingUtils.ConfigureLogging(a.logLevel, a.logPretty); err != nil {
		return fmt.Errorf("failed to configure logging: %w", err)
	}

	death := DEATH.NewDeath(SYS.SIGINT, SYS.SIGTERM)
	wg := sync.WaitGroup{}

	a.configureFiberApp()
	a.runFiberApp(&wg)

	death.WaitForDeathWithFunc(func() {
		log.Trace().Msg("Death callback called")

		a.shutdownFiberApp()

		log.Trace().Msg("Waiting for all goroutines to finish")
		wg.Wait()
	})

	return nil
}
