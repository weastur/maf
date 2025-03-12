package validate

import (
	"errors"

	"github.com/spf13/viper"
)

type LogLevel struct{}

var ErrLogLevel = errors.New(
	"log level misconfiguration. Log level must be one of: trace, debug, info, warn, error, fatal, panic",
)

func NewLogLevel() *LogLevel {
	return &LogLevel{}
}

func (v *LogLevel) Validate(viperInstance *viper.Viper) error {
	for _, key := range []string{"agent.log.level", "server.log.level"} {
		level := viperInstance.GetString(key)
		switch level {
		case "trace", "debug", "info", "warn", "error", "fatal", "panic":
		default:
			return ErrLogLevel
		}
	}

	return nil
}
