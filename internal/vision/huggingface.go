package vision

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/config"
)

// huggingFaceProvider implements Provider for HuggingFace Inference API.
//
// This provider uses the HuggingFace Inference API to send images and
// receive text responses via direct HTTP calls. It supports vision-language
// models hosted on HuggingFace.
type huggingFaceProvider struct {
	name    string
	model   string
	apiKey  string
	baseURL string
	timeout int
}

func newHuggingFaceProvider(cfg config.VisionProviderConfig) (Provider, error) {
	if cfg.Model == "" {
		return nil, fmt.Errorf("model is required for huggingface provider")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is required for huggingface provider")
	}
	return &huggingFaceProvider{
		name:    cfg.Name,
		model:   cfg.Model,
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		timeout: cfg.DefaultTimeout(),
	}, nil
}

func (p *huggingFaceProvider) Name() string {
	return p.name
}

func (p *huggingFaceProvider) ModelName() string {
	return p.model
}

func (p *huggingFaceProvider) Analyze(ctx context.Context, image []byte, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.timeout)*time.Second)
	defer cancel()

	base64Image := base64.StdEncoding.EncodeToString(image)

	url := p.baseURL
	if url == "" {
		url = fmt.Sprintf("https://api-inference.huggingface.co/models/%s", p.model)
	}

	body := map[string]interface{}{
		"inputs": map[string]interface{}{
			"image": "data:image/png;base64," + base64Image,
			"prompt": prompt,
		},
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	client := &http.Client{Timeout: time.Duration(p.timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return parseHuggingFaceResponse(respBody)
}

// parseHuggingFaceResponse extracts the text from various HF API response formats.
func parseHuggingFaceResponse(body []byte) (string, error) {
	var raw interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	switch v := raw.(type) {
	case string:
		return v, nil
	case []interface{}:
		if len(v) == 0 {
			return "", fmt.Errorf("empty response array")
		}
		if obj, ok := v[0].(map[string]interface{}); ok {
			if generatedText, ok := obj["generated_text"].(string); ok {
				return generatedText, nil
			}
		}
		if str, ok := v[0].(string); ok {
			return str, nil
		}
	case map[string]interface{}:
		if generatedText, ok := v["generated_text"].(string); ok {
			return generatedText, nil
		}
		if output, ok := v["output"].(string); ok {
			return output, nil
		}
		if choices, ok := v["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if message, ok := choice["message"].(map[string]interface{}); ok {
					if content, ok := message["content"].(string); ok {
						return content, nil
					}
				}
				if text, ok := choice["text"].(string); ok {
					return text, nil
				}
			}
		}
	}

	return "", fmt.Errorf("unable to parse response: %s", string(body))
}
