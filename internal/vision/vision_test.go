package vision

import (
	"context"
	"testing"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/config"
)

type mockProvider struct {
	name    string
	model   string
	analyze func(ctx context.Context, image []byte, prompt string) (string, error)
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) ModelName() string {
	return m.model
}

func (m *mockProvider) Analyze(ctx context.Context, image []byte, prompt string) (string, error) {
	return m.analyze(ctx, image, prompt)
}

func TestNewManagerNilConfig(t *testing.T) {
	m, err := NewManager(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m != nil {
		t.Error("expected nil manager for nil config")
	}
}

func TestNewManagerEmptyProviders(t *testing.T) {
	cfg := &config.VisionConfig{Providers: []config.VisionProviderConfig{}}
	m, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m != nil {
		t.Error("expected nil manager for empty providers")
	}
}

func TestNewManagerUnknownType(t *testing.T) {
	cfg := &config.VisionConfig{
		Providers: []config.VisionProviderConfig{
			{Name: "unknown", Type: "unknown-type", Model: "test"},
		},
	}
	_, err := NewManager(cfg)
	if err == nil {
		t.Fatal("expected error for unknown provider type")
	}
}

func TestManagerGetAndDefault(t *testing.T) {
	m := &Manager{
		providers: []Provider{
			&mockProvider{name: "first", model: "model-a"},
			&mockProvider{name: "second", model: "model-b"},
		},
		providerMap: map[string]Provider{
			"first":  &mockProvider{name: "first", model: "model-a"},
			"second": &mockProvider{name: "second", model: "model-b"},
		},
	}
	m.defaultProvider = m.providers[0]

	if m.Default().Name() != "first" {
		t.Errorf("expected default provider 'first', got '%s'", m.Default().Name())
	}

	if m.Get("").Name() != "first" {
		t.Errorf("expected Get('') to return default, got '%s'", m.Get("").Name())
	}

	if m.Get("second").Name() != "second" {
		t.Errorf("expected Get('second') to return 'second', got '%s'", m.Get("second").Name())
	}

	if m.Get("nonexistent") != nil {
		t.Error("expected nil for nonexistent provider")
	}
}

func TestManagerProviders(t *testing.T) {
	m := &Manager{
		providers: []Provider{
			&mockProvider{name: "first", model: "model-a"},
			&mockProvider{name: "second", model: "model-b"},
		},
		providerMap: map[string]Provider{
			"first":  &mockProvider{name: "first", model: "model-a"},
			"second": &mockProvider{name: "second", model: "model-b"},
		},
	}
	m.defaultProvider = m.providers[0]

	infos := m.Providers()
	if len(infos) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(infos))
	}

	if !infos[0].IsDefault {
		t.Error("expected first provider to be default")
	}
	if infos[1].IsDefault {
		t.Error("expected second provider to not be default")
	}
	if infos[0].Model != "model-a" {
		t.Errorf("expected model 'model-a', got '%s'", infos[0].Model)
	}
}

func TestManagerAnalyze(t *testing.T) {
	m := &Manager{
		providers: []Provider{
			&mockProvider{name: "test", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "response", nil
			}},
		},
		providerMap: map[string]Provider{
			"test": &mockProvider{name: "test", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "response", nil
			}},
		},
	}
	m.defaultProvider = m.providers[0]

	result, err := m.Analyze(context.Background(), []byte("image"), "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "response" {
		t.Errorf("expected 'response', got '%s'", result)
	}
}

func TestManagerAnalyzeWith(t *testing.T) {
	m := &Manager{
		providers: []Provider{
			&mockProvider{name: "first", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "first-response", nil
			}},
			&mockProvider{name: "second", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "second-response", nil
			}},
		},
		providerMap: map[string]Provider{
			"first": &mockProvider{name: "first", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "first-response", nil
			}},
			"second": &mockProvider{name: "second", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "second-response", nil
			}},
		},
	}
	m.defaultProvider = m.providers[0]

	result, err := m.AnalyzeWith(context.Background(), "second", []byte("image"), "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "second-response" {
		t.Errorf("expected 'second-response', got '%s'", result)
	}

	_, err = m.AnalyzeWith(context.Background(), "nonexistent", []byte("image"), "prompt")
	if err == nil {
		t.Fatal("expected error for nonexistent provider")
	}
}

func TestNilManagerMethods(t *testing.T) {
	var m *Manager

	if m.Default() != nil {
		t.Error("expected nil from nil manager Default()")
	}
	if m.Get("test") != nil {
		t.Error("expected nil from nil manager Get()")
	}
	if m.Providers() != nil {
		t.Error("expected nil from nil manager Providers()")
	}

	_, err := m.Analyze(context.Background(), []byte("image"), "prompt")
	if err == nil {
		t.Fatal("expected error from nil manager Analyze()")
	}

	_, err = m.AnalyzeWith(context.Background(), "test", []byte("image"), "prompt")
	if err == nil {
		t.Fatal("expected error from nil manager AnalyzeWith()")
	}
}
