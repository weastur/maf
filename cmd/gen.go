package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	docDir string
	manDir string
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate documentation",
	Long:  `Generate documentation for the maf.`,
}

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
	rootCmd.AddCommand(genCmd)
	genCmd.AddCommand(docCmd)
	genCmd.AddCommand(manCmd)

	docCmd.Flags().StringVar(&docDir, "dir", "./mafdoc", "The directory to write the markdown files to")

	manCmd.Flags().StringVar(&manDir, "dir", "./man", "The directory to write the man pages to")
}
