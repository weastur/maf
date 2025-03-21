package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var errRequiredAtLeastOneArg = errors.New("requires at least one arg")

var kvCmd = &cobra.Command{
	Use:   "kv",
	Short: "Key-value store commands",
	Long: `Commands to interact with the key-value store.
It's highly recommended to use these commands ONLY for debugging purposes.`,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get value for key",
	Args: func(_ *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errRequiredAtLeastOneArg
		}

		return nil
	},
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
	Use:   "set",
	Short: "Set value for key",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("set called")
	},
}

var delCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete value by key",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("delete called")
	},
}

func init() {
	serverCmd.AddCommand(kvCmd)

	kvCmd.AddCommand(getCmd)
	kvCmd.AddCommand(setCmd)
	kvCmd.AddCommand(delCmd)
}
