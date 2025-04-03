package logging

import (
	"fmt"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentryzerolog "github.com/getsentry/sentry-go/zerolog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	utils "github.com/weastur/maf/internal/utils"
	sentryWrapper "github.com/weastur/maf/internal/utils/sentry"
)

const ComponentCtxKey = "component"

type Sentry interface {
	IsConfigured() bool
	Fork(scopeTag string) *sentryWrapper.Wrapper
	GetHub() *sentry.Hub
}

func Init(level string, pretty bool, sentry Sentry) error {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}

	if sentry.IsConfigured() {
		var multiLevelWriter zerolog.LevelWriter

		sentryWriter, err := sentryzerolog.NewWithHub(
			sentry.Fork("zerolog").GetHub(),
			sentryzerolog.Options{
				Levels:          []zerolog.Level{zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel},
				WithBreadcrumbs: true,
				FlushTimeout:    utils.SentryFlushTimeout,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to create sentry zerolog writer: %w", err)
		}

		if pretty {
			multiLevelWriter = zerolog.MultiLevelWriter(consoleWriter, sentryWriter)
		} else {
			multiLevelWriter = zerolog.MultiLevelWriter(os.Stderr, sentryWriter)
		}

		log.Logger = zerolog.New(multiLevelWriter).With().Timestamp().Logger()
	} else if pretty {
		log.Logger = log.Output(consoleWriter)
	}

	log.Logger = log.With().Str(ComponentCtxKey, "core").Logger()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.DurationFieldInteger = true
	zerolog.FloatingPointPrecision = 3
	zerolog.DisableSampling(true)

	zLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %w", err)
	}

	zerolog.SetGlobalLevel(zLevel)
	log.Trace().Msgf("logging level set to %s", zLevel.String())

	return nil
}
