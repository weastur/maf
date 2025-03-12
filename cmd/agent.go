package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/weastur/maf/pkg/agent"
	"github.com/weastur/maf/pkg/config"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run a maf agent",
	Long: `Run a maf agent that will monitor the MySQL instance and perform failover if needed.
It is designed to run on the same host as the MySQL instance.`,
	Run: func(cmd *cobra.Command, args []string) { //nolint:revive
		var cfg Config = config.Get()

		viper := cfg.Viper()
		addr := viper.GetString("agent.addr")
		certFile := viper.GetString("agent.cert_file")
		keyFile := viper.GetString("agent.key_file")
		clientCertFile := viper.GetString("agent.client_cert_file")
		logLevel := viper.GetString("agent.log.level")
		logPretty := viper.GetBool("agent.log.pretty")
		readTimeout := viper.GetDuration("agent.http_read_timeout")
		writeTimeout := viper.GetDuration("agent.http_write_timeout")
		idleTimeout := viper.GetDuration("agent.http_idle_timeout")
		sentryDSN := viper.GetString("agent.sentry.dsn")

		agent := agent.Get(
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

		if err := agent.Run(); err != nil {
			log.Fatal().Err(err).Msg("agent failed")
		}
	},
}

func init() {
	var cfg Config = config.Get()

	viper := cfg.Viper()

	rootCmd.AddCommand(agentCmd)
	agentCmd.Flags().String("addr", ":7070", "Address to listen to")
	agentCmd.Flags().String("cert-file", "", "Path to the cert file (required if key-file is set)")
	agentCmd.Flags().String("key-file", "", "Path to the key file (required if cert-file is set)")
	agentCmd.Flags().String("client-cert-file", "", "Path to the client cert file (for mTLS)")
	agentCmd.Flags().Duration("http-read-timeout", defaultHTTPReadTimeout, "HTTP read timeout")
	agentCmd.Flags().Duration("http-write-timeout", defaultHTTPWriteTimeout, "HTTP write timeout")
	agentCmd.Flags().Duration("http-idle-timeout", defaultHTTPIdleTimeout, "HTTP idle timeout")
	agentCmd.Flags().String("log-level", "info", "Log level (trace, debug, info, warn, error, fatal, panic)")
	agentCmd.Flags().Bool("log-pretty", false, "Enable pretty logging")
	agentCmd.Flags().String("sentry-dsn", "", "Sentry DSN")
	agentCmd.MarkFlagFilename("cert-file")
	agentCmd.MarkFlagFilename("key-file")
	agentCmd.MarkFlagFilename("client-cert-file")

	viper.BindPFlag("agent.addr", agentCmd.Flags().Lookup("addr"))
	viper.BindPFlag("agent.cert_file", agentCmd.Flags().Lookup("cert-file"))
	viper.BindPFlag("agent.key_file", agentCmd.Flags().Lookup("key-file"))
	viper.BindPFlag("agent.client_cert_file", agentCmd.Flags().Lookup("client-cert-file"))
	viper.BindPFlag("agent.http_read_timeout", agentCmd.Flags().Lookup("http-read-timeout"))
	viper.BindPFlag("agent.http_write_timeout", agentCmd.Flags().Lookup("http-write-timeout"))
	viper.BindPFlag("agent.http_idle_timeout", agentCmd.Flags().Lookup("http-idle-timeout"))
	viper.BindPFlag("agent.log.level", agentCmd.Flags().Lookup("log-level"))
	viper.BindPFlag("agent.log.pretty", agentCmd.Flags().Lookup("log-pretty"))
	viper.BindPFlag("agent.sentry.dsn", agentCmd.Flags().Lookup("sentry-dsn"))
}
