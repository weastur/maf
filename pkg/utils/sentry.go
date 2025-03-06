package utils

import (
	sentry "github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"
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
}

func FlushSentry() {
	log.Trace().Msg("Flushing sentry")

	sentry.Flush(SentryFlushTtimeout)
}
