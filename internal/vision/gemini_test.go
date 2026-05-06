package vision

import (
	"testing"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/config"
)

func TestNewGeminiProviderMissingModel(t *testing.T) {
	cfg := config.VisionProviderConfig{
		Name:   "gemini",
		Type:   "gemini",
		APIKey: "test-key",
	}
	_, err := newGeminiProvider(cfg)
	if err == nil {
		t.Fatal("expected error for missing model")
	}
}

func TestNewGeminiProviderMissingAPIKey(t *testing.T) {
	cfg := config.VisionProviderConfig{
		Name:  "gemini",
		Type:  "gemini",
		Model: "gemini-2.5-flash",
	}
	_, err := newGeminiProvider(cfg)
	if err == nil {
		t.Fatal("expected error for missing api_key in Gemini API mode")
	}
}

func TestNewGeminiProviderGeminiAPIMode(t *testing.T) {
	cfg := config.VisionProviderConfig{
		Name:   "gemini",
		Type:   "gemini",
		Model:  "gemini-2.5-flash",
		APIKey: "test-key",
	}
	p, err := newGeminiProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gp, ok := p.(*geminiProvider)
	if !ok {
		t.Fatal("expected *geminiProvider")
	}
	if gp.apiKey != "test-key" {
		t.Errorf("expected api_key 'test-key', got '%s'", gp.apiKey)
	}
	if gp.project != "" {
		t.Errorf("expected empty project, got '%s'", gp.project)
	}
	if gp.location != "" {
		t.Errorf("expected empty location, got '%s'", gp.location)
	}
}

func TestNewGeminiProviderVertexAIMode(t *testing.T) {
	cfg := config.VisionProviderConfig{
		Name:     "gemini-vertex",
		Type:     "gemini",
		Model:    "gemini-2.5-flash",
		Project:  "my-project",
		Location: "us-central1",
	}
	p, err := newGeminiProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gp, ok := p.(*geminiProvider)
	if !ok {
		t.Fatal("expected *geminiProvider")
	}
	if gp.project != "my-project" {
		t.Errorf("expected project 'my-project', got '%s'", gp.project)
	}
	if gp.location != "us-central1" {
		t.Errorf("expected location 'us-central1', got '%s'", gp.location)
	}
}

func TestNewGeminiProviderVertexAIWithAPIKey(t *testing.T) {
	cfg := config.VisionProviderConfig{
		Name:     "gemini-vertex",
		Type:     "gemini",
		Model:    "gemini-2.5-flash",
		APIKey:   "test-key",
		Project:  "my-project",
		Location: "us-central1",
	}
	p, err := newGeminiProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gp, ok := p.(*geminiProvider)
	if !ok {
		t.Fatal("expected *geminiProvider")
	}
	if gp.apiKey != "test-key" {
		t.Errorf("expected api_key 'test-key', got '%s'", gp.apiKey)
	}
	if gp.project != "my-project" {
		t.Errorf("expected project 'my-project', got '%s'", gp.project)
	}
}

func TestGeminiProviderName(t *testing.T) {
	p := &geminiProvider{name: "test-gemini", model: "gemini-2.5-flash"}
	if p.Name() != "test-gemini" {
		t.Errorf("expected name 'test-gemini', got '%s'", p.Name())
	}
}

func TestGeminiProviderModelName(t *testing.T) {
	p := &geminiProvider{name: "test-gemini", model: "gemini-2.5-flash"}
	if p.ModelName() != "gemini-2.5-flash" {
		t.Errorf("expected model 'gemini-2.5-flash', got '%s'", p.ModelName())
	}
}
