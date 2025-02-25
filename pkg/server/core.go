package server

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/weastur/maf/pkg/utils"
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
			RequestMethods:        []string{fiber.MethodGet},
			ReadTimeout:           s.httpReadTimeout,
			WriteTimeout:          s.httpWriteTimeout,
			IdleTimeout:           s.httpIdleTimeout,
			DisableStartupMessage: true,
		},
	)

	switch {
	case s.clientCertFile != "":
		if err := app.ListenMutualTLS(s.addr, s.certFile, s.keyFile, s.clientCertFile); err != nil {
			return fmt.Errorf("failed to listen with mutual TLS: %w", err)
		}
	case s.certFile != "" && s.keyFile != "":
		if err := app.ListenTLS(s.addr, s.certFile, s.keyFile); err != nil {
			return fmt.Errorf("failed to listen with TLS: %w", err)
		}
	default:
		if err := app.Listen(s.addr); err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}
	}

	return nil
}
