package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var configInstance Config

type Config interface {
	Init(cfgFileName string) error
	Viper() *viper.Viper
	// Validate() error
}

type config struct {
	viperInstance *viper.Viper
}

func Get() Config {
	if configInstance == nil {
		configInstance = &config{
			viperInstance: viper.New(),
		}
	}

	return configInstance
}

func (c *config) Init(cfgFileName string) error {
	if cfgFileName != "" {
		c.viperInstance.SetConfigFile(cfgFileName)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		c.viperInstance.AddConfigPath(".")
		c.viperInstance.AddConfigPath(home)
		c.viperInstance.SetConfigType("yaml")
		c.viperInstance.SetConfigName(".maf")
	}

	c.viperInstance.AutomaticEnv()

	if err := c.viperInstance.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	return nil
}

func (c *config) Viper() *viper.Viper {
	return c.viperInstance
}
