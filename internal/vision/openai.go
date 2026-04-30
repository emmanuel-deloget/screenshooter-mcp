package vision

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	openai "github.com/sashabaranov/go-openai"

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
	cfg := openai.DefaultConfig(p.apiKey)
	if p.baseURL != "" {
		cfg.BaseURL = p.baseURL
	}
	client := openai.NewClientWithConfig(cfg)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.timeout)*time.Second)
	defer cancel()

	base64Image := base64.StdEncoding.EncodeToString(image)

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: prompt,
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL:    "data:image/png;base64," + base64Image,
							Detail: openai.ImageURLDetailAuto,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("chat completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from provider %s", p.name)
	}

	return resp.Choices[0].Message.Content, nil
}
