package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run a maf agent",
	Long: `Run a maf agent that will monitor the MySQL instance and perform failover if needed.
It is designed to run on the same host as the MySQL instance.`,
	Run: func(cmd *cobra.Command, args []string) { //nolint:revive
		fmt.Println("agent called")
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)

	agentCmd.PersistentFlags().String("foo", "", "A help for foo")
}
