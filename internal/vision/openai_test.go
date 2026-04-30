package vision

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/config"
)

func TestOpenAICompatibleProviderAnalyze(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected path /chat/completions, got %s", r.URL.Path)
		}

		var req struct {
			Model    string `json:"model"`
			Messages []struct {
				Role         string `json:"role"`
				MultiContent []struct {
					Type     string `json:"type"`
					Text     string `json:"text,omitempty"`
					ImageURL struct {
						URL    string `json:"url"`
						Detail string `json:"detail"`
					} `json:"image_url,omitempty"`
				} `json:"multi_content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Model != "test-model" {
			t.Errorf("expected model 'test-model', got '%s'", req.Model)
		}

		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": "Test response from AI",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := config.VisionProviderConfig{
		Name:    "test",
		Type:    "openai-compatible",
		Model:   "test-model",
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 5,
	}

	p, err := newOpenAICompatibleProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	result, err := p.Analyze(context.Background(), []byte("fake-image-data"), "Describe this image")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Test response from AI" {
		t.Errorf("expected 'Test response from AI', got '%s'", result)
	}
}

func TestOpenAICompatibleProviderEmptyModel(t *testing.T) {
	cfg := config.VisionProviderConfig{
		Name: "test",
		Type: "openai-compatible",
	}
	_, err := newOpenAICompatibleProvider(cfg)
	if err == nil {
		t.Fatal("expected error for empty model")
	}
}

func TestOpenAICompatibleProviderNoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"choices": []interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := config.VisionProviderConfig{
		Name:    "test",
		Type:    "openai-compatible",
		Model:   "test-model",
		BaseURL: server.URL,
		Timeout: 5,
	}

	p, err := newOpenAICompatibleProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	_, err = p.Analyze(context.Background(), []byte("image"), "prompt")
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
}

func TestOpenAICompatibleProviderTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		}
	}))
	defer server.Close()

	cfg := config.VisionProviderConfig{
		Name:    "test",
		Type:    "openai-compatible",
		Model:   "test-model",
		BaseURL: server.URL,
		Timeout: 1,
	}

	p, err := newOpenAICompatibleProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100)
	defer cancel()

	_, err = p.Analyze(ctx, []byte("image"), "prompt")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
