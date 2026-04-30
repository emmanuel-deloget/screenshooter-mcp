// Package config provides configuration loading and management for the MCP server.
//
// This package implements configuration management following the XDG Base Directory
// specification. Configuration is loaded from JSON files with the following
// precedence (first found wins):
//
//  1. Path specified as argument to Load (useful for testing)
//  2. Path in SCREENSHOOTER_CONFIG environment variable
//  3. $XDG_CONFIG_HOME/screenshooter-mcp/config.json
//  4. ~/.config/screenshooter-mcp/config.json (if XDG_CONFIG_HOME not set)
//  5. /etc/screenshooter-mcp/config.json (system fallback)
//
// If no configuration file is found, DefaultConfig is returned, which provides
// sensible defaults. The server does not require a configuration file to operate.
//
// Configuration Structure:
//
//	A Config struct contains configuration fields:
//	  - LogLevel: Controls logging verbosity (debug, info, warn, error)
//	  - Color: Controls colored output (always, never, auto)
//	  - Listen: TCP address for HTTP mode (empty = stdio mode)
//	  - Vision: Vision provider configuration for AI image analysis
//
// Example config.json:
//
//	{
//	  "log_level": "info",
//	  "color": "auto",
//	  "listen": "127.0.0.1:11777",
//	  "vision": {
//	    "providers": [
//	      {
//	        "name": "ollama",
//	        "type": "openai-compatible",
//	        "base_url": "http://localhost:11434/v1",
//	        "model": "llava:7b",
//	        "timeout": 30
//	      },
//	      {
//	        "name": "openai",
//	        "type": "openai-compatible",
//	        "model": "gpt-4o",
//	        "api_key": "sk-...",
//	        "timeout": 20
//	      }
//	    ]
//	  }
//	}
//
// The Save method writes configuration to a JSON file, creating directories
// as needed. This can be used to create an initial configuration file or
// persist modified settings.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the server configuration.
//
// Config contains all configurable settings for the MCP server.
// All fields are optional - if not specified, sensible defaults are used.
//
// The JSON tags correspond to the configuration file format.
// A minimal configuration file can be empty or contain only the fields
// you want to override.
type Config struct {
	// LogLevel controls the verbosity of logging output.
	// Valid values: debug, info, warn, error
	// Default: info
	LogLevel string `json:"log_level"`

	// Color controls whether log output uses ANSI color codes.
	// Valid values: always, never, auto
	// Default: auto (use colors if terminal supports them)
	Color string `json:"color"`

	// Listen specifies the TCP address for HTTP mode.
	// If empty, the server runs in stdio mode.
	// Example: "127.0.0.1:11777"
	Listen string `json:"listen"`

	// Vision contains configuration for AI vision providers.
	// If empty or nil, vision tools are not available.
	Vision *VisionConfig `json:"vision,omitempty"`
}

// VisionConfig holds the configuration for AI vision providers.
//
// VisionConfig contains a list of providers that can analyze images.
// The first provider in the list is used as the default when no
// specific provider is requested.
type VisionConfig struct {
	// Providers is the list of configured vision providers.
	// The first provider is the default.
	Providers []VisionProviderConfig `json:"providers"`
}

// VisionProviderConfig configures a single vision provider.
//
// Each provider has a unique name, a type that determines the API
// protocol, and connection details. The timeout field controls
// how long to wait for a response before failing.
type VisionProviderConfig struct {
	// Name is the unique identifier for this provider.
	// Used to select a specific provider in tool calls.
	Name string `json:"name"`

	// Type specifies the API protocol to use.
	// Valid values: "openai-compatible", "anthropic", "huggingface"
	Type string `json:"type"`

	// BaseURL is the API endpoint URL.
	// For openai-compatible providers, this overrides the default URL.
	// For HuggingFace, this is the Inference API URL.
	// Optional for OpenAI and Anthropic (uses their default URLs).
	BaseURL string `json:"base_url,omitempty"`

	// Model specifies which model to use for analysis.
	// Examples: "gpt-4o", "llava:7b", "claude-sonnet-4-20250514"
	Model string `json:"model"`

	// APIKey is the authentication key for the provider.
	// Optional for local providers like Ollama.
	APIKey string `json:"api_key,omitempty"`

	// Timeout is the maximum time in seconds to wait for a response.
	// Default: 20 seconds if not specified.
	Timeout int `json:"timeout,omitempty"`
}

// DefaultTimeout returns the configured timeout or the default of 20 seconds.
func (p *VisionProviderConfig) DefaultTimeout() int {
	if p.Timeout <= 0 {
		return 20
	}
	return p.Timeout
}

// DefaultConfig returns a new Config with default values.
//
// This function creates a Config struct with all fields set to their
// default values. It serves as the fallback when no configuration
// file is found.
//
// Default values:
//   - LogLevel: "info"
//   - Color: "auto"
//   - Listen: "" (stdio mode)
func DefaultConfig() *Config {
	return &Config{
		LogLevel: "info",
		Color:    "auto",
		Listen:   "",
	}
}

// Path returns the default configuration file path for the current user.
//
// The path is determined by following the XDG Base Directory spec.
// If XDG_CONFIG_HOME is set, the path is $XDG_CONFIG_HOME/screenshooter-mcp/config.json.
// Otherwise, it returns ~/.config/screenshooter-mcp/config.json.
//
// If the user config directory cannot be determined (e.g., home directory
// not available), an empty string is returned. In this case, callers should
// handle the fallback to system configuration or default values.
func (c *Config) Path() string {
	userDir, _ := userConfigDir()
	if userDir != "" {
		return filepath.Join(userDir, "config.json")
	}
	return ""
}

// Load loads configuration from a file or falls back to defaults.
//
// The configPath argument specifies the path to a configuration file.
// If empty, the function searches for configuration in standard locations
// following XDG precedence (see package documentation).
//
// The search order is:
//  1. configPath if non-empty (allows explicit path specification)
//  2. SCREENSHOOTER_CONFIG environment variable
//  3. User config directory ($XDG_CONFIG_HOME/screenshooter-mcp/ or ~/.config/)
//  4. System config directory (/etc/screenshooter-mcp/)
//  5. DefaultConfig if no file found
//
// Returns the loaded configuration, or DefaultConfig if no file exists.
// Returns an error only if a file exists but cannot be parsed.
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

// loadFrom loads configuration from a specific file path.
//
// This is an internal helper function that reads and parses a JSON configuration file.
// It returns an error if the file cannot be read or parsed.
//
// The function applies defaults from DefaultConfig() before parsing,
// which allows partial configuration files to override only specific fields.
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

// userConfigDir returns the user's configuration directory path.
//
// This function follows the XDG Base Directory specification:
//   - If XDG_CONFIG_HOME is set, use that
//   - Otherwise, use ~/.config/screenshooter-mcp
//
// Returns an error if neither the environment variable nor the home
// directory is available.
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

// Save writes the configuration to a file.
//
// The path argument specifies where to save the configuration.
// If empty, the default user configuration path is used.
// Directories are created as needed.
//
// The configuration is written as indented JSON for readability.
// File permissions are set to 0644 (readable by all, writable by owner).
//
// Returns an error if:
//   - The directory cannot be created
//   - The configuration cannot be marshaled
//   - The file cannot be written
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
