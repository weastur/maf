package fiber

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/weastur/maf/internal/server/worker/fiber/http/api/v1alpha"
	"github.com/weastur/maf/internal/server/worker/raft"
	"github.com/weastur/maf/internal/utils"
	httpUtils "github.com/weastur/maf/internal/utils/http"
	"github.com/weastur/maf/internal/utils/logging"
	sentryUtils "github.com/weastur/maf/internal/utils/sentry"
)

const LeaderAPIAddrKey = "leaderAPIAddr"

type Consensus interface {
	IsReady() bool
	IsLive() bool
	Set(key, value string) error
	SubscribeOnLeadershipChanges(ch raft.LeadershipChangesCh)
}

type Config struct {
	Addr            string
	Advertise       string
	CertFile        string
	KeyFile         string
	ClientCertFile  string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type Fiber struct {
	config              *Config
	app                 *fiber.App
	co                  Consensus
	logger              zerolog.Logger
	leadershipChangesCh raft.LeadershipChangesCh
}

func New(config *Config, co Consensus) *Fiber {
	log.Trace().Msg("Configuring fiber worker")

	f := &Fiber{
		config:              config,
		co:                  co,
		logger:              log.With().Str(logging.ComponentCtxKey, "fiber").Logger(),
		leadershipChangesCh: make(raft.LeadershipChangesCh, 1),
	}

	f.app = fiber.New(
		fiber.Config{
			AppName:               "maf-server " + utils.AppVersion(),
			ServerHeader:          "maf-server/" + utils.AppVersion(),
			RequestMethods:        []string{fiber.MethodGet, fiber.MethodHead, fiber.MethodPost, fiber.MethodDelete},
			ReadTimeout:           f.config.ReadTimeout,
			WriteTimeout:          f.config.WriteTimeout,
			IdleTimeout:           f.config.IdleTimeout,
			DisableStartupMessage: true,
			ErrorHandler:          httpUtils.ErrorHandler,
		},
	)
	httpUtils.AttachGenericMiddlewares(f.app, f.logger, f)
	f.app.Hooks().OnShutdown(func() error {
		f.logger.Info().Msg("Shutting down server handler")

		return nil
	})

	co.SubscribeOnLeadershipChanges(f.leadershipChangesCh)

	api := httpUtils.APIGroup(f.app)

	v1alphaConsensus, ok := f.co.(v1alpha.Consensus)
	if !ok {
		f.logger.Fatal().Msg("Consensus does not implement v1alpha interface")
	}

	v1alpha.Get().Init(api, f.logger, v1alphaConsensus)

	return f
}

func (f *Fiber) IsLive(_ *fiber.Ctx) bool {
	f.logger.Trace().Msg("Live check called")

	return f.co.IsLive()
}

func (f *Fiber) IsReady(_ *fiber.Ctx) bool {
	f.logger.Trace().Msg("Ready check called")

	return f.co.IsReady()
}

func (f *Fiber) WatchLeadershipChanges(done <-chan struct{}) {
	f.logger.Info().Msg("Watching leadership changes")

	for {
		select {
		case <-done:
			f.logger.Info().Msg("Shutting down leadership changes watcher")

			return
		case isLeader := <-f.leadershipChangesCh:
			f.logger.Info().Msg("Leadership changes detected")

			if isLeader {
				if err := f.co.Set(LeaderAPIAddrKey, f.config.Advertise); err != nil {
					f.logger.Fatal().Err(err).Msg("failed to set leader API address, this should not happen")
				}
			}
		}
	}
}

func (f *Fiber) Run(wg *sync.WaitGroup) {
	f.logger.Info().Msg("Running")

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer sentryUtils.Recover(sentryUtils.Fork("fiber"))

		watcherDone := make(chan struct{})
		go f.WatchLeadershipChanges(watcherDone)
		defer close(watcherDone)

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
