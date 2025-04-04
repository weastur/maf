package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/weastur/maf/internal/config"
)

type Config interface {
	Init(cfgFileName string) error
	Viper() *viper.Viper
}

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "maf",
	Short: "MySQL auto failover",
	Long: `MySQL auto failover is a high-availability solution for MySQL.
It is designed to rule out the need for manual intervention in case of a
failure of the primary node.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.maf.yaml)")
}

func initConfig() {
	var cfg Config = config.Get()

	cobra.CheckErr(cfg.Init(cfgFile))
}
