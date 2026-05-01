package vision

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/config"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
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
	logging.Debug().Str("provider", p.name).Str("model", p.model).Str("base_url", p.baseURL).Int("timeout", p.timeout).Msg("Sending request to Anthropic API")

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
		logging.Error().Str("provider", p.name).Str("model", p.model).Err(err).Msg("Anthropic API request failed")
		return "", fmt.Errorf("message completion failed: %w", err)
	}

	if len(msg.Content) == 0 {
		logging.Warn().Str("provider", p.name).Str("model", p.model).Msg("No content in response")
		return "", fmt.Errorf("no response from provider %s", p.name)
	}

	var result string
	for _, block := range msg.Content {
		if block.Type == "text" {
			result += block.Text
		}
	}

	if result == "" {
		logging.Warn().Str("provider", p.name).Str("model", p.model).Msg("No text content in response")
		return "", fmt.Errorf("no text content in response from provider %s", p.name)
	}

	logging.Debug().Str("provider", p.name).Str("model", p.model).Int("response_size", len(result)).Msg("Anthropic API response received")
	return result, nil
}

func (p *anthropicProvider) CompareImages(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
	opts := []option.RequestOption{
		option.WithAPIKey(p.apiKey),
	}
	if p.baseURL != "" {
		opts = append(opts, option.WithBaseURL(p.baseURL))
	}

	client := anthropic.NewClient(opts...)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.timeout)*time.Second)
	defer cancel()

	base64Image1 := base64.StdEncoding.EncodeToString(image1)
	base64Image2 := base64.StdEncoding.EncodeToString(image2)
	logging.Debug().Str("provider", p.name).Str("model", p.model).Int("image1_size", len(image1)).Int("image2_size", len(image2)).Msg("Sending comparison request to Anthropic API")

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
								Data:      base64Image1,
								MediaType: anthropic.Base64ImageSourceMediaTypeImagePNG,
							},
						},
					},
				},
				anthropic.ContentBlockParamUnion{
					OfImage: &anthropic.ImageBlockParam{
						Source: anthropic.ImageBlockParamSourceUnion{
							OfBase64: &anthropic.Base64ImageSourceParam{
								Data:      base64Image2,
								MediaType: anthropic.Base64ImageSourceMediaTypeImagePNG,
							},
						},
					},
				},
			),
		},
	})
	if err != nil {
		logging.Error().Str("provider", p.name).Str("model", p.model).Err(err).Msg("Anthropic API comparison request failed")
		return "", fmt.Errorf("message completion failed: %w", err)
	}

	if len(msg.Content) == 0 {
		logging.Warn().Str("provider", p.name).Str("model", p.model).Msg("No content in response")
		return "", fmt.Errorf("no response from provider %s", p.name)
	}

	var result string
	for _, block := range msg.Content {
		if block.Type == "text" {
			result += block.Text
		}
	}

	if result == "" {
		logging.Warn().Str("provider", p.name).Str("model", p.model).Msg("No text content in response")
		return "", fmt.Errorf("no text content in response from provider %s", p.name)
	}

	logging.Debug().Str("provider", p.name).Str("model", p.model).Int("response_size", len(result)).Msg("Anthropic API comparison response received")
	return result, nil
}
