package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	MaxResults      int      `json:"max_results"`
	ScriptDirs      []string `json:"script_dirs"`
	Theme           string   `json:"theme"`
	EnableClipboard bool     `json:"enable_clipboard"`
	EnableHyprland  bool     `json:"enable_hyprland"`
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		MaxResults: 50,
		ScriptDirs: []string{
			filepath.Join(home, ".config", "quickcast", "scripts"),
		},
		Theme:           "raycast-dark",
		EnableClipboard: true,
		EnableHyprland:  true,
	}
}

func LoadConfig() *Config {
	cfg := DefaultConfig()
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "quickcast", "config.json")

	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, cfg)
	}

	return cfg
}
