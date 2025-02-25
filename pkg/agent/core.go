package agent

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/weastur/maf/pkg/utils"
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
			RequestMethods:        []string{fiber.MethodGet},
			ReadTimeout:           a.httpReadTimeout,
			WriteTimeout:          a.httpWriteTimeout,
			IdleTimeout:           a.httpIdleTimeout,
			DisableStartupMessage: true,
		},
	)

	switch {
	case a.clientCertFile != "":
		if err := app.ListenMutualTLS(a.addr, a.certFile, a.keyFile, a.clientCertFile); err != nil {
			return fmt.Errorf("failed to listen with mutual TLS: %w", err)
		}
	case a.certFile != "" && a.keyFile != "":
		if err := app.ListenTLS(a.addr, a.certFile, a.keyFile); err != nil {
			return fmt.Errorf("failed to listen with TLS: %w", err)
		}
	default:
		if err := app.Listen(a.addr); err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}
	}

	return nil
}
