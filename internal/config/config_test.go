package config

import (
	"encoding/json"
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
			wantErr: false,
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
			wantErr: false,
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
			wantErr: false,
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
			wantErr: false,
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
