package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type appConfig struct {
	theme int `json:"theme"`
	color int `json:"color"`
}

func getConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir, _ = os.UserHomeDir()
	}
	return filepath.Join(dir, "hilo", "config.json")
}

func loadConfig() appConfig {
	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		return appConfig{theme: 0, color: 0}
	}

	var cfg appConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return appConfig{theme: 0, color: 0}
	}

	if cfg.theme < 0 || cfg.theme >= len(themes) {
		cfg.theme = 0
	}
	if cfg.color < 0 || cfg.color >= len(colorSchemes) {
		cfg.color = 0
	}

	return cfg
}

func saveConfig(cfg appConfig) {
	path := getConfigPath()
	os.MkdirAll(filepath.Dir(path), 0o755)

	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(path, data, 0o644)
}
