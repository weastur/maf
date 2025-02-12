package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "v0.0.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of maf",
	Long:  "All software has versions. This is maf's",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
