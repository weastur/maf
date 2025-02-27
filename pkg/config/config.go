package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var configInstance Config

type Config interface {
	Init(cfgFileName string) error
	Viper() *viper.Viper
}

type config struct {
	viperInstance *viper.Viper
	validators    []validator
}

func Get() Config {
	if configInstance == nil {
		configInstance = &config{
			viperInstance: viper.New(),
			validators: []validator{
				&validatorMutualTLSMisconfig{},
			},
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

	return c.validate()
}

func (c *config) Viper() *viper.Viper {
	return c.viperInstance
}

func (c *config) validate() error {
	errs := make([]error, len(c.validators))

	for i, validator := range c.validators {
		errs[i] = validator.Validate(c.viperInstance)
	}

	err := errors.Join(errs...)
	if err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	return nil
}
