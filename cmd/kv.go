package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var kvCmd = &cobra.Command{
	Use:   "kv",
	Short: "Key-value store commands",
	Long: `Commands to interact with the key-value store.
It's highly recommended to use these commands ONLY for debugging purposes.`,
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

func init() {
	serverCmd.AddCommand(kvCmd)

	kvCmd.AddCommand(getCmd)
	kvCmd.AddCommand(setCmd)
	kvCmd.AddCommand(delCmd)
}
