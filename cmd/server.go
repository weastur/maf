package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/weastur/maf/pkg/config"
	"github.com/weastur/maf/pkg/server"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run a maf server",
	Long: `Run a maf server that will communicate with the agents and perform failover if needed.
It is designed to run on a separate host.`,
	Run: func(cmd *cobra.Command, args []string) { //nolint:revive
		var cfg Config = config.Get()

		viper := cfg.Viper()
		addr := viper.GetString("server.addr")
		certFile := viper.GetString("server.cert_file")
		keyFile := viper.GetString("server.key_file")
		clientCertFile := viper.GetString("server.client_cert_file")
		logLevel := viper.GetString("server.log.level")
		logPretty := viper.GetBool("server.log.pretty")
		readTimeout := viper.GetDuration("server.http_read_timeout")
		writeTimeout := viper.GetDuration("server.http_write_timeout")
		idleTimeout := viper.GetDuration("server.http_idle_timeout")
		sentryDSN := viper.GetString("server.sentry.dsn")

		server := server.Get(
			addr,
			certFile,
			keyFile,
			clientCertFile,
			logLevel,
			logPretty,
			readTimeout,
			writeTimeout,
			idleTimeout,
			sentryDSN,
		)

		if err := server.Run(); err != nil {
			log.Fatal().Err(err).Msg("server failed")
		}
	},
}

func init() {
	var cfg Config = config.Get()

	viper := cfg.Viper()

	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().String("addr", ":7080", "Address to listen to")
	serverCmd.Flags().String("cert-file", "", "Path to the cert file (required if key-file is set)")
	serverCmd.Flags().String("key-file", "", "Path to the key file (required if cert-file is set)")
	serverCmd.Flags().String("client-cert-file", "", "Path to the client cert file (for mTLS)")
	serverCmd.Flags().Duration("http-read-timeout", defaultHTTPReadTimeout, "HTTP read timeout")
	serverCmd.Flags().Duration("http-write-timeout", defaultHTTPWriteTimeout, "HTTP write timeout")
	serverCmd.Flags().Duration("http-idle-timeout", defaultHTTPIdleTimeout, "HTTP idle timeout")
	serverCmd.Flags().String("log-level", "info", "Log level (trace, debug, info, warn, error, fatal, panic)")
	serverCmd.Flags().Bool("log-pretty", false, "Enable pretty logging")
	serverCmd.Flags().String("sentry-dsn", "", "Sentry DSN")
	serverCmd.MarkFlagsRequiredTogether("cert-file", "key-file")
	serverCmd.MarkFlagFilename("cert-file")
	serverCmd.MarkFlagFilename("key-file")
	serverCmd.MarkFlagFilename("client-cert-file")

	viper.BindPFlag("server.addr", serverCmd.Flags().Lookup("addr"))
	viper.BindPFlag("server.cert_file", serverCmd.Flags().Lookup("cert-file"))
	viper.BindPFlag("server.key_file", serverCmd.Flags().Lookup("key-file"))
	viper.BindPFlag("server.client_cert_file", serverCmd.Flags().Lookup("client-cert-file"))
	viper.BindPFlag("server.http_read_timeout", serverCmd.Flags().Lookup("http-read-timeout"))
	viper.BindPFlag("server.http_write_timeout", serverCmd.Flags().Lookup("http-write-timeout"))
	viper.BindPFlag("server.http_idle_timeout", serverCmd.Flags().Lookup("http-idle-timeout"))
	viper.BindPFlag("server.log.level", serverCmd.Flags().Lookup("log-level"))
	viper.BindPFlag("server.log.pretty", serverCmd.Flags().Lookup("log-pretty"))
	viper.BindPFlag("server.sentry.dsn", serverCmd.Flags().Lookup("sentry-dsn"))
}
