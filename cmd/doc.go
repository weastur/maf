package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docDir string

var docCmd = &cobra.Command{
	Use:   "doc",
	Short: "Generate Markdown documentation for the maf CLI",
	Long: `This command generates Markdown documentation for the maf CLI.
By default, it creates the md files in the "mafdoc" directory under the current directory.`,
	Run: func(_ *cobra.Command, _ []string) {
		if _, err := os.Stat(docDir); !os.IsNotExist(err) {
			cobra.CheckErr(err)
		}
		err := doc.GenMarkdownTree(rootCmd, docDir)
		cobra.CheckErr(err)
	},
}

func init() {
	genCmd.AddCommand(docCmd)

	docCmd.Flags().StringVar(&docDir, "dir", "./mafdoc", "The directory to write the markdown files to")
}
