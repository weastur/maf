package cmd

import (
	"github.com/spf13/cobra"
	"github.com/weastur/maf/internal/agent"
	"github.com/weastur/maf/internal/config"

	"github.com/weastur/maf/internal/agent/worker/fiber"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run a maf agent",
	Long: `Run a maf agent that will monitor the MySQL instance and perform failover if needed.
It is designed to run on the same host as the MySQL instance.`,
	Run: func(cmd *cobra.Command, args []string) { //nolint:revive
		var cfg Config = config.Get()

		viper := cfg.Viper()

		agentConfig := &agent.Config{
			LogLevel:  viper.GetString("agent.log.level"),
			LogPretty: viper.GetBool("agent.log.pretty"),
			SentryDSN: viper.GetString("agent.sentry.dsn"),
		}

		fiberConfig := &fiber.Config{
			Addr:            viper.GetString("agent.http.addr"),
			CertFile:        viper.GetString("agent.http.cert_file"),
			KeyFile:         viper.GetString("agent.http.key_file"),
			ClientCertFile:  viper.GetString("agent.http.client_cert_file"),
			ReadTimeout:     viper.GetDuration("agent.http.read_timeout"),
			WriteTimeout:    viper.GetDuration("agent.http.write_timeout"),
			IdleTimeout:     viper.GetDuration("agent.http.idle_timeout"),
			ShutdownTimeout: viper.GetDuration("agent.http.graceful_shutdown_timeout"),
		}

		agent := agent.Get(agentConfig, fiberConfig)
		cobra.CheckErr(agent.Init())

		agent.Run()
	},
}

func init() {
	var cfg Config = config.Get()

	viper := cfg.Viper()

	rootCmd.AddCommand(agentCmd)

	agentCmd.Flags().String("http-addr", ":7070", "Address to listen to")
	agentCmd.Flags().String("http-cert-file", "", "Path to the cert file (required if key-file is set)")
	agentCmd.Flags().String("http-key-file", "", "Path to the key file (required if cert-file is set)")
	agentCmd.Flags().String("http-client-cert-file", "", "Path to the client cert file (for mTLS)")
	agentCmd.Flags().Duration("http-read-timeout", defaultHTTPReadTimeout, "HTTP read timeout")
	agentCmd.Flags().Duration("http-write-timeout", defaultHTTPWriteTimeout, "HTTP write timeout")
	agentCmd.Flags().Duration("http-idle-timeout", defaultHTTPIdleTimeout, "HTTP idle timeout")
	agentCmd.Flags().Duration(
		"http-graceful-shutdown-timeout",
		defaultHTTPGracefulShutdownTimeout,
		"HTTP graceful shutdown timeout",
	)

	agentCmd.Flags().String("log-level", "info", "Log level (trace, debug, info, warn, error, fatal, panic)")
	agentCmd.Flags().Bool("log-pretty", false, "Enable pretty logging")

	agentCmd.Flags().String("sentry-dsn", "", "Sentry DSN")

	agentCmd.MarkFlagFilename("cert-file")
	agentCmd.MarkFlagFilename("key-file")
	agentCmd.MarkFlagFilename("client-cert-file")

	viper.BindPFlag("agent.http.addr", agentCmd.Flags().Lookup("http-addr"))
	viper.BindPFlag("agent.http.cert_file", agentCmd.Flags().Lookup("http-cert-file"))
	viper.BindPFlag("agent.http.key_file", agentCmd.Flags().Lookup("http-key-file"))
	viper.BindPFlag("agent.http.client_cert_file", agentCmd.Flags().Lookup("http-client-cert-file"))
	viper.BindPFlag("agent.http.read_timeout", agentCmd.Flags().Lookup("http-read-timeout"))
	viper.BindPFlag("agent.http.write_timeout", agentCmd.Flags().Lookup("http-write-timeout"))
	viper.BindPFlag("agent.http.idle_timeout", agentCmd.Flags().Lookup("http-idle-timeout"))
	viper.BindPFlag("agent.http.graceful_shutdown_timeout", agentCmd.Flags().Lookup("http-graceful-shutdown-timeout"))

	viper.BindPFlag("agent.log.level", agentCmd.Flags().Lookup("log-level"))
	viper.BindPFlag("agent.log.pretty", agentCmd.Flags().Lookup("log-pretty"))

	viper.BindPFlag("agent.sentry.dsn", agentCmd.Flags().Lookup("sentry-dsn"))
}
