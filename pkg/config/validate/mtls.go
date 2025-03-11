package validate

import (
	"errors"

	"github.com/spf13/viper"
)

type ValidatorMutualTLSMisconfig struct{}

var ErrMutualTLSMisconfig = errors.New(
	"mTLS misconfiguration. Both cert-file and key-file must be set with the client-cert-file",
)

func NewValidatorMutualTLSMisconfig() *ValidatorMutualTLSMisconfig {
	return &ValidatorMutualTLSMisconfig{}
}

func (v *ValidatorMutualTLSMisconfig) Validate(viperInstance *viper.Viper) error {
	for _, key := range []string{"agent", "server"} {
		if viperInstance.IsSet(key+".client_cert_file") &&
			!(viperInstance.IsSet(key+".cert_file") && viperInstance.IsSet(key+".key_file")) {
			return ErrMutualTLSMisconfig
		}
	}

	return nil
}
