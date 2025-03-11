package validate

import (
	"errors"

	"github.com/spf13/viper"
)

type ValidatorLogLevel struct{}

var ErrLogLevelMisconfig = errors.New(
	"log level misconfiguration. Log level must be one of: trace, debug, info, warn, error, fatal, panic",
)

func NewValidatorLogLevel() *ValidatorLogLevel {
	return &ValidatorLogLevel{}
}

func (v *ValidatorLogLevel) Validate(viperInstance *viper.Viper) error {
	for _, key := range []string{"agent.log.level", "server.log.level"} {
		level := viperInstance.GetString(key)
		switch level {
		case "trace", "debug", "info", "warn", "error", "fatal", "panic":
		default:
			return ErrLogLevelMisconfig
		}
	}

	return nil
}
