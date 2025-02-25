package cmd //nolint:dupl

import (
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
		viper := config.Get().Viper()

		addr := viper.GetString("server.addr")
		certFile := viper.GetString("server.cert_file")
		keyFile := viper.GetString("server.key_file")
		clientCertFile := viper.GetString("server.client_cert_file")
		readTimeout := viper.GetDuration("server.http_read_timeout")
		writeTimeout := viper.GetDuration("server.http_write_timeout")
		idleTimeout := viper.GetDuration("server.http_idle_timeout")

		server := server.Get(addr, certFile, keyFile, clientCertFile, readTimeout, writeTimeout, idleTimeout)

		cobra.CheckErr(server.Run())
	},
}

func init() {
	viper := config.Get().Viper()

	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().String("addr", ":7080", "Address to listen to")
	serverCmd.Flags().String("cert-file", "", "Path to the cert file (required if key-file is set)")
	serverCmd.Flags().String("key-file", "", "Path to the key file (required if cert-file is set)")
	serverCmd.Flags().String("client-cert-file", "", "Path to the client cert file (for mTLS)")
	serverCmd.Flags().Duration("http-read-timeout", defaultHTTPReadTimeout, "HTTP read timeout")
	serverCmd.Flags().Duration("http-write-timeout", defaultHTTPWriteTimeout, "HTTP write timeout")
	serverCmd.Flags().Duration("http-idle-timeout", defaultHTTPIdleTimeout, "HTTP idle timeout")
	serverCmd.MarkFlagsRequiredTogether("cert-file", "key-file")
	serverCmd.MarkFlagFilename("cert-file")
	serverCmd.MarkFlagFilename("key-file")
	serverCmd.MarkFlagFilename("client-cert-file")

	viper.BindPFlag("server.addr", serverCmd.Flags().Lookup("addr"))
	viper.BindPFlag("server.cert_file", serverCmd.Flags().Lookup("cert-file"))
	viper.BindPFlag("server.key_file", serverCmd.Flags().Lookup("key-file"))
	viper.BindPFlag("server.client_cert_file", serverCmd.Flags().Lookup("client-cert-file"))
	viper.BindPFlag("server.http_read_timeout", agentCmd.Flags().Lookup("http-read-timeout"))
	viper.BindPFlag("server.http_write_timeout", agentCmd.Flags().Lookup("http-write-timeout"))
	viper.BindPFlag("server.http_idle_timeout", agentCmd.Flags().Lookup("http-idle-timeout"))
}
