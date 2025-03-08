package agent

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/weastur/maf/pkg/agent/http/api/v1alpha"
	"github.com/weastur/maf/pkg/utils"
	httpUtils "github.com/weastur/maf/pkg/utils/http"
	sentryUtils "github.com/weastur/maf/pkg/utils/sentry"
)

func (a *agent) configureFiberApp() {
	log.Trace().Msg("Configuring fiber app")

	a.fiberApp = fiber.New(
		fiber.Config{
			AppName:               "maf-agent " + utils.AppVersion(),
			ServerHeader:          "maf-agent/" + utils.AppVersion(),
			RequestMethods:        []string{fiber.MethodGet, fiber.MethodHead},
			ReadTimeout:           a.httpReadTimeout,
			WriteTimeout:          a.httpWriteTimeout,
			IdleTimeout:           a.httpIdleTimeout,
			DisableStartupMessage: true,
			ErrorHandler:          httpUtils.ErrorHandler,
		},
	)
	httpUtils.AttachGenericMiddlewares(a.fiberApp, a)
	a.fiberApp.Hooks().OnShutdown(func() error {
		log.Info().Msg("Shutting down agent handler")

		return nil
	})

	api := httpUtils.APIGroup(a.fiberApp)

	v1alpha.Get().Router(api)
}

func (a *agent) runFiberApp(wg *sync.WaitGroup) {
	log.Trace().Msg("Running fiber app")

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer sentryUtils.Recover(sentryUtils.Fork("fiber"))

		if err := httpUtils.Listen(a.fiberApp, a.addr, a.certFile, a.keyFile, a.clientCertFile); err != nil {
			log.Error().Err(err).Msg("failed to listen")
		}
	}()
}

func (a *agent) shutdownFiberApp() {
	log.Trace().Msg("Shutting down fiber app")

	if err := a.fiberApp.ShutdownWithTimeout(utils.AppShutdownTimeout); err != nil {
		log.Error().Err(err).Msg("failed to shutdown fiber app")
	}
}
