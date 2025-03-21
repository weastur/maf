package cmd

import (
	"github.com/spf13/cobra"
)

var forgetCmd = &cobra.Command{
	Use:   "forget [serverID]",
	Short: "Forget server",
	Long: `Remove server from the cluster. Server will be demoted and removed.
Make sure you know what you are doing and will have enough servers to keep the quorum.`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := getServerAPIClient()
		cobra.CheckErr(client.RaftForget(args[0]))
	},
}

func init() {
	serverCmd.AddCommand(forgetCmd)
}
