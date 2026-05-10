package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ShortcutsConfig struct {
	Modifier        string `json:"modifier"`
	ToggleRecording string `json:"toggleRecording"`
	CycleDisplay    string `json:"cycleDisplay"`
	QuitApp         string `json:"quitApp"`
	ToggleMode      string `json:"toggleMode"`
}

type GifConfig struct {
	MaxDuration int `json:"maxDuration"`
	MaxWidth    int `json:"maxWidth"`
	Fps         int `json:"fps"`
}

type HoverConfig struct {
	TriggerWidth  int `json:"triggerWidth"`
	TriggerHeight int `json:"triggerHeight"`
	HideDelayMs   int `json:"hideDelayMs"`
	TopMargin     int `json:"topMargin"`
}

type Config struct {
	OutputDirectory string          `json:"outputDirectory"`
	RecordingMode   string          `json:"recordingMode"`
	Shortcuts       ShortcutsConfig `json:"shortcuts"`
	Hover           HoverConfig     `json:"hover"`
	Gif             GifConfig       `json:"gif"`
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
				RecordingMode:   "video",
				Shortcuts: ShortcutsConfig{
					Modifier:        "Ctrl+Shift",
					ToggleRecording: "R",
					CycleDisplay:    "D",
					QuitApp:         "Q",
					ToggleMode:      "G",
				},
				Hover: HoverConfig{
					TriggerWidth:  20,
					TriggerHeight: 20,
					HideDelayMs:   500,
					TopMargin:     40,
				},
				Gif: GifConfig{
					MaxDuration: 15,
					MaxWidth:    800,
					Fps:         15,
				},
			}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.Shortcuts.Modifier == "" {
		cfg.Shortcuts.Modifier = "Ctrl+Shift"
	}
	if cfg.Shortcuts.ToggleRecording == "" {
		cfg.Shortcuts.ToggleRecording = "R"
	}
	if cfg.Shortcuts.CycleDisplay == "" {
		cfg.Shortcuts.CycleDisplay = "D"
	}
	if cfg.Shortcuts.QuitApp == "" {
		cfg.Shortcuts.QuitApp = "Q"
	}
	if cfg.Shortcuts.ToggleMode == "" {
		cfg.Shortcuts.ToggleMode = "G"
	}
	if cfg.Hover.TriggerWidth == 0 {
		cfg.Hover.TriggerWidth = 20
	}
	if cfg.Hover.TriggerHeight == 0 {
		cfg.Hover.TriggerHeight = 20
	}
	if cfg.Hover.HideDelayMs == 0 {
		cfg.Hover.HideDelayMs = 500
	}
	if cfg.Hover.TopMargin == 0 {
		cfg.Hover.TopMargin = 40
	}
	if cfg.RecordingMode == "" {
		cfg.RecordingMode = "video"
	}
	if cfg.Gif.MaxDuration == 0 {
		cfg.Gif.MaxDuration = 15
	}
	if cfg.Gif.MaxWidth == 0 {
		cfg.Gif.MaxWidth = 800
	}
	if cfg.Gif.Fps == 0 {
		cfg.Gif.Fps = 15
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
