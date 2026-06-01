package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type appConfig struct {
	Theme        int    `json:"theme"`
	Color        int    `json:"color"`
	Mode         int    `json:"mode"`
	ActiveEnv    string `json:"active_env,omitempty"`
	HistoryLimit int    `json:"history_limit,omitempty"`
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
		return appConfig{Theme: 0, Color: 0, Mode: 0}
	}

	var cfg appConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return appConfig{Theme: 0, Color: 0, Mode: 0}
	}

	// Upper bound is validated in the model once custom themes are loaded,
	// since custom themes extend the index space beyond the built-ins.
	if cfg.Theme < 0 {
		cfg.Theme = 0
	}
	if cfg.Color < 0 || cfg.Color >= len(colorSchemes) {
		cfg.Color = 0
	}
	if cfg.Mode < 0 || cfg.Mode > 1 {
		cfg.Mode = 0
	}
	if cfg.HistoryLimit <= 0 {
		cfg.HistoryLimit = 500
	}

	return cfg
}

func saveConfig(cfg appConfig) {
	path := getConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}

	data, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return
	}
}

func deleteConfig() {
	os.Remove(getConfigPath())
}
