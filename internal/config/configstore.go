package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type config struct {
	*viper.Viper
}

func New() *config {
	return &config{Viper: viper.New()}
}

var Cfg *config

var (
	ErrConfigDirNotFound       = errors.New("config file not found")
	ErrFailedDirectoryCreation = errors.New("failed to create the config directory")
	ErrKeyNotFound             = errors.New("key not found")
	ErrConfigFileWrite         = errors.New("error writing to config file")
)

func (c *config) InitConfig(appName, configFile string) error {
	baseConfigPath, err := os.UserConfigDir()
	if err != nil {
		return ErrConfigDirNotFound
	}

	appConfigDirPath := filepath.Join(baseConfigPath, appName)
	appConfigFilePath := filepath.Join(appConfigDirPath, configFile)

	if err := os.MkdirAll(appConfigDirPath, fs.FileMode(0o755)); err != nil {
		return ErrFailedDirectoryCreation
	}

	file, err := os.OpenFile(appConfigFilePath, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	file.Close() // nolint: errcheck

	c.SetConfigFile(appConfigFilePath)

	c.SetConfigType("yaml")

	if err := c.ReadInConfig(); err != nil {
		var emptyErr viper.ConfigFileNotFoundError

		if !errors.As(err, &emptyErr) {
			fileInfo, statErr := os.Stat(appConfigFilePath)
			if statErr == nil && fileInfo.Size() > 0 {
				return fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	return nil
}

func (c *config) Get(k string) (string, error) {
	// check if the key is present
	if !c.IsSet(k) {
		return "", ErrKeyNotFound
	}

	value := c.GetString(k)
	return value, nil
}

func (c *config) Upsert(k, v string) error {
	c.Set(k, v)

	if err := c.WriteConfig(); err != nil {
		return fmt.Errorf("%w: %w", ErrConfigFileWrite, err)
	}

	return nil
}
