package vision

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/config"
)

func TestHuggingFaceProviderAnalyze(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models/test-model" {
			t.Errorf("expected path /models/test-model, got %s", r.URL.Path)
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Errorf("expected auth 'Bearer test-key', got '%s'", auth)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"generated_text": "Test response from HF"}`))
	}))
	defer server.Close()

	cfg := config.VisionProviderConfig{
		Name:    "test-hf",
		Type:    "huggingface",
		Model:   "test-model",
		APIKey:  "test-key",
		BaseURL: server.URL + "/models/test-model",
		Timeout: 5,
	}

	p, err := newHuggingFaceProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	result, err := p.Analyze(context.Background(), []byte("fake-image-data"), "Describe this image")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Test response from HF" {
		t.Errorf("expected 'Test response from HF', got '%s'", result)
	}
}

func TestHuggingFaceProviderEmptyModel(t *testing.T) {
	cfg := config.VisionProviderConfig{
		Name:   "test",
		Type:   "huggingface",
		APIKey: "test-key",
	}
	_, err := newHuggingFaceProvider(cfg)
	if err == nil {
		t.Fatal("expected error for empty model")
	}
}

func TestHuggingFaceProviderEmptyAPIKey(t *testing.T) {
	cfg := config.VisionProviderConfig{
		Name:  "test",
		Type:  "huggingface",
		Model: "test-model",
	}
	_, err := newHuggingFaceProvider(cfg)
	if err == nil {
		t.Fatal("expected error for empty api_key")
	}
}

func TestHuggingFaceProviderCustomURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"generated_text": "ok"}`))
	}))
	defer server.Close()

	cfg := config.VisionProviderConfig{
		Name:    "test-hf",
		Type:    "huggingface",
		Model:   "org/model",
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 5,
	}

	p, err := newHuggingFaceProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	_, err = p.Analyze(context.Background(), []byte("image"), "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHuggingFaceProviderArrayResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"generated_text": "Array response"}]`))
	}))
	defer server.Close()

	cfg := config.VisionProviderConfig{
		Name:    "test-hf",
		Type:    "huggingface",
		Model:   "test-model",
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 5,
	}

	p, err := newHuggingFaceProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	result, err := p.Analyze(context.Background(), []byte("image"), "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Array response" {
		t.Errorf("expected 'Array response', got '%s'", result)
	}
}

func TestHuggingFaceProviderStringResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`"Direct string response"`))
	}))
	defer server.Close()

	cfg := config.VisionProviderConfig{
		Name:    "test-hf",
		Type:    "huggingface",
		Model:   "test-model",
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 5,
	}

	p, err := newHuggingFaceProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	result, err := p.Analyze(context.Background(), []byte("image"), "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Direct string response" {
		t.Errorf("expected 'Direct string response', got '%s'", result)
	}
}

func TestHuggingFaceProviderErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Model loading"}`))
	}))
	defer server.Close()

	cfg := config.VisionProviderConfig{
		Name:    "test-hf",
		Type:    "huggingface",
		Model:   "test-model",
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 5,
	}

	p, err := newHuggingFaceProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	_, err = p.Analyze(context.Background(), []byte("image"), "prompt")
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}
