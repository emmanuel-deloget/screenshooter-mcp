package vision

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/capture"
)

type VisionQuality string

const (
	QualityLow    VisionQuality = "low"
	QualityMiddle VisionQuality = "middle"
	QualityHigh   VisionQuality = "high"
)

var qualityToModel = map[VisionQuality]string{
	QualityLow:    "qwen3-vl:2b",
	QualityMiddle: "qwen3-vl:4b",
	QualityHigh:   "qwen3-vl:8b",
}

type Manager struct {
	modelPath string
	model     string
	quality   VisionQuality
	port      int
	pid       int
	url       string
	client    *http.Client
	cmd       *exec.Cmd
}

type ManagerOption func(*Manager)

func WithModelPath(path string) ManagerOption {
	return func(m *Manager) {
		m.modelPath = path
	}
}

func WithModel(model string) ManagerOption {
	return func(m *Manager) {
		m.model = model
	}
}

func WithQuality(quality VisionQuality) ManagerOption {
	return func(m *Manager) {
		m.quality = quality
	}
}

func NewManager(opts ...ManagerOption) (*Manager, error) {
	m := &Manager{
		quality: QualityMiddle,
	}

	for _, opt := range opts {
		opt(m)
	}

	if m.model == "" {
		m.model = qualityToModel[m.quality]
	}

	if err := m.startOllama(); err != nil {
		return nil, fmt.Errorf("failed to start Ollama: %w", err)
	}

	return m, nil
}

func (m *Manager) startOllama() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}
	m.port = listener.Addr().(*net.TCPAddr).Port
	m.url = fmt.Sprintf("http://127.0.0.1:%d", m.port)
	listener.Close()

	ollamaPath, err := findOllamaBinary()
	if err != nil {
		return fmt.Errorf("failed to find ollama binary: %w", err)
	}

	args := []string{"serve", "--host", "127.0.0.1", "--port", fmt.Sprintf("%d", m.port)}
	if m.modelPath != "" {
		args = append(args, "--model-path", m.modelPath)
	}

	m.cmd = exec.Command(ollamaPath, args...)
	m.cmd.Stdout = nil
	m.cmd.Stderr = nil

	ollamaDir := filepath.Dir(ollamaPath)
	m.cmd.Env = append(os.Environ(), "LD_LIBRARY_PATH="+filepath.Join(ollamaDir, "lib"))

	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ollama: %w", err)
	}

	m.pid = m.cmd.Process.Pid

	m.client = &http.Client{
		Timeout: 120 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			m.Stop()
			return fmt.Errorf("timeout waiting for ollama to start")
		default:
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.url, nil)
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			resp, err := m.client.Do(req)
			if err == nil {
				resp.Body.Close()
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func findOllamaBinary() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	localBin := filepath.Join(filepath.Dir(exe), "bin", "ollama")
	if _, err := os.Stat(localBin); err == nil {
		return localBin, nil
	}

	wd, err := os.Getwd()
	if err == nil {
		wdBin := filepath.Join(wd, "bin", "ollama")
		if _, err := os.Stat(wdBin); err == nil {
			return wdBin, nil
		}
	}

	return "", fmt.Errorf("ollama binary not found in ./bin/ollama")
}

func (m *Manager) Stop() {
	if m.pid > 0 {
		syscall.Kill(m.pid, syscall.SIGTERM)
		m.pid = 0
	}
}

func (m *Manager) URL() string {
	return m.url
}

func (m *Manager) Model() string {
	return m.model
}

func (m *Manager) PID() int {
	return m.pid
}

type FindElementResponse struct {
	BBox        *[4]int  `json:"bbox"`
	Description string   `json:"description,omitempty"`
	Error       string   `json:"error,omitempty"`
	Confidence  *float64 `json:"confidence,omitempty"`
}

func (m *Manager) FindElement(ctx context.Context, img image.Image, description string) (*capture.Element, error) {
	imgData, format, err := encodeImage(img, m.quality)
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	prompt := fmt.Sprintf(`Analyze this %s image and find the region described.
%s
Respond ONLY with a JSON object containing the bounding box:
{"bbox": [x1, y1, x2, y2], "description": "what you found", "confidence": 0.0-1.0}

Coordinates are in pixels, with (0,0) at top-left.
Return the exact bounding box of the described element.
If not found, respond: {"bbox": null, "error": "element not found"}`, format, getQualityDescription(m.quality))

	resp, err := m.chat(ctx, prompt, imgData, format)
	if err != nil {
		return nil, fmt.Errorf("vision request failed: %w", err)
	}

	var fer FindElementResponse
	if err := json.Unmarshal([]byte(resp), &fer); err != nil {
		return nil, fmt.Errorf("failed to parse vision response: %w", err)
	}

	if fer.Error != "" {
		return nil, fmt.Errorf("vision error: %s", fer.Error)
	}

	if fer.BBox == nil {
		return nil, fmt.Errorf("element not found")
	}

	confidence := 0.0
	if fer.Confidence != nil {
		confidence = *fer.Confidence
	}

	return &capture.Element{
		BoundingBox: capture.BoundingBox{
			X1: (*fer.BBox)[0],
			Y1: (*fer.BBox)[1],
			X2: (*fer.BBox)[2],
			Y2: (*fer.BBox)[3],
		},
		Description: fer.Description,
		Confidence:  confidence,
	}, nil
}

func (m *Manager) chat(ctx context.Context, prompt string, imageData []byte, format string) (string, error) {
	body := map[string]interface{}{
		"model":  m.model,
		"prompt": prompt,
		"stream": false,
		"images": []map[string]interface{}{{"data": string(imageData), "format": format}},
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.url+"/api/chat", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Message.Content, nil
}

func getQualityDescription(q VisionQuality) string {
	switch q {
	case QualityLow:
		return "Use minimal detail for fast processing."
	case QualityHigh:
		return "Use maximum detail for accurate processing."
	default:
		return "Use balanced detail for processing."
	}
}
