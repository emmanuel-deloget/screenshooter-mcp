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
		return nil, nil
	}

	m := &Manager{
		providerMap: make(map[string]Provider),
	}

	for _, pc := range cfg.Providers {
		p, err := newProvider(pc)
		if err != nil {
			return nil, fmt.Errorf("failed to create provider %q: %w", pc.Name, err)
		}

		if _, exists := m.providerMap[pc.Name]; exists {
			return nil, fmt.Errorf("duplicate provider name: %q", pc.Name)
		}

		m.providers = append(m.providers, p)
		m.providerMap[pc.Name] = p
	}

	m.defaultProvider = m.providers[0]
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
		return "", fmt.Errorf("provider %q not found", name)
	}
	return p.Analyze(ctx, image, prompt)
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
