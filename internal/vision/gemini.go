package vision

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/genai"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/config"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
)

// geminiProvider implements Provider for Google's Gemini API and Vertex AI.
//
// This provider supports two modes:
//   - Gemini API: Uses an API key for direct access to Google's Gemini models
//   - Vertex AI: Uses GCP project/location with Application Default Credentials
//
// Mode is determined by configuration: if project and location are set,
// Vertex AI mode is used. Otherwise, Gemini API mode is used.
type geminiProvider struct {
	name     string
	model    string
	apiKey   string
	project  string
	location string
	timeout  int
}

func newGeminiProvider(cfg config.VisionProviderConfig) (Provider, error) {
	if cfg.Model == "" {
		return nil, fmt.Errorf("model is required for gemini provider")
	}

	isVertexAI := cfg.Project != "" && cfg.Location != ""

	if !isVertexAI && cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is required for gemini provider (Gemini API mode)")
	}

	if isVertexAI && cfg.APIKey == "" {
		logging.Warn().Str("provider", cfg.Name).Msg("Vertex AI mode: using Application Default Credentials")
	}

	return &geminiProvider{
		name:     cfg.Name,
		model:    cfg.Model,
		apiKey:   cfg.APIKey,
		project:  cfg.Project,
		location: cfg.Location,
		timeout:  cfg.DefaultTimeout(),
	}, nil
}

func (p *geminiProvider) Name() string {
	return p.name
}

func (p *geminiProvider) ModelName() string {
	return p.model
}

func (p *geminiProvider) createClient(ctx context.Context) (*genai.Client, error) {
	isVertexAI := p.project != "" && p.location != ""

	var cc *genai.ClientConfig

	if isVertexAI {
		logging.Debug().Str("provider", p.name).Str("project", p.project).Str("location", p.location).Msg("Creating Vertex AI client")
		cc = &genai.ClientConfig{
			Project:  p.project,
			Location: p.location,
			Backend:  genai.BackendVertexAI,
		}
		if p.apiKey != "" {
			cc.APIKey = p.apiKey
		}
	} else {
		logging.Debug().Str("provider", p.name).Msg("Creating Gemini API client")
		cc = &genai.ClientConfig{
			APIKey:  p.apiKey,
			Backend: genai.BackendGeminiAPI,
		}
	}

	client, err := genai.NewClient(ctx, cc)
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	return client, nil
}

func (p *geminiProvider) Analyze(ctx context.Context, image []byte, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.timeout)*time.Second)
	defer cancel()

	client, err := p.createClient(ctx)
	if err != nil {
		return "", err
	}

	logging.Debug().Str("provider", p.name).Str("model", p.model).Int("timeout", p.timeout).Msg("Sending request to Gemini API")

	contents := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: prompt},
				{InlineData: &genai.Blob{Data: image, MIMEType: "image/png"}},
			},
		},
	}

	result, err := client.Models.GenerateContent(ctx, p.model, contents, nil)
	if err != nil {
		logging.Error().Str("provider", p.name).Str("model", p.model).Err(err).Msg("Gemini API request failed")
		return "", fmt.Errorf("content generation failed: %w", err)
	}

	text := result.Text()
	if text == "" {
		logging.Warn().Str("provider", p.name).Str("model", p.model).Msg("No text content in response")
		return "", fmt.Errorf("no text content in response from provider %s", p.name)
	}

	logging.Debug().Str("provider", p.name).Str("model", p.model).Int("response_size", len(text)).Msg("Gemini API response received")
	return text, nil
}

func (p *geminiProvider) CompareImages(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.timeout)*time.Second)
	defer cancel()

	client, err := p.createClient(ctx)
	if err != nil {
		return "", err
	}

	logging.Debug().Str("provider", p.name).Str("model", p.model).Int("image1_size", len(image1)).Int("image2_size", len(image2)).Msg("Sending comparison request to Gemini API")

	contents := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: prompt},
				{InlineData: &genai.Blob{Data: image1, MIMEType: "image/png"}},
				{InlineData: &genai.Blob{Data: image2, MIMEType: "image/png"}},
			},
		},
	}

	result, err := client.Models.GenerateContent(ctx, p.model, contents, nil)
	if err != nil {
		logging.Error().Str("provider", p.name).Str("model", p.model).Err(err).Msg("Gemini API comparison request failed")
		return "", fmt.Errorf("content generation failed: %w", err)
	}

	text := result.Text()
	if text == "" {
		logging.Warn().Str("provider", p.name).Str("model", p.model).Msg("No text content in response")
		return "", fmt.Errorf("no text content in response from provider %s", p.name)
	}

	logging.Debug().Str("provider", p.name).Str("model", p.model).Int("response_size", len(text)).Msg("Gemini API comparison response received")
	return text, nil
}
