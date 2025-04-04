package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/weastur/maf/internal/config"
	serverAPIClient "github.com/weastur/maf/internal/server/client"
	"github.com/weastur/maf/internal/server/worker/fiber"
)

var errLeaderAddrNotFound = errors.New("leader API address not found")

const (
	defaultHTTPReadTimeout             = 5 * time.Second
	defaultHTTPWriteTimeout            = 5 * time.Second
	defaultHTTPIdleTimeout             = 60 * time.Second
	defaultHTTPGracefulShutdownTimeout = 5 * time.Second
)

type ServerAPIClient interface {
	RaftKVGet(key string) (string, bool, error)
	RaftKVSet(key, value string) error
	RaftKVDelete(key string) error
	RaftForget(serverID string) error
	RaftInfo(includeStats bool) (any, error)
}

func clientTLSConfig() *serverAPIClient.TLSConfig {
	var cfg Config = config.Get()

	viper := cfg.Viper()

	return &serverAPIClient.TLSConfig{
		CertFile:       viper.GetString("server.http.clients.server.cert_file"),
		KeyFile:        viper.GetString("server.http.clients.server.key_file"),
		ServerCertFile: viper.GetString("server.http.clients.server.server_cert_file"),
	}
}

func readLeaderAddr() (string, error) {
	var cfg Config = config.Get()

	viper := cfg.Viper()

	tlsConfig := clientTLSConfig()

	client := serverAPIClient.NewWithAutoTLS(viper.GetString("server.http.advertise"), tlsConfig, false)

	addr, ok, err := client.RaftKVGet(fiber.LeaderAPIAddrKey)
	if err != nil {
		return "", fmt.Errorf("failed to get leader API address: %w", err)
	}

	if !ok {
		return "", errLeaderAddrNotFound
	}

	return addr, nil
}

func getServerAPIClient(leader bool) ServerAPIClient {
	var addr string

	var err error

	if leader {
		addr, err = readLeaderAddr()
		cobra.CheckErr(err)
	} else {
		var cfg Config = config.Get()

		viper := cfg.Viper()

		addr = viper.GetString("server.http.advertise")
	}

	tlsConfig := clientTLSConfig()

	return serverAPIClient.NewWithAutoTLS(addr, tlsConfig, false)
}
