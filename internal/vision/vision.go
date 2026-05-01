// Package vision provides AI vision provider management for image analysis.
//
// This package implements a multi-provider architecture for sending images
// to AI models and receiving text responses. It supports multiple provider
// types (OpenAI-compatible, Anthropic, HuggingFace) through a unified interface.
//
// Provider Selection:
//
// The Manager maintains a list of configured providers. The first provider
// in the list is the default and is used when no specific provider is requested.
// Callers can select a specific provider by name using the Get method.
//
// Example:
//
//	mgr, err := vision.NewManager(cfg.Vision)
//	if err != nil {
//	    return err
//	}
//
//	// Use default provider
//	result, err := mgr.Analyze(ctx, imageData, "What is in this image?")
//
//	// Use specific provider
//	result, err := mgr.AnalyzeWith(ctx, "ollama", imageData, "Describe this image")
package vision

import (
	"context"
	"fmt"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/config"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
)

// Provider defines the interface for AI vision providers.
//
// A Provider can analyze an image and return a text response based on
// a prompt. The image is provided as PNG-encoded bytes.
type Provider interface {
	// Name returns the unique identifier for this provider.
	Name() string

	// Analyze sends the image and prompt to the AI model and returns
	// the text response. The image should be PNG-encoded bytes.
	Analyze(ctx context.Context, image []byte, prompt string) (string, error)
}

// ImageComparer extends Provider with the ability to compare two images.
//
// Not all providers support multi-image input. Providers that implement
// this interface can compare two images side by side.
type ImageComparer interface {
	Provider

	// CompareImages sends two images and a prompt to the AI model and
	// returns a text response describing the comparison.
	CompareImages(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error)
}

// ProviderInfo contains metadata about a configured provider.
type ProviderInfo struct {
	// Name is the unique identifier for this provider.
	Name string

	// Model is the AI model name used by this provider.
	Model string

	// IsDefault indicates whether this is the default provider.
	IsDefault bool
}

// Manager manages multiple vision providers.
//
// Manager holds a list of providers and provides methods to select
// and use them. The first provider in the list is the default.
type Manager struct {
	providers       []Provider
	defaultProvider Provider
	providerMap     map[string]Provider
}

// NewManager creates a Manager from configuration.
//
// If cfg is nil or has no providers, returns (nil, nil) to indicate
// that vision functionality is not configured.
//
// Returns an error if a provider type is not recognized or if provider
// names are duplicated.
func NewManager(cfg *config.VisionConfig) (*Manager, error) {
	if cfg == nil || len(cfg.Providers) == 0 {
		logging.Debug().Msg("No vision providers configured")
		return nil, nil
	}

	logging.Info().Int("count", len(cfg.Providers)).Msg("Initializing vision providers")

	m := &Manager{
		providerMap: make(map[string]Provider),
	}

	for _, pc := range cfg.Providers {
		logging.Debug().Str("provider", pc.Name).Str("type", pc.Type).Str("model", pc.Model).Int("timeout", pc.DefaultTimeout()).Msg("Creating provider")
		p, err := newProvider(pc)
		if err != nil {
			logging.Error().Str("provider", pc.Name).Err(err).Msg("Failed to create provider")
			return nil, fmt.Errorf("failed to create provider %q: %w", pc.Name, err)
		}

		if _, exists := m.providerMap[pc.Name]; exists {
			return nil, fmt.Errorf("duplicate provider name: %q", pc.Name)
		}

		m.providers = append(m.providers, p)
		m.providerMap[pc.Name] = p
		logging.Info().Str("provider", pc.Name).Str("type", pc.Type).Str("model", pc.Model).Msg("Provider initialized")
	}

	m.defaultProvider = m.providers[0]
	logging.Info().Str("default", m.defaultProvider.Name()).Msg("Default vision provider set")
	return m, nil
}

// Default returns the default provider (first in the list).
//
// Returns nil if no providers are configured.
func (m *Manager) Default() Provider {
	if m == nil {
		return nil
	}
	return m.defaultProvider
}

