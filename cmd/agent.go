package cmd //nolint:dupl

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weastur/maf/pkg/config"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run a maf agent",
	Long: `Run a maf agent that will monitor the MySQL instance and perform failover if needed.
It is designed to run on the same host as the MySQL instance.`,
	Run: func(cmd *cobra.Command, args []string) { //nolint:revive
		viper := config.Get().Viper()
		fmt.Println(viper.GetString("agent.addr"))
		fmt.Println(viper.IsSet("agent.cert_file"))
	},
}

func init() {
	viper := config.Get().Viper()

	rootCmd.AddCommand(agentCmd)
	agentCmd.Flags().String("addr", ":7070", "Address to listen to")
	agentCmd.Flags().String("cert-file", "", "Path to the cert file (required if key-file is set)")
	agentCmd.Flags().String("key-file", "", "Path to the key file (required if cert-file is set)")
	agentCmd.Flags().String("client-cert-file", "", "Path to the client cert file (for mTLS)")
	agentCmd.MarkFlagsRequiredTogether("cert-file", "key-file")
	agentCmd.MarkFlagFilename("cert-file")
	agentCmd.MarkFlagFilename("key-file")
	agentCmd.MarkFlagFilename("client-cert-file")

	viper.BindPFlag("agent.addr", agentCmd.Flags().Lookup("addr"))
	viper.BindPFlag("agent.cert_file", agentCmd.Flags().Lookup("cert-file"))
	viper.BindPFlag("agent.key_file", agentCmd.Flags().Lookup("key-file"))
	viper.BindPFlag("agent.client_cert_file", agentCmd.Flags().Lookup("client-cert-file"))
}
