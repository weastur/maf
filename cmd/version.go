package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weastur/maf/pkg/utils"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of maf",
	Long:  "All software has versions. This is maf's",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(utils.AppVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
