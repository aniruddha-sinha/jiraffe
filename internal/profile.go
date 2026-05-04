package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const APPLICATION_NAME = "jiraffe"

func getPaths() (ProfilePath, error) {
	XDGConfigDir, err := os.UserConfigDir()
	if err != nil {
		return ProfilePath{}, fmt.Errorf("could not find the user config directory %w", err)
	}

	dirPath := filepath.Join(XDGConfigDir, APPLICATION_NAME)
	return ProfilePath{
		DirPath:  dirPath,
		FilePath: filepath.Join(dirPath, "config.yaml"),
	}, nil
}

func loadProfileConfig() error {
	paths, err := getPaths()
	if err != nil {
		return err
	}

	viper.SetConfigFile(paths.FilePath)
	return viper.ReadInConfig()
}

func Save(p UserProfile, encodedToken string) (string, error) {
	paths, err := getPaths()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(paths.DirPath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(paths.DirPath, 0o700); err != nil {
				return "", fmt.Errorf("failed to create the config directory")
			}
		}
	}

	viper.Set("auth.email", p.Email)
	viper.Set("auth.org", p.Org)
	viper.Set("auth.encoded_token", encodedToken)

	if err := viper.WriteConfig(); err != nil {
		return "", fmt.Errorf("failed to save the config file %w", err)
	}

	return paths.FilePath, nil
}
