package vision

import (
	"context"
	"fmt"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/config"
)

// anthropicProvider implements Provider for Anthropic's Claude API.
//
// This provider uses the Anthropic Messages API to send images and
// receive text responses. It supports all Claude models with vision
// capabilities.
type anthropicProvider struct {
	name      string
	model     string
	apiKey    string
	baseURL   string
	timeout   int
}

func newAnthropicProvider(cfg config.VisionProviderConfig) (Provider, error) {
	if cfg.Model == "" {
		return nil, fmt.Errorf("model is required for anthropic provider")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is required for anthropic provider")
	}
	return &anthropicProvider{
		name:    cfg.Name,
		model:   cfg.Model,
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		timeout: cfg.DefaultTimeout(),
	}, nil
}

func (p *anthropicProvider) Name() string {
	return p.name
}

func (p *anthropicProvider) ModelName() string {
	return p.model
}

func (p *anthropicProvider) Analyze(ctx context.Context, image []byte, prompt string) (string, error) {
	return "", fmt.Errorf("anthropic provider not yet implemented")
}
