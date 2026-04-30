package vision

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/config"
)

// anthropicProvider implements Provider for Anthropic's Claude API.
//
// This provider uses the Anthropic Messages API to send images and
// receive text responses. It supports all Claude models with vision
// capabilities.
type anthropicProvider struct {
	name    string
	model   string
	apiKey  string
	baseURL string
	timeout int
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
	opts := []option.RequestOption{
		option.WithAPIKey(p.apiKey),
	}
	if p.baseURL != "" {
		opts = append(opts, option.WithBaseURL(p.baseURL))
	}

	client := anthropic.NewClient(opts...)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.timeout)*time.Second)
	defer cancel()

	base64Image := base64.StdEncoding.EncodeToString(image)

	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     p.model,
		MaxTokens: 4096,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.ContentBlockParamUnion{
					OfText: &anthropic.TextBlockParam{
						Text: prompt,
					},
				},
				anthropic.ContentBlockParamUnion{
					OfImage: &anthropic.ImageBlockParam{
						Source: anthropic.ImageBlockParamSourceUnion{
							OfBase64: &anthropic.Base64ImageSourceParam{
								Data:      base64Image,
								MediaType: anthropic.Base64ImageSourceMediaTypeImagePNG,
							},
						},
					},
				},
			),
		},
	})
	if err != nil {
		return "", fmt.Errorf("message completion failed: %w", err)
	}

	if len(msg.Content) == 0 {
		return "", fmt.Errorf("no response from provider %s", p.name)
	}

	var result string
	for _, block := range msg.Content {
		if block.Type == "text" {
			result += block.Text
		}
	}

	if result == "" {
		return "", fmt.Errorf("no text content in response from provider %s", p.name)
	}

	return result, nil
}
