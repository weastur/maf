package logging

import (
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func ConfigureLogging() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.DurationFieldInteger = true
	zerolog.FloatingPointPrecision = 3
	zerolog.TimestampFieldName = "t"
	zerolog.MessageFieldName = "m"
	zerolog.LevelFieldName = "l"
	zerolog.ErrorFieldName = "e"
	zerolog.CallerFieldName = "c"
	zerolog.ErrorStackFieldName = "s"
	zerolog.DisableSampling(true)

	log.Print("hello world")
}
