package vision

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/config"
)

func TestAnthropicProviderAnalyze(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Errorf("expected path /v1/messages, got %s", r.URL.Path)
		}

		var req struct {
			Model     string `json:"model"`
			MaxTokens int64  `json:"max_tokens"`
			Messages  []struct {
				Role    string `json:"role"`
				Content []struct {
					Type   string `json:"type"`
					Text   string `json:"text,omitempty"`
					Source struct {
						Type      string `json:"type"`
						Data      string `json:"data"`
						MediaType string `json:"media_type"`
					} `json:"source,omitempty"`
				} `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Model != "claude-test" {
			t.Errorf("expected model 'claude-test', got '%s'", req.Model)
		}

		resp := map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "Test response from Claude",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := config.VisionProviderConfig{
		Name:    "test-claude",
		Type:    "anthropic",
		Model:   "claude-test",
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 5,
	}

	p, err := newAnthropicProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	result, err := p.Analyze(context.Background(), []byte("fake-image-data"), "Describe this image")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Test response from Claude" {
		t.Errorf("expected 'Test response from Claude', got '%s'", result)
	}
}

func TestAnthropicProviderEmptyModel(t *testing.T) {
	cfg := config.VisionProviderConfig{
		Name:   "test",
		Type:   "anthropic",
		APIKey: "test-key",
	}
	_, err := newAnthropicProvider(cfg)
	if err == nil {
		t.Fatal("expected error for empty model")
	}
}

func TestAnthropicProviderEmptyAPIKey(t *testing.T) {
	cfg := config.VisionProviderConfig{
		Name:  "test",
		Type:  "anthropic",
		Model: "claude-test",
	}
	_, err := newAnthropicProvider(cfg)
	if err == nil {
		t.Fatal("expected error for empty api_key")
	}
}

func TestAnthropicProviderNoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"content": []interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := config.VisionProviderConfig{
		Name:    "test-claude",
		Type:    "anthropic",
		Model:   "claude-test",
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 5,
	}

	p, err := newAnthropicProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	_, err = p.Analyze(context.Background(), []byte("image"), "prompt")
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}
