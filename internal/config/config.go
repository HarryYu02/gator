package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(home, configFileName)
	return configPath, nil
}

func Read() (Config, error) {
	configPath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(configContent, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func (cfg *Config) SetUser(user string) error {
	cfg.CurrentUserName = user

	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	configBytes, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	err = os.WriteFile(configPath, configBytes, 0666)
	if err != nil {
		return err
	}

	return nil
}
