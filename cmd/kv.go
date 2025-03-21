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
	Use:   "get",
	Short: "Get value for key",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("get called")
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
	rootCmd.AddCommand(kvCmd)

	kvCmd.AddCommand(getCmd)
	kvCmd.AddCommand(setCmd)
	kvCmd.AddCommand(delCmd)
}
