package config

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
	"github.com/weastur/maf/internal/config/validate"
)

type Validator interface {
	Validate(viperInstance *viper.Viper) error
}

type Config struct {
	viperInstance *viper.Viper
	validators    []Validator
}

var (
	instance *Config
	once     sync.Once
)

func Get() *Config {
	once.Do(func() {
		instance = &Config{
			viperInstance: viper.New(),
			validators: []Validator{
				validate.NewMutualTLS(),
				validate.NewTLS(),
				validate.NewLogLevel(),
				validate.NewRaft(),
			},
		}
	})

	return instance
}

func (c *Config) Init(cfgFileName string) error {
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
		var errCfgNotFound viper.ConfigFileNotFoundError
		if errors.As(err, &errCfgNotFound) {
			// No config file found, so using defaults
			return c.validate()
		}

		return fmt.Errorf("failed to read config file: %w", err)
	}

	return c.validate()
}

func (c *Config) Viper() *viper.Viper {
	return c.viperInstance
}

func (c *Config) validate() error {
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
