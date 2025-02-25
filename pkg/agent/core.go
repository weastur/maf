package agent

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type Agent interface {
	Run() error
}

type agent struct {
	addr           string
	certFile       string
	keyFile        string
	clientCertFile string
}

var agentInstance Agent

func Get(addr, certFile, keyFile, clientCertFile string) Agent {
	if agentInstance == nil {
		agentInstance = &agent{
			addr:           addr,
			certFile:       certFile,
			keyFile:        keyFile,
			clientCertFile: clientCertFile,
		}
	}

	return agentInstance
}

func (a *agent) Run() error {
	app := fiber.New(
		fiber.Config{
			AppName:               "maf-agent",
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
