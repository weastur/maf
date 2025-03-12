package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/weastur/maf/pkg/config"
	"github.com/weastur/maf/pkg/server"
	"github.com/weastur/maf/pkg/server/worker/fiber"
	"github.com/weastur/maf/pkg/server/worker/raft"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run a maf server",
	Long: `Run a maf server that will communicate with the agents and perform failover if needed.
It is designed to run on a separate host.`,
	Run: func(cmd *cobra.Command, args []string) { //nolint:revive
		var cfg Config = config.Get()

		viper := cfg.Viper()

		serverConfig := &server.Config{
			LogLevel:  viper.GetString("server.log.level"),
			LogPretty: viper.GetBool("server.log.pretty"),
			SentryDSN: viper.GetString("server.sentry.dsn"),
		}

		fiberConfig := &fiber.Config{
			Addr:            viper.GetString("server.http.addr"),
			CertFile:        viper.GetString("server.http.cert_file"),
			KeyFile:         viper.GetString("server.http.key_file"),
			ClientCertFile:  viper.GetString("server.http.client_cert_file"),
			ReadTimeout:     viper.GetDuration("server.http.read_timeout"),
			WriteTimeout:    viper.GetDuration("server.http.write_timeout"),
			IdleTimeout:     viper.GetDuration("server.http.idle_timeout"),
			ShutdownTimeout: viper.GetDuration("server.http.graceful_shutdown_timeout"),
		}

		raftConfig := &raft.Config{
			Addr:    viper.GetString("server.raft.addr"),
			NodeID:  viper.GetString("server.raft.node_id"),
			Devmode: viper.GetBool("server.raft.devmode"),
			Peers:   viper.GetStringSlice("server.raft.peers"),
			Datadir: viper.GetString("server.raft.data_dir"),
		}

		srv := server.Get(serverConfig, raftConfig, fiberConfig)

		if err := srv.Run(); err != nil {
			log.Fatal().Err(err).Msg("server failed")
		}
	},
}

func init() {
	var cfg Config = config.Get()

	viper := cfg.Viper()

	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().String("http-addr", ":7080", "Address to listen to")
	serverCmd.Flags().String("http-cert-file", "", "Path to the cert file (required if key-file is set)")
	serverCmd.Flags().String("http-key-file", "", "Path to the key file (required if cert-file is set)")
	serverCmd.Flags().String("http-client-cert-file", "", "Path to the client cert file (for mTLS)")
	serverCmd.Flags().Duration("http-read-timeout", defaultHTTPReadTimeout, "HTTP read timeout")
	serverCmd.Flags().Duration("http-write-timeout", defaultHTTPWriteTimeout, "HTTP write timeout")
	serverCmd.Flags().Duration("http-idle-timeout", defaultHTTPIdleTimeout, "HTTP idle timeout")
	serverCmd.Flags().Duration(
		"http-graceful-shutdown-timeout",
		defaultHTTPGracefulShutdownTimeout,
		"HTTP graceful shutdown timeout",
	)

	serverCmd.Flags().String("log-level", "info", "Log level (trace, debug, info, warn, error, fatal, panic)")
	serverCmd.Flags().Bool("log-pretty", false, "Enable pretty logging")

	serverCmd.Flags().String("sentry-dsn", "", "Sentry DSN")

	serverCmd.Flags().String("raft-addr", ":7081", "Raft address to listen to")
	serverCmd.Flags().String("raft-node-id", "", "Raft node ID")
	serverCmd.Flags().String("raft-data-dir", "/var/lib/maf", "Raft data directory")
	serverCmd.Flags().Bool("raft-devmode", false, "Store Raft data in memory")
	serverCmd.Flags().StringArray("raft-peers", []string{}, "Raft peers")

	serverCmd.MarkFlagFilename("http-cert-file")
	serverCmd.MarkFlagFilename("http-key-file")
	serverCmd.MarkFlagFilename("http-client-cert-file")

	viper.BindPFlag("server.http.addr", serverCmd.Flags().Lookup("addr"))
	viper.BindPFlag("server.http.cert_file", serverCmd.Flags().Lookup("cert-file"))
	viper.BindPFlag("server.http.key_file", serverCmd.Flags().Lookup("key-file"))
	viper.BindPFlag("server.http.client_cert_file", serverCmd.Flags().Lookup("client-cert-file"))
	viper.BindPFlag("server.http.read_timeout", serverCmd.Flags().Lookup("http-read-timeout"))
	viper.BindPFlag("server.http.write_timeout", serverCmd.Flags().Lookup("http-write-timeout"))
	viper.BindPFlag("server.http.idle_timeout", serverCmd.Flags().Lookup("http-idle-timeout"))
	viper.BindPFlag("server.http.graceful_shutdown_timeout", serverCmd.Flags().Lookup("http-graceful-shutdown-timeout"))

	viper.BindPFlag("server.log.level", serverCmd.Flags().Lookup("log-level"))
	viper.BindPFlag("server.log.pretty", serverCmd.Flags().Lookup("log-pretty"))

	viper.BindPFlag("server.sentry.dsn", serverCmd.Flags().Lookup("sentry-dsn"))

	viper.BindPFlag("server.raft.addr", serverCmd.Flags().Lookup("raft-addr"))
	viper.BindPFlag("server.raft.node_id", serverCmd.Flags().Lookup("raft-node-id"))
	viper.BindPFlag("server.raft.data_dir", serverCmd.Flags().Lookup("raft-data-dir"))
	viper.BindPFlag("server.raft.devmode", serverCmd.Flags().Lookup("raft-devmode"))
	viper.BindPFlag("server.raft.peers", serverCmd.Flags().Lookup("raft-peers"))
}
