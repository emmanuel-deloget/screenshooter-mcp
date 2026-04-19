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
	Listen  string `json:"listen"`
}

func DefaultConfig() *Config {
	return &Config{
		LogLevel: "info",
		Color:    "auto",
		Listen:  "",
	}
}

func (c *Config) Path() string {
	userDir, _ := userConfigDir()
	if userDir != "" {
		return filepath.Join(userDir, "config.json")
	}
	return ""
}

func Load(configPath string) (*Config, error) {
	if configPath != "" {
		return loadFrom(configPath)
	}

	if envPath := os.Getenv("SCREENSHOOTER_CONFIG"); envPath != "" {
		return loadFrom(envPath)
	}

	userDir, err := userConfigDir()
	if err == nil && userDir != "" {
		if cfg, err := loadFrom(filepath.Join(userDir, "config.json")); err == nil {
			return cfg, nil
		}
	}

	systemDir := "/etc/screenshooter-mcp"
	if cfg, err := loadFrom(filepath.Join(systemDir, "config.json")); err == nil {
		return cfg, nil
	}

	return DefaultConfig(), nil
}

func loadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return cfg, nil
}

func userConfigDir() (string, error) {
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig != "" {
		return filepath.Join(xdgConfig, "screenshooter-mcp"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine user home directory: %w", err)
	}
	return filepath.Join(home, ".config", "screenshooter-mcp"), nil
}

func (c *Config) Save(path string) error {
	userDir, err := userConfigDir()
	if err != nil {
		return fmt.Errorf("cannot determine user config directory: %w", err)
	}

	if path == "" {
		path = filepath.Join(userDir, "config.json")
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
