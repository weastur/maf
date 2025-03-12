package validate

import (
	"errors"

	"github.com/spf13/viper"
)

type MutualTLS struct{}

var ErrMutualTLS = errors.New(
	"both cert-file and key-file must be set with the client-cert-file",
)

func NewMutualTLS() *MutualTLS {
	return &MutualTLS{}
}

func (v *MutualTLS) Validate(viperInstance *viper.Viper) error {
	for _, key := range []string{"agent", "server"} {
		if viperInstance.IsSet(key+".http.client_cert_file") &&
			!(viperInstance.IsSet(key+".http.cert_file") && viperInstance.IsSet(key+".http.key_file")) {
			return ErrMutualTLS
		}
	}

	return nil
}
