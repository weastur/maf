package logging

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func ConfigureLogging(level string) error {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.DurationFieldInteger = true
	zerolog.FloatingPointPrecision = 3
	zerolog.TimestampFieldName = "t"
	zerolog.MessageFieldName = "m"
	zerolog.LevelFieldName = "l"
	zerolog.ErrorFieldName = "e"
	zerolog.DisableSampling(true)

	zLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %w", err)
	}

	zerolog.SetGlobalLevel(zLevel)
	log.Trace().Msgf("logging level set to %s", zLevel.String())

	return nil
}
