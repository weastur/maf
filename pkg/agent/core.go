package agent

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	SYS "syscall"

	DEATH "github.com/vrecan/death/v3"
)

type Agent interface {
	Run() error
}

type agent struct {
	addr             string
	certFile         string
	keyFile          string
	clientCertFile   string
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
			httpReadTimeout:  httpReadTimeout,
			httpWriteTimeout: httpWriteTimeout,
			httpIdleTimeout:  httpIdleTimeout,
		}
	}

	return agentInstance
}

func (a *agent) Run() error {
	death := DEATH.NewDeath(SYS.SIGINT, SYS.SIGTERM)
	wg := sync.WaitGroup{}

	a.configureFiberApp()
	a.runFiberApp(&wg)

	death.WaitForDeathWithFunc(func() {
		a.shutdownFiberApp()

		wg.Wait()
	})

	return nil
}
