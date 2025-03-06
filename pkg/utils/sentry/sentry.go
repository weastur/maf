package sentry

import (
	sentry "github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"
	utils "github.com/weastur/maf/pkg/utils"
)

func ConfigureSentry(dsn string) {
	log.Trace().Msg("Configuring sentry")

	if dsn == "" {
		log.Debug().Msg("Sentry DSN is not set, skipping configuration")

		return
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Debug:            true,
		AttachStacktrace: true,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to configure sentry")
	}

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(utils.SentryScopeTag, "main")
	})
}

func FlushSentry() {
	log.Trace().Msg("Flushing sentry")

	sentry.Flush(utils.SentryFlushTtimeout)
}

func RecoverForSentry(hub *sentry.Hub) {
	err := recover()

	log.Trace().Msg("Recovering from panic to send event to sentry")

	if err != nil {
		hub.Recover(err)
		FlushSentry()
		panic(err)
	}
}

func ForkSentryHub(scopeTag string) *sentry.Hub {
	localHub := sentry.CurrentHub().Clone()
	localHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(utils.SentryScopeTag, scopeTag)
	})

	return localHub
}
