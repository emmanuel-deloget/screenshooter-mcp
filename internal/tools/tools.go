package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/capture"
)

type Tools struct {
	capture capture.ScreenCapture
}

func NewTools(c capture.ScreenCapture) *Tools {
	return &Tools{
		capture: c,
	}
}

func (t *Tools) ListMonitors(ctx context.Context) ([]capture.Monitor, error) {
	return t.capture.ListMonitors()
}

func (t *Tools) ListWindows(ctx context.Context) ([]capture.Window, error) {
	return t.capture.ListWindows()
}

func (t *Tools) CaptureScreen(ctx context.Context, monitor string) ([]byte, error) {
	img, err := t.capture.CaptureScreen(monitor)
	if err != nil {
		return nil, fmt.Errorf("capture screen failed: %w", err)
	}
	return encodeImage(img)
}

func (t *Tools) CaptureWindow(ctx context.Context, title string) ([]byte, error) {
	img, err := t.capture.CaptureWindow(title)
	if err != nil {
		return nil, fmt.Errorf("capture window failed: %w", err)
	}
	return encodeImage(img)
}

func (t *Tools) CaptureRegion(ctx context.Context, x, y, w, h int) ([]byte, error) {
	img, err := t.capture.CaptureRegion(x, y, w, h)
	if err != nil {
		return nil, fmt.Errorf("capture region failed: %w", err)
	}
	return encodeImage(img)
}

func encodeImage(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}
	return buf.Bytes(), nil
}

type ToolResult struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

func SuccessResult(data interface{}) (*ToolResult, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &ToolResult{
		Success: true,
		Data:    jsonData,
	}, nil
}

func ErrorResult(err string) *ToolResult {
	return &ToolResult{
		Success: false,
		Error:   err,
	}
}
