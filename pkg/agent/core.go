package agent

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/weastur/maf/pkg/utils"

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
	app := fiber.New(
		fiber.Config{
			AppName:               "maf-agent " + utils.AppVersion(),
			ServerHeader:          "maf-agent/" + utils.AppVersion(),
			RequestMethods:        []string{fiber.MethodGet, fiber.MethodHead},
			ReadTimeout:           a.httpReadTimeout,
			WriteTimeout:          a.httpWriteTimeout,
			IdleTimeout:           a.httpIdleTimeout,
			DisableStartupMessage: true,
		},
	)
	app.Hooks().OnShutdown(func() error {
		fmt.Println("Shutting down agent handler")

		return nil
	})
	app.Get("/version", utils.HTTPVersionHandler)
	utils.ConfigureMetrics(app)

	death := DEATH.NewDeath(SYS.SIGINT, SYS.SIGTERM)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := utils.RunFiberApp(app, a.addr, a.certFile, a.keyFile, a.clientCertFile); err != nil {
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
