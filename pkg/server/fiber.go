package server

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/weastur/maf/pkg/server/http/api/v1alpha"
	"github.com/weastur/maf/pkg/utils"
	httpUtils "github.com/weastur/maf/pkg/utils/http"
)

func (s *server) configureFiberApp() {
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
		fmt.Println("Shutting down server handler") // TODO: logging

		return nil
	})

	api := httpUtils.APIGroup(s.fiberApp)

	v1alpha.Get().Router(api)
}

func (s *server) runFiberApp(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := httpUtils.Listen(s.fiberApp, s.addr, s.certFile, s.keyFile, s.clientCertFile); err != nil {
			fmt.Println(err) // TODO: logging
		}
	}()
}

func (s *server) shutdownFiberApp() {
	if err := s.fiberApp.ShutdownWithTimeout(utils.AppShutdownTimeout); err != nil {
		fmt.Println(err) // TODO: logging
	}
}
