package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestVisionConfigParsing(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantErr  bool
		validate func(t *testing.T, cfg *Config)
	}{
		{
			name:    "empty config",
			json:    `{}`,
			wantErr:  false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Vision != nil {
					t.Error("expected nil Vision for empty config")
				}
			},
		},
		{
			name: "single openai-compatible provider",
			json: `{
				"vision": {
					"providers": [
						{
							"name": "ollama",
							"type": "openai-compatible",
							"base_url": "http://localhost:11434/v1",
							"model": "llava:7b",
							"timeout": 30
						}
					]
				}
			}`,
			wantErr:  false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Vision == nil {
					t.Fatal("expected non-nil Vision")
				}
				if len(cfg.Vision.Providers) != 1 {
					t.Fatalf("expected 1 provider, got %d", len(cfg.Vision.Providers))
				}
				p := cfg.Vision.Providers[0]
				if p.Name != "ollama" {
					t.Errorf("expected name 'ollama', got '%s'", p.Name)
				}
				if p.Type != "openai-compatible" {
					t.Errorf("expected type 'openai-compatible', got '%s'", p.Type)
				}
				if p.BaseURL != "http://localhost:11434/v1" {
					t.Errorf("expected base_url 'http://localhost:11434/v1', got '%s'", p.BaseURL)
				}
				if p.Model != "llava:7b" {
					t.Errorf("expected model 'llava:7b', got '%s'", p.Model)
				}
				if p.Timeout != 30 {
					t.Errorf("expected timeout 30, got %d", p.Timeout)
				}
			},
		},
		{
			name: "multiple providers with defaults",
			json: `{
				"vision": {
					"providers": [
						{
							"name": "openai",
							"type": "openai-compatible",
							"model": "gpt-4o",
							"api_key": "sk-test"
						},
						{
							"name": "claude",
							"type": "anthropic",
							"model": "claude-sonnet-4-20250514",
							"api_key": "sk-ant-test"
						}
					]
				}
			}`,
			wantErr:  false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Vision == nil {
					t.Fatal("expected non-nil Vision")
				}
				if len(cfg.Vision.Providers) != 2 {
					t.Fatalf("expected 2 providers, got %d", len(cfg.Vision.Providers))
				}
				if cfg.Vision.Providers[0].Name != "openai" {
					t.Errorf("expected first provider 'openai', got '%s'", cfg.Vision.Providers[0].Name)
				}
				if cfg.Vision.Providers[1].Name != "claude" {
					t.Errorf("expected second provider 'claude', got '%s'", cfg.Vision.Providers[1].Name)
				}
			},
		},
		{
			name: "timeout defaults via DefaultTimeout",
			json: `{
				"vision": {
					"providers": [
						{
							"name": "no-timeout",
							"type": "openai-compatible",
							"model": "test"
						},
						{
							"name": "zero-timeout",
							"type": "openai-compatible",
							"model": "test",
							"timeout": 0
						},
						{
							"name": "custom-timeout",
							"type": "openai-compatible",
							"model": "test",
							"timeout": 45
						}
					]
				}
			}`,
			wantErr:  false,
			validate: func(t *testing.T, cfg *Config) {
				providers := cfg.Vision.Providers
				if providers[0].DefaultTimeout() != 20 {
					t.Errorf("expected default timeout 20, got %d", providers[0].DefaultTimeout())
				}
				if providers[1].DefaultTimeout() != 20 {
					t.Errorf("expected default timeout 20 for zero, got %d", providers[1].DefaultTimeout())
				}
				if providers[2].DefaultTimeout() != 45 {
					t.Errorf("expected custom timeout 45, got %d", providers[2].DefaultTimeout())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			err := json.Unmarshal([]byte(tt.json), &cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, &cfg)
			}
		})
	}
}

func TestDefaultConfigHasNilVision(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Vision != nil {
		t.Error("DefaultConfig() should have nil Vision")
	}
}