// Get returns a provider by name.
//
// If name is empty, returns the default provider.
// Returns nil if the provider is not found or no providers are configured.
func (m *Manager) Get(name string) Provider {
	if m == nil {
		return nil
	}
	if name == "" {
		return m.defaultProvider
	}
	return m.providerMap[name]
}

// Providers returns metadata for all configured providers.
func (m *Manager) Providers() []ProviderInfo {
	if m == nil {
		return nil
	}

	result := make([]ProviderInfo, 0, len(m.providers))
	for _, p := range m.providers {
		info := ProviderInfo{
			Name:      p.Name(),
			IsDefault: p == m.defaultProvider,
		}
		if mp, ok := p.(interface{ ModelName() string }); ok {
			info.Model = mp.ModelName()
		}
		result = append(result, info)
	}
	return result
}

// Analyze uses the default provider to analyze an image.
//
// Returns an error if no providers are configured.
func (m *Manager) Analyze(ctx context.Context, image []byte, prompt string) (string, error) {
	if m == nil || m.defaultProvider == nil {
		return "", fmt.Errorf("no vision providers configured")
	}
	return m.defaultProvider.Analyze(ctx, image, prompt)
}

// AnalyzeWith uses a specific provider to analyze an image.
//
// If name is empty, uses the default provider.
// Returns an error if the provider is not found.
func (m *Manager) AnalyzeWith(ctx context.Context, name string, image []byte, prompt string) (string, error) {
	if m == nil {
		return "", fmt.Errorf("no vision providers configured")
	}

	p := m.Get(name)
	if p == nil {
		logging.Error().Str("provider", name).Msg("Provider not found")
		return "", fmt.Errorf("provider %q not found", name)
	}

	logging.Debug().Str("provider", p.Name()).Int("image_size", len(image)).Str("prompt_preview", truncatePrompt(prompt)).Msg("Sending image to provider")
	result, err := p.Analyze(ctx, image, prompt)
	if err != nil {
		logging.Error().Str("provider", p.Name()).Err(err).Msg("Provider analysis failed")
		return "", err
	}

	logging.Debug().Str("provider", p.Name()).Int("response_size", len(result)).Msg("Provider analysis complete")
	return result, nil
}

// CompareImages uses a specific provider to compare two images.
//
// If name is empty, uses the default provider.
// Returns an error if the provider does not support image comparison
// or if the provider is not found.
func (m *Manager) CompareImages(ctx context.Context, name string, image1 []byte, image2 []byte, prompt string) (string, error) {
	if m == nil {
		return "", fmt.Errorf("no vision providers configured")
	}

	p := m.Get(name)
	if p == nil {
		logging.Error().Str("provider", name).Msg("Provider not found")
		return "", fmt.Errorf("provider %q not found", name)
	}

	cp, ok := p.(ImageComparer)
	if !ok {
		return "", fmt.Errorf("provider %q does not support image comparison", p.Name())
	}

	logging.Debug().Str("provider", p.Name()).Int("image1_size", len(image1)).Int("image2_size", len(image2)).Str("prompt_preview", truncatePrompt(prompt)).Msg("Sending images to provider for comparison")
	result, err := cp.CompareImages(ctx, image1, image2, prompt)
	if err != nil {
		logging.Error().Str("provider", p.Name()).Err(err).Msg("Provider image comparison failed")
		return "", err
	}

	logging.Debug().Str("provider", p.Name()).Int("response_size", len(result)).Msg("Provider image comparison complete")
	return result, nil
}

// newProvider creates a Provider from configuration.
//
// This function is the factory that dispatches to the correct provider
// implementation based on the Type field.
func newProvider(cfg config.VisionProviderConfig) (Provider, error) {
	switch cfg.Type {
	case "openai-compatible":
		return newOpenAICompatibleProvider(cfg)
	case "anthropic":
		return newAnthropicProvider(cfg)
	case "huggingface":
		return newHuggingFaceProvider(cfg)
	default:
		return nil, fmt.Errorf("unknown provider type: %q", cfg.Type)
	}
}

// truncatePrompt returns a truncated version of the prompt for logging.
func truncatePrompt(prompt string) string {
	if len(prompt) <= 50 {
		return prompt
	}
	return prompt[:47] + "..."
}
