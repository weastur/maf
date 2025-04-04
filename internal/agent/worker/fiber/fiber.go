package fiber

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/weastur/maf/internal/agent/worker/fiber/http/api/v1alpha"
	"github.com/weastur/maf/internal/utils"
	httpUtils "github.com/weastur/maf/internal/utils/http"
	"github.com/weastur/maf/internal/utils/logging"
)

type API interface {
	Init(topRouter fiber.Router, logger zerolog.Logger)
}

type Sentry interface {
	Recover()
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
	logger zerolog.Logger
	sentry Sentry
}

func New(config *Config, sentry Sentry) *Fiber {
	log.Trace().Msg("Configuring fiber worker")

	f := &Fiber{
		config: config,
		logger: log.With().Str(logging.ComponentCtxKey, "fiber").Logger(),
		sentry: sentry,
	}

	f.app = fiber.New(
		fiber.Config{
			AppName:               "maf-agent " + utils.AppVersion(),
			ServerHeader:          "maf-agent/" + utils.AppVersion(),
			RequestMethods:        []string{fiber.MethodGet, fiber.MethodHead},
			ReadTimeout:           f.config.ReadTimeout,
			WriteTimeout:          f.config.WriteTimeout,
			IdleTimeout:           f.config.IdleTimeout,
			DisableStartupMessage: true,
			ErrorHandler:          httpUtils.ErrorHandler,
		},
	)
	httpUtils.AttachGenericMiddlewares(f.app, f.logger, f)
	f.app.Hooks().OnShutdown(func() error {
		f.logger.Info().Msg("Shutting down agent handler")

		return nil
	})

	api := httpUtils.APIGroup(f.app)

	var v1AlphaInstance API = v1alpha.Get()

	v1AlphaInstance.Init(api, f.logger)

	return f
}

func (f *Fiber) IsLive(_ *fiber.Ctx) bool {
	f.logger.Trace().Msg("Live check called")

	return true
}

func (f *Fiber) IsReady(_ *fiber.Ctx) bool {
	f.logger.Trace().Msg("Ready check called")

	return true
}

func (f *Fiber) Run(wg *sync.WaitGroup) {
	f.logger.Info().Msg("Running")

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer f.sentry.Recover()

		if err := httpUtils.Listen(
			f.app, f.logger, f.config.Addr, f.config.CertFile, f.config.KeyFile, f.config.ClientCertFile,
		); err != nil {
			f.logger.Error().Err(err).Msg("failed to listen")
		}
	}()
}

func (f *Fiber) Stop() {
	f.logger.Info().Msg("Stopping")

	if err := f.app.ShutdownWithTimeout(f.config.ShutdownTimeout); err != nil {
		f.logger.Error().Err(err).Msg("failed to shutdown fiber app")
	}
}
