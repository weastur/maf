package server

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/weastur/maf/pkg/server/http/api/v1alpha"
	"github.com/weastur/maf/pkg/utils"
	httpUtils "github.com/weastur/maf/pkg/utils/http"
	sentryUtils "github.com/weastur/maf/pkg/utils/sentry"
)

type API interface {
	Init(topRouter fiber.Router)
}

func (s *Server) initFiberApp() {
	log.Trace().Msg("Configuring fiber app")

	s.fiberApp = fiber.New(
		fiber.Config{
			AppName:               "maf-server " + utils.AppVersion(),
			ServerHeader:          "maf-server/" + utils.AppVersion(),
			RequestMethods:        []string{fiber.MethodGet, fiber.MethodHead},
			ReadTimeout:           s.httpReadTimeout,
			WriteTimeout:          s.httpWriteTimeout,
			IdleTimeout:           s.httpIdleTimeout,
			DisableStartupMessage: true,
			ErrorHandler:          httpUtils.ErrorHandler,
		},
	)
	httpUtils.AttachGenericMiddlewares(s.fiberApp, s)
	s.fiberApp.Hooks().OnShutdown(func() error {
		log.Info().Msg("Shutting down server handler")

		return nil
	})

	api := httpUtils.APIGroup(s.fiberApp)

	var v1AlphaInstance API = v1alpha.Get()

	v1AlphaInstance.Init(api)
}

func (s *Server) runFiberApp(wg *sync.WaitGroup) {
	log.Trace().Msg("Running fiber app")

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer sentryUtils.Recover(sentryUtils.Fork("fiber"))

		if err := httpUtils.Listen(s.fiberApp, s.addr, s.certFile, s.keyFile, s.clientCertFile); err != nil {
			log.Error().Err(err).Msg("failed to listen")
		}
	}()
}

func (s *Server) shutdownFiberApp() {
	log.Trace().Msg("Shutting down fiber app")

	if err := s.fiberApp.ShutdownWithTimeout(utils.AppShutdownTimeout); err != nil {
		log.Error().Err(err).Msg("failed to shutdown fiber app")
	}
}
