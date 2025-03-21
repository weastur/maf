package cmd

import (
	"time"

	"github.com/weastur/maf/pkg/config"
	serverAPIClient "github.com/weastur/maf/pkg/server/client"
)

const (
	defaultHTTPReadTimeout             = 5 * time.Second
	defaultHTTPWriteTimeout            = 5 * time.Second
	defaultHTTPIdleTimeout             = 60 * time.Second
	defaultHTTPGracefulShutdownTimeout = 5 * time.Second
)

type ServerAPIClient interface {
	RaftKVGet(key string) (string, bool, error)
}

func getServerAPIClient() ServerAPIClient {
	var cfg Config = config.Get()

	viper := cfg.Viper()

	tlsConfig := &serverAPIClient.TLSConfig{
		CertFile:       viper.GetString("server.http.clients.server.cert_file"),
		KeyFile:        viper.GetString("server.http.clients.server.key_file"),
		ServerCertFile: viper.GetString("server.http.clients.server.server_cert_file"),
	}

	return serverAPIClient.NewWithAutoTLS(viper.GetString("server.http.advertise"), tlsConfig, false)
}
