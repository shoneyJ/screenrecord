package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ShortcutsConfig struct {
	ToggleRecording string `json:"toggleRecording"`
	CycleDisplay    string `json:"cycleDisplay"`
}

type Config struct {
	OutputDirectory string          `json:"outputDirectory"`
	Shortcuts       ShortcutsConfig `json:"shortcuts"`
}

func GetConfigDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "screenrecord")
}

func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.json")
}

func Load() (*Config, error) {
	configPath := GetConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			defaultDir, _ := os.UserHomeDir()
			defaultDir = filepath.Join(defaultDir, "Videos")
			return &Config{
				OutputDirectory: defaultDir,
				Shortcuts: ShortcutsConfig{
					ToggleRecording: "Ctrl+R",
					CycleDisplay:    "Ctrl+Shift+D",
				},
			}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.Shortcuts.ToggleRecording == "" {
		cfg.Shortcuts.ToggleRecording = "Ctrl+R"
	}
	if cfg.Shortcuts.CycleDisplay == "" {
		cfg.Shortcuts.CycleDisplay = "Ctrl+Shift+D"
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	configDir := GetConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(GetConfigPath(), data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
