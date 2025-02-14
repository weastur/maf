package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var checkConfigCmd = &cobra.Command{
	Use:   "check-config",
	Short: "Check the configuration file",
	Long: `Check the configuration file for any errors or missing values. It's a basic
validation of the configuration file. The successful check does not guarantee that
the application will work as you expected. It just save you from some basic mistakes,
like typos.`,
	Run: func(cmd *cobra.Command, args []string) { //nolint:revive
		fmt.Println("Config is valid")
	},
}

func init() {
	rootCmd.AddCommand(checkConfigCmd)
}
