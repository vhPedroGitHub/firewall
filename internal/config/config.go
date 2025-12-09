package config

import (
	"encoding/json"
	"os"
)

// Config represents application configuration.
type Config struct {
	// Database settings
	DBPath string `json:"db_path"`

	// Logging settings
	LogPath string `json:"log_path"`

	// Default profile
	DefaultProfile string `json:"default_profile"`

	// GUI settings
	GUI GUIConfig `json:"gui"`
}

// GUIConfig represents GUI-specific settings.
type GUIConfig struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Theme  string `json:"theme"`
}

// Default returns a Config with sensible defaults.
func Default() Config {
	return Config{
		DBPath:         "firewall.db",
		LogPath:        "firewall.log",
		DefaultProfile: "",
		GUI: GUIConfig{
			Width:  1024,
			Height: 768,
			Theme:  "dark",
		},
	}
}

// Load reads configuration from a JSON file.
// If the file doesn't exist, returns default configuration.
func Load(path string) (Config, error) {
	// If file doesn't exist, return defaults
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Default(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	// Apply defaults for missing fields
	def := Default()
	if cfg.DBPath == "" {
		cfg.DBPath = def.DBPath
	}
	if cfg.LogPath == "" {
		cfg.LogPath = def.LogPath
	}
	if cfg.GUI.Width == 0 {
		cfg.GUI = def.GUI
	}

	return cfg, nil
}

// Save writes configuration to a JSON file.
func (c Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
