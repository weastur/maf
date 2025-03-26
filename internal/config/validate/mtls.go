package validate

import (
	"errors"

	"github.com/spf13/viper"
)

type MutualTLS struct{}

var ErrMutualTLS = errors.New(
	"both cert-file and key-file must be set with the client-cert-file/server-cert-file",
)

func NewMutualTLS() *MutualTLS {
	return &MutualTLS{}
}

func (v *MutualTLS) Validate(viperInstance *viper.Viper) error {
	for _, key := range []string{"agent.http", "server.http"} {
		if viperInstance.IsSet(key+".client_cert_file") &&
			(!viperInstance.IsSet(key+".cert_file") || !viperInstance.IsSet(key+".key_file")) {
			return ErrMutualTLS
		}
	}

	for _, key := range []string{"server.http.clients.server", "server.http.clients.agent"} {
		if viperInstance.IsSet(key+".server_cert_file") &&
			(!viperInstance.IsSet(key+".cert_file") || !viperInstance.IsSet(key+".key_file")) {
			return ErrMutualTLS
		}
	}

	return nil
}
