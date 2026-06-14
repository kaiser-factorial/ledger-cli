package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Thresholds struct {
		YellowDays int `json:"yellowDays"`
		RedDays    int `json:"redDays"`
	} `json:"thresholds"`
	PromptTime string `json:"promptTime"`
}

func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ledger", "config.json")
}

func Load() (*Config, error) {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}