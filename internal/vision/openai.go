package vision

import (
	"context"
	"fmt"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/config"
)

// openAICompatibleProvider implements Provider for OpenAI-compatible APIs.
//
// This provider works with any API that follows the OpenAI Chat Completions format,
// including OpenAI, Ollama, Mistral, Groq, and other compatible services.
type openAICompatibleProvider struct {
	name    string
	model   string
	baseURL string
	apiKey  string
	timeout int
}

func newOpenAICompatibleProvider(cfg config.VisionProviderConfig) (Provider, error) {
	if cfg.Model == "" {
		return nil, fmt.Errorf("model is required for openai-compatible provider")
	}
	return &openAICompatibleProvider{
		name:    cfg.Name,
		model:   cfg.Model,
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		timeout: cfg.DefaultTimeout(),
	}, nil
}

func (p *openAICompatibleProvider) Name() string {
	return p.name
}

func (p *openAICompatibleProvider) ModelName() string {
	return p.model
}

func (p *openAICompatibleProvider) Analyze(ctx context.Context, image []byte, prompt string) (string, error) {
	return "", fmt.Errorf("openai-compatible provider not yet implemented")
}
