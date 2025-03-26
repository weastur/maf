package validate

import (
	"errors"

	"github.com/spf13/viper"
)

type TLS struct{}

var ErrTLS = errors.New(
	"both cert-file and key-file must be set",
)

func NewTLS() *TLS {
	return &TLS{}
}

func (v *TLS) Validate(viperInstance *viper.Viper) error {
	for _, key := range []string{"agent.http", "server.http"} {
		if viperInstance.IsSet(key+".cert_file") != viperInstance.IsSet(key+".key_file") {
			return ErrTLS
		}
	}

	for _, key := range []string{"server.http.clients.server", "server.http.clients.agent"} {
		if viperInstance.IsSet(key+".cert_file") != viperInstance.IsSet(key+".key_file") {
			return ErrTLS
		}
	}

	return nil
}
