package agent

import (
	"fmt"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	loggingUtils "github.com/weastur/maf/pkg/utils/logging"
	sentryUtils "github.com/weastur/maf/pkg/utils/sentry"

	SYS "syscall"

	DEATH "github.com/vrecan/death/v3"
)

type Agent struct {
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

var (
	instance *Agent
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
) *Agent {
	once.Do(func() {
		instance = &Agent{
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
	})

	return instance
}

func (a *Agent) IsLive(_ *fiber.Ctx) bool {
	log.Trace().Msg("Live check called")

	return true
}

func (a *Agent) IsReady(_ *fiber.Ctx) bool {
	log.Trace().Msg("Ready check called")

	return true
}

func (a *Agent) Run() error {
	if err := sentryUtils.Init(a.sentryDSN); err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}
	defer sentryUtils.Recover(sentry.CurrentHub())

	if err := loggingUtils.Init(a.logLevel, a.logPretty); err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}

	death := DEATH.NewDeath(SYS.SIGINT, SYS.SIGTERM)
	wg := sync.WaitGroup{}

	a.initFiberApp()
	a.runFiberApp(&wg)

	death.WaitForDeathWithFunc(func() {
		log.Trace().Msg("Death callback called")

		a.shutdownFiberApp()
		sentryUtils.Flush()

		log.Trace().Msg("Waiting for all goroutines to finish")
		wg.Wait()
	})

	return nil
}
