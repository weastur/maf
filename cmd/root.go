package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/weastur/maf/pkg/config"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "maf",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(_ *cobra.Command, _ []string) {
		cfg := config.Get()
		fmt.Println(cfg.Viper().AllSettings())
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	cfg := config.Get()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.maf.yaml)")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	_ = cfg.Viper().BindPFlag("toggle", rootCmd.Flags().Lookup("toggle"))
}

func initConfig() {
	cfg := config.Get()
	cobra.CheckErr(cfg.Init(cfgFile))
}
