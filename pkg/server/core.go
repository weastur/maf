package server

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	SYS "syscall"

	DEATH "github.com/vrecan/death/v3"
)

type Server interface {
	Run() error
}

type server struct {
	addr             string
	certFile         string
	keyFile          string
	clientCertFile   string
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
			httpReadTimeout:  httpReadTimeout,
			httpWriteTimeout: httpWriteTimeout,
			httpIdleTimeout:  httpIdleTimeout,
		}
	}

	return serverInstance
}

func (s *server) Run() error {
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