func TestConfigPath(t *testing.T) {
	tests := []struct {
		name      string
		xdgConfig string
		home      string
		want      string
	}{
		{
			name:      "XDG_CONFIG_HOME set",
			xdgConfig: "/custom/config",
			want:       "/custom/config/screenshooter-mcp/config.json",
		},
		{
			name: "use home directory",
			home: "/home/testuser",
			want: "/home/testuser/.config/screenshooter-mcp/config.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.xdgConfig != "" {
				t.Setenv("XDG_CONFIG_HOME", tt.xdgConfig)
			}
			if tt.home != "" {
				t.Setenv("HOME", tt.home)
			}

			cfg := &Config{}
			got := cfg.Path()
			if got != tt.want {
				t.Errorf("Config.Path() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.json")
	configData := `{"log_level": "debug", "listen": "127.0.0.1:9999"}`
	err := os.WriteFile(configPath, []byte(configData), 0644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	t.Setenv("SCREENSHOOTER_CONFIG", configPath)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() with env var error = %v", err)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("Load() LogLevel = %v, want 'debug'", cfg.LogLevel)
	}
	if cfg.Listen != "127.0.0.1:9999" {
		t.Errorf("Load() Listen = %v, want '127.0.0.1:9999'", cfg.Listen)
	}
}

func TestLoadFromUserConfig(t *testing.T) {
	// Create user config directory and file
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "screenshooter-mcp")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	configData := `{"log_level": "warn", "color": "never"}`
	err = os.WriteFile(configPath, []byte(configData), 0644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Temporarily change home directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)
	// Clear XDG_CONFIG_HOME to force using HOME
	oldXdg := os.Getenv("XDG_CONFIG_HOME")
	os.Unsetenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXdg)
	// Clear SCREENSHOOTER_CONFIG
	oldConfig := os.Getenv("SCREENSHOOTER_CONFIG")
	os.Unsetenv("SCREENSHOOTER_CONFIG")
	defer os.Setenv("SCREENSHOOTER_CONFIG", oldConfig)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() with user config error = %v", err)
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("Load() LogLevel = %v, want 'warn'", cfg.LogLevel)
	}
	if cfg.Color != "never" {
		t.Errorf("Load() Color = %v, want 'never'", cfg.Color)
	}
}

func TestLoadNoConfig(t *testing.T) {
	// Ensure no config files exist
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "empty"))
	t.Setenv("HOME", filepath.Join(tmpDir, "empty-home"))

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() with no config error = %v", err)
	}
	// Should return DefaultConfig
	if cfg.LogLevel != "info" {
		t.Errorf("Load() LogLevel = %v, want 'info'", cfg.LogLevel)
	}
	if cfg.Color != "auto" {
		t.Errorf("Load() Color = %v, want 'auto'", cfg.Color)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &Config{
		LogLevel: "error",
		Color:    "always",
		Listen:   "0.0.0.0:11777",
	}

	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load the saved config
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() after Save() error = %v", err)
	}

	if loaded.LogLevel != "error" {
		t.Errorf("loaded LogLevel = %v, want 'error'", loaded.LogLevel)
	}
	if loaded.Color != "always" {
		t.Errorf("loaded Color = %v, want 'always'", loaded.Color)
	}
	if loaded.Listen != "0.0.0.0:11777" {
		t.Errorf("loaded Listen = %v, want '0.0.0.0:11777'", loaded.Listen)
	}
}

func TestLoadFromInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")
	err := os.WriteFile(configPath, []byte("not valid json"), 0644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err = Load(configPath)
	if err == nil {
		t.Error("Load() with invalid JSON should return error")
	}
}

func TestLoadWithExplicitPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "explicit.json")
	configData := `{"log_level": "error", "color": "always"}`
	err := os.WriteFile(configPath, []byte(configData), 0644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() with explicit path error = %v", err)
	}
	if cfg.LogLevel != "error" {
		t.Errorf("Load() LogLevel = %v, want 'error'", cfg.LogLevel)
	}
}

func TestSaveCreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir1", "subdir2", "config.json")

	cfg := &Config{
		LogLevel: "info",
	}

	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Save() should create directories and file")
	}
}
