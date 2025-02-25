package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/weastur/maf/pkg/utils"

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
	app := fiber.New(
		fiber.Config{
			AppName:               "maf-server " + utils.AppVersion(),
			ServerHeader:          "maf-server/" + utils.AppVersion(),
			RequestMethods:        []string{fiber.MethodGet, fiber.MethodHead},
			ReadTimeout:           s.httpReadTimeout,
			WriteTimeout:          s.httpWriteTimeout,
			IdleTimeout:           s.httpIdleTimeout,
			DisableStartupMessage: true,
		},
	)
	app.Hooks().OnShutdown(func() error {
		fmt.Println("Shutting down server handler")

		return nil
	})
	utils.ConfigureMetrics(app)

	death := DEATH.NewDeath(SYS.SIGINT, SYS.SIGTERM)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := utils.RunFiberApp(app, s.addr, s.certFile, s.keyFile, s.clientCertFile); err != nil {
			fmt.Println(err)
		}
	}()

	death.WaitForDeathWithFunc(func() {
		if err := app.ShutdownWithTimeout(utils.AppShutdownTimeout); err != nil {
			fmt.Println(err)
		}

		wg.Wait()
	})

	return nil
}
