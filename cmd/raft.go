package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var raftCmd = &cobra.Command{
	Use:   "raft",
	Short: "Low-level Raft commands",
	Long: `Commands to interact with the raft consensus mechanism.
It's highly recommended to use these commands ONLY for debugging purposes.`,
}

var kvCmd = &cobra.Command{
	Use:   "kv",
	Short: "Key-value store commands",
	Long:  `Commands to interact with the key-value store.`,
}

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get value for key",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := getServerAPIClient()
		value, ok, err := client.RaftKVGet(args[0])
		cobra.CheckErr(err)
		if !ok {
			return
		}
		fmt.Println(value)
	},
}

var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set value for key",
	Args:  cobra.ExactArgs(2), //nolint:mnd
	Run: func(_ *cobra.Command, args []string) {
		client := getServerAPIClient()
		cobra.CheckErr(client.RaftKVSet(args[0], args[1]))
	},
}

var delCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete value by key",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := getServerAPIClient()
		cobra.CheckErr(client.RaftKVDelete(args[0]))
	},
}

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
	serverCmd.AddCommand(raftCmd)

	raftCmd.AddCommand(kvCmd)
	raftCmd.AddCommand(forgetCmd)

	kvCmd.AddCommand(getCmd)
	kvCmd.AddCommand(setCmd)
	kvCmd.AddCommand(delCmd)
}
