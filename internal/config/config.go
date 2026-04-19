package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	LogLevel string `json:"log_level"`
	Color    string `json:"color"`
}

func DefaultConfig() *Config {
	return &Config{
		LogLevel: "info",
		Color:    "auto",
	}
}

func (c *Config) Path() string {
	return filepath.Join(DefaultConfigDir(), "config.json")
}

func DefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".screenshooter-mcp"
	}
	return filepath.Join(home, ".local", "share", "screenshooter-mcp")
}

func Load(configPath string) (*Config, error) {
	config := DefaultConfig()

	if configPath == "" {
		configPath = os.Getenv("SCREENSHOOTER_CONFIG")
	}

	if configPath == "" {
		configPath = config.Path()
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

func (c *Config) Save(path string) error {
	if path == "" {
		path = c.Path()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
