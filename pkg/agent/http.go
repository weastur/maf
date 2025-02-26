package agent

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/weastur/maf/pkg/utils"
	httpUtils "github.com/weastur/maf/pkg/utils/http"
)

func (a *agent) configureFiberApp() {
	a.fiberApp = fiber.New(
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
	a.fiberApp.Use(compress.New())
	a.fiberApp.Hooks().OnShutdown(func() error {
		fmt.Println("Shutting down agent handler")

		return nil
	})

	api := httpUtils.APIGroup(a.fiberApp)

	v1alpha := httpUtils.APIVersionGroup(api, "v1alpha")

	v1alpha.Get("/version", httpUtils.VersionHandler)
}

func (a *agent) runFiberApp(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := httpUtils.Listen(a.fiberApp, a.addr, a.certFile, a.keyFile, a.clientCertFile); err != nil {
			fmt.Println(err)
		}
	}()
}

func (a *agent) shutdownFiberApp() {
	if err := a.fiberApp.ShutdownWithTimeout(utils.AppShutdownTimeout); err != nil {
		fmt.Println(err)
	}
}
