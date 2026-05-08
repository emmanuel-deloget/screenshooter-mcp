package vision

import (
	"context"
	"fmt"
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

type mockComparer struct {
	name    string
	model   string
	analyze func(ctx context.Context, image []byte, prompt string) (string, error)
	compare func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error)
}

func (m *mockComparer) Name() string {
	return m.name
}

func (m *mockComparer) ModelName() string {
	return m.model
}

func (m *mockComparer) Analyze(ctx context.Context, image []byte, prompt string) (string, error) {
	return m.analyze(ctx, image, prompt)
}

func (m *mockComparer) CompareImages(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
	return m.compare(ctx, image1, image2, prompt)
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

	result, err := m.AnalyzeWith(context.Background(), "", []byte("image"), "prompt")
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

	result, err := m.analyzeWith(context.Background(), "second", []byte("image"), "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "second-response" {
		t.Errorf("expected 'second-response', got '%s'", result)
	}

	_, err = m.analyzeWith(context.Background(), "nonexistent", []byte("image"), "prompt")
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

	_, err := m.AnalyzeWith(context.Background(), "test", []byte("image"), "prompt")
	if err == nil {
		t.Fatal("expected error from nil manager AnalyzeWith()")
	}
}

func TestAnalyzeWithFallbackFirstFailsSecondSucceeds(t *testing.T) {
	m := &Manager{
		providers: []Provider{
			&mockProvider{name: "first", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "", fmt.Errorf("first provider error")
			}},
			&mockProvider{name: "second", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "second-response", nil
			}},
		},
		providerMap: map[string]Provider{
			"first": &mockProvider{name: "first", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "", fmt.Errorf("first provider error")
			}},
			"second": &mockProvider{name: "second", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "second-response", nil
			}},
		},
		enableFallback: true,
	}

	result, err := m.AnalyzeWith(context.Background(), "", []byte("image"), "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "second-response" {
		t.Errorf("expected 'second-response', got '%s'", result)
	}
}

func TestAnalyzeWithFallbackAllFail(t *testing.T) {
	m := &Manager{
		providers: []Provider{
			&mockProvider{name: "first", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "", fmt.Errorf("first provider error")
			}},
			&mockProvider{name: "second", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "", fmt.Errorf("second provider error")
			}},
		},
		providerMap: map[string]Provider{
			"first": &mockProvider{name: "first", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "", fmt.Errorf("first provider error")
			}},
			"second": &mockProvider{name: "second", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "", fmt.Errorf("second provider error")
			}},
		},
		enableFallback: true,
	}

	_, err := m.AnalyzeWith(context.Background(), "", []byte("image"), "prompt")
	if err == nil {
		t.Fatal("expected error when all providers fail")
	}
}

func TestAnalyzeWithNoFallbackOnError(t *testing.T) {
	m := &Manager{
		providers: []Provider{
			&mockProvider{name: "first", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "", fmt.Errorf("first provider error")
			}},
			&mockProvider{name: "second", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "second-response", nil
			}},
		},
		providerMap: map[string]Provider{
			"first": &mockProvider{name: "first", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "", fmt.Errorf("first provider error")
			}},
			"second": &mockProvider{name: "second", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "second-response", nil
			}},
		},
		enableFallback: false,
	}

	_, err := m.AnalyzeWith(context.Background(), "first", []byte("image"), "prompt")
	if err == nil {
		t.Fatal("expected error when fallback is disabled")
	}
}

func TestCompareImagesFallbackFirstFailsSecondSucceeds(t *testing.T) {
	m := &Manager{
		providers: []Provider{
			&mockComparer{
				name:    "first",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "", fmt.Errorf("first provider error")
				},
			},
			&mockComparer{
				name:    "second",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "second-response", nil
				},
			},
		},
		providerMap: map[string]Provider{
			"first": &mockComparer{
				name:    "first",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "", fmt.Errorf("first provider error")
				},
			},
			"second": &mockComparer{
				name:    "second",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "second-response", nil
				},
			},
		},
		enableFallback: true,
	}

	result, err := m.CompareImages(context.Background(), "", []byte("image1"), []byte("image2"), "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "second-response" {
		t.Errorf("expected 'second-response', got '%s'", result)
	}
}

