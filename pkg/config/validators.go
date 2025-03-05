package config

import (
	"errors"

	"github.com/spf13/viper"
)

type validator interface {
	Validate(v *viper.Viper) error
}

type (
	validatorMutualTLSMisconfig struct{}
	validatorLogLevel           struct{}
)

var errMutualTLSMisconfig = errors.New(
	"mTLS misconfiguration. Both cert-file and key-file must be set with the client-cert-file",
)

var errLogLevelMisconfig = errors.New(
	"log level misconfiguration. Log level must be one of: trace, debug, info, warn, error, fatal, panic",
)

func (v *validatorMutualTLSMisconfig) Validate(viperInstance *viper.Viper) error {
	for _, key := range []string{"agent", "server"} {
		if viperInstance.IsSet(key+".client_cert_file") &&
			!(viperInstance.IsSet(key+".cert_file") && viperInstance.IsSet(key+".key_file")) {
			return errMutualTLSMisconfig
		}
	}

	return nil
}

func (v *validatorLogLevel) Validate(viperInstance *viper.Viper) error {
	for _, key := range []string{"agent.log.level", "server.log.level"} {
		level := viperInstance.GetString(key)
		switch level {
		case "trace", "debug", "info", "warn", "error", "fatal", "panic":
		default:
			return errLogLevelMisconfig
		}
	}

	return nil
}
