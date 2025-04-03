package sentry

import (
	"fmt"

	sentry "github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"
	utils "github.com/weastur/maf/internal/utils"
)

type Wrapper struct {
	Hub *sentry.Hub
}

func New(dsn string) (*Wrapper, error) {
	if dsn == "" {
		return &Wrapper{}, nil
	}

	hub := sentry.NewHub(nil, sentry.NewScope())

	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:              dsn,
		Debug:            false,
		AttachStacktrace: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize sentry: %w", err)
	}

	hub.BindClient(client)
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(utils.SentryScopeTag, "main")
	})

	return &Wrapper{Hub: hub}, nil
}

func (w *Wrapper) GetHub() *sentry.Hub {
	if !w.IsConfigured() {
		log.Trace().Msg("Sentry is not configured, skipping getting hub")

		return nil
	}

	return w.Hub
}

func (w *Wrapper) Flush() {
	if !w.IsConfigured() {
		log.Trace().Msg("Sentry is not configured, skipping flushing")

		return
	}

	log.Trace().Msg("Flushing sentry")

	w.Hub.Flush(utils.SentryFlushTimeout)
}

func (w *Wrapper) Recover() {
	if !w.IsConfigured() {
		log.Trace().Msg("Sentry is not configured, skipping recovery")

		return
	}

	err := recover()

	log.Trace().Msg("Recovering from panic to send event to sentry (if enabled)")

	if err != nil {
		w.Hub.Recover(err)
		w.Flush()

		log.Trace().Msg("Repanic from panic to die")

		panic(err)
	}
}

func (w *Wrapper) Fork(scopeTag string) *Wrapper {
	if !w.IsConfigured() {
		log.Trace().Msg("Sentry is not configured, skipping fork")

		// As soon as sentry is not configured, we don't need to multiply hubs
		return w
	}

	localHub := w.Hub.Clone()
	localHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(utils.SentryScopeTag, scopeTag)
	})

	return &Wrapper{Hub: localHub}
}

func (w *Wrapper) IsConfigured() bool {
	return w.Hub != nil
}