func TestCompareImagesFallbackAllFail(t *testing.T) {
	m := &Manager{
		providers: []Provider{
			&mockComparer{
				name:    "first",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "", fmt.Errorf("first provider error")
				},
			},
			&mockComparer{
				name:    "second",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "", fmt.Errorf("second provider error")
				},
			},
		},
		providerMap: map[string]Provider{
			"first": &mockComparer{
				name:    "first",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "", fmt.Errorf("first provider error")
				},
			},
			"second": &mockComparer{
				name:    "second",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "", fmt.Errorf("second provider error")
				},
			},
		},
		enableFallback: true,
	}

	_, err := m.CompareImages(context.Background(), "", []byte("image1"), []byte("image2"), "prompt")
	if err == nil {
		t.Fatal("expected error when all providers fail")
	}
}

func TestCompareImagesNoFallbackOnError(t *testing.T) {
	m := &Manager{
		providers: []Provider{
			&mockComparer{
				name:    "first",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "", fmt.Errorf("first provider error")
				},
			},
			&mockComparer{
				name:    "second",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "second-response", nil
				},
			},
		},
		providerMap: map[string]Provider{
			"first": &mockComparer{
				name:    "first",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "", fmt.Errorf("first provider error")
				},
			},
			"second": &mockComparer{
				name:    "second",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "second-response", nil
				},
			},
		},
		enableFallback: false,
	}

	_, err := m.CompareImages(context.Background(), "first", []byte("image1"), []byte("image2"), "prompt")
	if err == nil {
		t.Fatal("expected error when fallback is disabled")
	}
}

func TestNewManagerDuplicateProviderName(t *testing.T) {
	cfg := &config.VisionConfig{
		Providers: []config.VisionProviderConfig{
			{Name: "dup", Type: "openai-compatible", Model: "model-a", BaseURL: "http://localhost:11434/v1"},
			{Name: "dup", Type: "openai-compatible", Model: "model-b", BaseURL: "http://localhost:11434/v1"},
		},
	}
	_, err := NewManager(cfg)
	if err == nil {
		t.Fatal("expected error for duplicate provider name")
	}
}

func TestNewManagerEnableFallback(t *testing.T) {
	cfg := &config.VisionConfig{
		EnableFallback: true,
		Providers: []config.VisionProviderConfig{
			{Name: "test", Type: "openai-compatible", Model: "model", BaseURL: "http://localhost:11434/v1"},
		},
	}
	m, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.enableFallback {
		t.Error("expected enableFallback to be true")
	}
}

func TestNewManagerDisableFallback(t *testing.T) {
	cfg := &config.VisionConfig{
		EnableFallback: false,
		Providers: []config.VisionProviderConfig{
			{Name: "test", Type: "openai-compatible", Model: "model", BaseURL: "http://localhost:11434/v1"},
		},
	}
	m, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.enableFallback {
		t.Error("expected enableFallback to be false")
	}
}

func TestCompareImagesNilManager(t *testing.T) {
	var m *Manager
	_, err := m.CompareImages(context.Background(), "", []byte("image1"), []byte("image2"), "prompt")
	if err == nil {
		t.Fatal("expected error from nil manager CompareImages()")
	}
}

func TestCompareImagesProviderNotComparer(t *testing.T) {
	m := &Manager{
		providers: []Provider{
			&mockProvider{name: "no-compare", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "response", nil
			}},
		},
		providerMap: map[string]Provider{
			"no-compare": &mockProvider{name: "no-compare", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "response", nil
			}},
		},
	}

	_, err := m.compareImages(context.Background(), "", []byte("image1"), []byte("image2"), "prompt")
	if err == nil {
		t.Fatal("expected error when provider does not support comparison")
	}
}

func TestCompareImagesNilManagerDirect(t *testing.T) {
	var m *Manager
	_, err := m.compareImages(context.Background(), "", []byte("image1"), []byte("image2"), "prompt")
	if err == nil {
		t.Fatal("expected error from nil manager compareImages()")
	}
}

func TestCompareImagesNamedProvider(t *testing.T) {
	m := &Manager{
		providers: []Provider{
			&mockComparer{
				name:    "first",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "first-response", nil
				},
			},
			&mockComparer{
				name:    "second",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "second-response", nil
				},
			},
		},
		providerMap: map[string]Provider{
			"first": &mockComparer{
				name:    "first",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "first-response", nil
				},
			},
			"second": &mockComparer{
				name:    "second",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "second-response", nil
				},
			},
		},
	}

	result, err := m.compareImages(context.Background(), "second", []byte("image1"), []byte("image2"), "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "second-response" {
		t.Errorf("expected 'second-response', got '%s'", result)
	}
}

