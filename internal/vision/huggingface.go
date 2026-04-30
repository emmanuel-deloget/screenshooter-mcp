package vision

import (
	"context"
	"fmt"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/config"
)

// huggingFaceProvider implements Provider for HuggingFace Inference API.
//
// This provider uses the HuggingFace Inference API to send images and
// receive text responses via direct HTTP calls.
type huggingFaceProvider struct {
	name    string
	model   string
	apiKey  string
	baseURL string
	timeout int
}

func newHuggingFaceProvider(cfg config.VisionProviderConfig) (Provider, error) {
	if cfg.Model == "" {
		return nil, fmt.Errorf("model is required for huggingface provider")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is required for huggingface provider")
	}
	return &huggingFaceProvider{
		name:    cfg.Name,
		model:   cfg.Model,
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		timeout: cfg.DefaultTimeout(),
	}, nil
}

func (p *huggingFaceProvider) Name() string {
	return p.name
}

func (p *huggingFaceProvider) ModelName() string {
	return p.model
}

func (p *huggingFaceProvider) Analyze(ctx context.Context, image []byte, prompt string) (string, error) {
	return "", fmt.Errorf("huggingface provider not yet implemented")
}
