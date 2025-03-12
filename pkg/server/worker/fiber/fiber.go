package fiber

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/weastur/maf/pkg/server/worker/fiber/http/api/v1alpha"
	"github.com/weastur/maf/pkg/utils"
	httpUtils "github.com/weastur/maf/pkg/utils/http"
	sentryUtils "github.com/weastur/maf/pkg/utils/sentry"
)

type API interface {
	Init(topRouter fiber.Router)
}

type ElectionCoordinator interface {
	IsLeader() bool
}

type Config struct {
	Addr            string
	CertFile        string
	KeyFile         string
	ClientCertFile  string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type Fiber struct {
	config *Config
	app    *fiber.App
	ec     ElectionCoordinator
}

func New(config *Config, ec ElectionCoordinator) *Fiber {
	log.Trace().Msg("Configuring fiber worker")

	f := &Fiber{config: config, ec: ec}

	f.app = fiber.New(
		fiber.Config{
			AppName:               "maf-server " + utils.AppVersion(),
			ServerHeader:          "maf-server/" + utils.AppVersion(),
			RequestMethods:        []string{fiber.MethodGet, fiber.MethodHead},
			ReadTimeout:           f.config.ReadTimeout,
			WriteTimeout:          f.config.WriteTimeout,
			IdleTimeout:           f.config.IdleTimeout,
			DisableStartupMessage: true,
			ErrorHandler:          httpUtils.ErrorHandler,
		},
	)
	httpUtils.AttachGenericMiddlewares(f.app, f)
	f.app.Hooks().OnShutdown(func() error {
		log.Info().Msg("Shutting down server handler")

		return nil
	})

	api := httpUtils.APIGroup(f.app)

	var v1AlphaInstance API = v1alpha.Get()

	v1AlphaInstance.Init(api)

	return f
}

func (f *Fiber) IsLive(_ *fiber.Ctx) bool {
	log.Trace().Msg("Live check called")

	return true
}

func (f *Fiber) IsReady(_ *fiber.Ctx) bool {
	log.Trace().Msg("Ready check called")

	return true
}

func (f *Fiber) Run(wg *sync.WaitGroup) {
	log.Info().Msg("Running fiber worker")

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer sentryUtils.Recover(sentryUtils.Fork("fiber"))

		if err := httpUtils.Listen(
			f.app, f.config.Addr, f.config.CertFile, f.config.KeyFile, f.config.ClientCertFile,
		); err != nil {
			log.Error().Err(err).Msg("failed to listen")
		}
	}()
}

func (f *Fiber) Stop() {
	log.Info().Msg("Stopping fiber worker")

	if err := f.app.ShutdownWithTimeout(f.config.ShutdownTimeout); err != nil {
		log.Error().Err(err).Msg("failed to shutdown fiber app")
	}
}
