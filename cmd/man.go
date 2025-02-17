package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var manDir string

var manCmd = &cobra.Command{
	Use:   "man",
	Short: "Generate man pages for the maf CLI",
	Long: `This command generates man pages for the maf CLI. By default, it creates the
man page files in the "man" directory under the current directory.`,
	Run: func(_ *cobra.Command, _ []string) {
		header := &doc.GenManHeader{
			Title:   "MAF",
			Section: "1",
			Source:  "MySQL auto failover",
		}
		if _, err := os.Stat(manDir); !os.IsNotExist(err) {
			cobra.CheckErr(err)
		}
		err := doc.GenManTree(rootCmd, header, manDir)
		cobra.CheckErr(err)
	},
}

func init() {
	genCmd.AddCommand(manCmd)

	manCmd.Flags().StringVar(&manDir, "dir", "./man", "The directory to write the man pages to")
}