func TestCompareImagesFallbackSkipsNonComparer(t *testing.T) {
	m := &Manager{
		providers: []Provider{
			&mockProvider{name: "no-compare", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "response", nil
			}},
			&mockComparer{
				name:    "comparer",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "compare-response", nil
				},
			},
		},
		providerMap: map[string]Provider{
			"no-compare": &mockProvider{name: "no-compare", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "response", nil
			}},
			"comparer": &mockComparer{
				name:    "comparer",
				model:   "model",
				analyze: func(ctx context.Context, image []byte, prompt string) (string, error) { return "", nil },
				compare: func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
					return "compare-response", nil
				},
			},
		},
		enableFallback: true,
	}

	result, err := m.CompareImages(context.Background(), "", []byte("image1"), []byte("image2"), "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "compare-response" {
		t.Errorf("expected 'compare-response', got '%s'", result)
	}
}

func TestCompareImagesFallbackAllNonComparer(t *testing.T) {
	m := &Manager{
		providers: []Provider{
			&mockProvider{name: "first", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "response", nil
			}},
			&mockProvider{name: "second", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "response", nil
			}},
		},
		providerMap: map[string]Provider{
			"first": &mockProvider{name: "first", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "response", nil
			}},
			"second": &mockProvider{name: "second", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "response", nil
			}},
		},
		enableFallback: true,
	}

	_, err := m.CompareImages(context.Background(), "", []byte("image1"), []byte("image2"), "prompt")
	if err == nil {
		t.Fatal("expected error when no providers support comparison")
	}
}

func TestTruncatePrompt(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"short prompt", "hello", "hello"},
		{"exactly 50 chars", "12345678901234567890123456789012345678901234567890", "12345678901234567890123456789012345678901234567890"},
		{"51 chars", "123456789012345678901234567890123456789012345678901", "12345678901234567890123456789012345678901234567..."},
		{"long prompt", "this is a very long prompt that should be truncated because it exceeds fifty characters", "this is a very long prompt that should be trunc..."},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncatePrompt(tt.input)
			if got != tt.want {
				t.Errorf("truncatePrompt(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestAnalyzeWithFallbackSpecificProviderFirst(t *testing.T) {
	m := &Manager{
		providers: []Provider{
			&mockProvider{name: "first", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "", fmt.Errorf("first provider error")
			}},
			&mockProvider{name: "second", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "second-response", nil
			}},
			&mockProvider{name: "third", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "third-response", nil
			}},
		},
		providerMap: map[string]Provider{
			"first": &mockProvider{name: "first", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "", fmt.Errorf("first provider error")
			}},
			"second": &mockProvider{name: "second", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "second-response", nil
			}},
			"third": &mockProvider{name: "third", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "third-response", nil
			}},
		},
		enableFallback: true,
	}

	result, err := m.AnalyzeWith(context.Background(), "second", []byte("image"), "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "second-response" {
		t.Errorf("expected 'second-response', got '%s'", result)
	}
}

func TestAnalyzeWithFallbackMiddleProviderFails(t *testing.T) {
	m := &Manager{
		providers: []Provider{
			&mockProvider{name: "first", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "", fmt.Errorf("first provider error")
			}},
			&mockProvider{name: "second", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "", fmt.Errorf("second provider error")
			}},
			&mockProvider{name: "third", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "third-response", nil
			}},
		},
		providerMap: map[string]Provider{
			"first": &mockProvider{name: "first", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "", fmt.Errorf("first provider error")
			}},
			"second": &mockProvider{name: "second", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "", fmt.Errorf("second provider error")
			}},
			"third": &mockProvider{name: "third", model: "model", analyze: func(ctx context.Context, image []byte, prompt string) (string, error) {
				return "third-response", nil
			}},
		},
		enableFallback: true,
	}

	result, err := m.AnalyzeWith(context.Background(), "", []byte("image"), "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "third-response" {
		t.Errorf("expected 'third-response', got '%s'", result)
	}
}

func TestNilManagerAnalyzeWith(t *testing.T) {
	var m *Manager
	_, err := m.AnalyzeWith(context.Background(), "test", []byte("image"), "prompt")
	if err == nil {
		t.Fatal("expected error from nil manager AnalyzeWith()")
	}
}

func TestEmptyManagerAnalyzeWith(t *testing.T) {
	m := &Manager{
		providerMap: map[string]Provider{},
	}
	_, err := m.AnalyzeWith(context.Background(), "", []byte("image"), "prompt")
	if err == nil {
		t.Fatal("expected error from empty manager AnalyzeWith()")
	}
}

func TestEmptyManagerCompareImages(t *testing.T) {
	m := &Manager{
		providerMap: map[string]Provider{},
	}
	_, err := m.CompareImages(context.Background(), "", []byte("image1"), []byte("image2"), "prompt")
	if err == nil {
		t.Fatal("expected error from empty manager CompareImages()")
	}
}
