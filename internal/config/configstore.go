package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	*viper.Viper
}

var (
	Cfg *Config

	ErrKeyNameMustNotBeEmpty = errors.New("key name must not be empty string")
	ErrKeyNotFound           = errors.New("key not found")
)

func New() *Config {
	return &Config{Viper: viper.New()}
}

func (c *Config) InitConfig(configFile string) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("error fetching the user config dir %w", err)
	}

	err = os.MkdirAll(configDir, fs.FileMode(0o644))
	if err != nil {
		return fmt.Errorf("error ensuring the existence of config dir %w", err)
	}

	dbConfigPath := filepath.Join(configDir, configFile)
	if _, err := os.Stat(dbConfigPath); os.IsNotExist(err) {
		f, err := os.Create(dbConfigPath)
		if err != nil {
			return fmt.Errorf("failed to create config file %w", err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("error closing file %w", err)
		}
	}

	c.SetConfigFile(dbConfigPath)

	if err := c.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	return nil
}

func (c *Config) Upsert(k, v string) error {
	if k == "" {
		return ErrKeyNameMustNotBeEmpty
	}

	c.Set(k, v)

	if err := c.WriteConfig(); err != nil {
		if err := c.SafeWriteConfig(); err != nil {
			return fmt.Errorf("error writing config file: %w", err)
		}
	}

	return nil
}

func (c *Config) Get(k string) (string, error) {
	if k == "" {
		return "", ErrKeyNameMustNotBeEmpty
	}

	if !c.IsSet(k) {
		return "", fmt.Errorf("%s: %w", k, ErrKeyNotFound)
	}

	return c.GetString(k), nil
}
