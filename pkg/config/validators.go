package config

import (
	"errors"

	"github.com/spf13/viper"
)

type validator interface {
	Validate(v *viper.Viper) error
}

type validatorMutualTLSMisconfig struct{}

var errMutualTLSMisconfig = errors.New(
	"mTLS misconfiguration. Both cert-file and key-file must be set with the client-cert-file",
)

func (v *validatorMutualTLSMisconfig) Validate(viperInstance *viper.Viper) error {
	if viperInstance.IsSet("agent.client_cert_file") &&
		!(viperInstance.IsSet("agent.cert_file") && viperInstance.IsSet("agent.key_file")) {
		return errMutualTLSMisconfig
	}

	return nil
}
