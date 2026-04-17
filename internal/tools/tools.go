package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/capture"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/vision"
)

type Tools struct {
	capture capture.ScreenCapture
	vision  *vision.Manager
}

func NewTools(c capture.ScreenCapture, v *vision.Manager) *Tools {
	return &Tools{
		capture: c,
		vision:  v,
	}
}

func (t *Tools) ListMonitors(ctx context.Context) ([]capture.Monitor, error) {
	return t.capture.ListMonitors()
}

func (t *Tools) CaptureScreen(ctx context.Context, monitor string) (string, error) {
	img, err := t.capture.CaptureScreen(monitor)
	if err != nil {
		return "", fmt.Errorf("capture screen failed: %w", err)
	}
	return encodeImageToBase64(img)
}

func (t *Tools) CaptureWindow(ctx context.Context, windowID int64) (string, error) {
	img, err := t.capture.CaptureWindow(capture.WindowID(windowID))
	if err != nil {
		return "", fmt.Errorf("capture window failed: %w", err)
	}
	return encodeImageToBase64(img)
}

func (t *Tools) CaptureRegion(ctx context.Context, x, y, w, h int) (string, error) {
	img, err := t.capture.CaptureRegion(x, y, w, h)
	if err != nil {
		return "", fmt.Errorf("capture region failed: %w", err)
	}
	return encodeImageToBase64(img)
}

func (t *Tools) FindElement(ctx context.Context, imageData string, description string) (*capture.Element, error) {
	img, err := decodeImageFromBase64(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return t.vision.FindElement(ctx, img, description)
}

func (t *Tools) CaptureElement(ctx context.Context, imageData string, description string) (string, error) {
	element, err := t.FindElement(ctx, imageData, description)
	if err != nil {
		return "", err
	}

	img, err := decodeImageFromBase64(imageData)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	cropped := capture.CropToBoundingBox(img, &element.BoundingBox)
	return encodeImageToBase64(cropped)
}

func encodeImageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func decodeImageFromBase64(data string) (image.Image, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	img, err := png.Decode(bytes.NewReader(decoded))
	if err != nil {
		return nil, fmt.Errorf("failed to decode png: %w", err)
	}

	return img, nil
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
