package logging

import (
	"fmt"
	"os"
	"time"

	sentryzerolog "github.com/getsentry/sentry-go/zerolog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	utils "github.com/weastur/maf/pkg/utils"
	sentryUtils "github.com/weastur/maf/pkg/utils/sentry"
)

func Configure(level string, pretty bool) error {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}

	if sentryUtils.IsConfgured() {
		var multiLevelWriter zerolog.LevelWriter

		sentryWriter, err := sentryzerolog.NewWithHub(
			sentryUtils.Fork("zerolog"),
			sentryzerolog.Options{
				Levels:          []zerolog.Level{zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel},
				WithBreadcrumbs: true,
				FlushTimeout:    utils.SentryFlushTtimeout,
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
