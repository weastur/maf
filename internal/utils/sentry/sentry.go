package sentry

import (
	"fmt"

	sentry "github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"
	utils "github.com/weastur/maf/internal/utils"
)

func Init(dsn string) error {
	if dsn == "" {
		return nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Debug:            false,
		AttachStacktrace: true,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize sentry: %w", err)
	}

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(utils.SentryScopeTag, "main")
	})

	return nil
}

func Flush() {
	if !IsConfgured() {
		log.Trace().Msg("Sentry is not configured, skipping flushing")

		return
	}

	log.Trace().Msg("Flushing sentry")

	sentry.Flush(utils.SentryFlushTtimeout)
}

func Recover(hub *sentry.Hub) {
	if !IsConfgured() {
		log.Trace().Msg("Sentry is not configured, skipping recovery")

		return
	}

	err := recover()

	log.Trace().Msg("Recovering from panic to send event to sentry (if enabled)")

	if err != nil {
		hub.Recover(err)
		Flush()

		log.Trace().Msg("Repanic from panic to die")

		panic(err)
	}
}

func Fork(scopeTag string) *sentry.Hub {
	if !IsConfgured() {
		log.Trace().Msg("Sentry is not configured, skipping fork")

		// As soon as sentry is not configured, we don't need to multiply hubs
		return sentry.CurrentHub()
	}

	localHub := sentry.CurrentHub().Clone()
	localHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(utils.SentryScopeTag, scopeTag)
	})

	return localHub
}

func IsConfgured() bool {
	return sentry.CurrentHub().Client() != nil
}
