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
	// determine the base config path
	baseConfigPath, err := os.UserConfigDir()
	if err != nil {
		return ErrConfigDirNotFound
	}

	// frame the application config path where appConfigPath = baseConfigPath + appName
	appConfigDirPath := filepath.Join(baseConfigPath, appName)
	appConfigFilePath := filepath.Join(appConfigDirPath, configFile)

	// create the base directories regardless of their presence (stat); os.MkdirAll is an idempotent process
	// which works just like mkdir -p
	if err := os.MkdirAll(appConfigDirPath, fs.FileMode(0o755)); err != nil {
		return ErrFailedDirectoryCreation
	}

	// once the dir creation succeeds; we will move towards config file creation
	file, err := os.OpenFile(appConfigFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	defer file.Close() //nolint:errcheck // TODO: check error and return

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
		return ErrConfigFileWrite
	}

	return nil
}
