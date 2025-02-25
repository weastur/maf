package cmd //nolint:dupl

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weastur/maf/pkg/config"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run a maf server",
	Long: `Run a maf server that will communicate with the agents and perform failover if needed.
It is designed to run on a separate host.`,
	Run: func(cmd *cobra.Command, args []string) { //nolint:revive
		viper := config.Get().Viper()
		fmt.Println(viper.GetString("server.addr"))
		fmt.Println(viper.IsSet("server.cert_file"))
	},
}

func init() {
	viper := config.Get().Viper()

	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().String("addr", ":7080", "Address to listen to")
	serverCmd.Flags().String("cert-file", "", "Path to the cert file (required if key-file is set)")
	serverCmd.Flags().String("key-file", "", "Path to the key file (required if cert-file is set)")
	serverCmd.Flags().String("client-cert-file", "", "Path to the client cert file (for mTLS)")
	serverCmd.MarkFlagsRequiredTogether("cert-file", "key-file")
	serverCmd.MarkFlagFilename("cert-file")
	serverCmd.MarkFlagFilename("key-file")
	serverCmd.MarkFlagFilename("client-cert-file")

	viper.BindPFlag("server.addr", serverCmd.Flags().Lookup("addr"))
	viper.BindPFlag("server.cert_file", serverCmd.Flags().Lookup("cert-file"))
	viper.BindPFlag("server.key_file", serverCmd.Flags().Lookup("key-file"))
	viper.BindPFlag("server.client_cert_file", serverCmd.Flags().Lookup("client-cert-file"))
}
