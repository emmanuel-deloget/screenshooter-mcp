// Package tools provides the MCP tool implementations for screen capture operations.
//
// This package bridges the MCP protocol with the capture package, exposing the screen capture
// functionality as MCP tools. Each tool corresponds to a specific capture operation:
//   - ListMonitors: Enumerate available monitors
//   - ListWindows: Enumerate open windows
//   - CaptureScreen: Capture all or specific monitors
//   - CaptureWindow: Capture a specific window by title
//   - CaptureRegion: Capture an arbitrary screen region
//   - CompareImages: Compare two images and describe differences
//   - ExecutePipeline: Chain multiple capture and vision operations
//
// The tools accept context.Context for cancellation support, though currently
// the underlying capture operations do not support context cancellation.
// The context is included for future extensibility and API consistency.
//
// Image encoding is handled internally - all capture functions return PNG-encoded
// bytes ready for transmission to MCP clients.
package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"time"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/capture"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/vision"
)

// Tools provides MCP tool implementations for screen capture.
//
// Tools wraps a ScreenCapture implementation and exposes its functionality
// as MCP-compatible operations. The actual capture logic is delegated
// to the underlying ScreenCapture implementation, which handles the platform-
// specific details (X11 vs Wayland).
//
// The Tools struct is safe for concurrent use, as the capture implementation
// typically handles concurrency internally. However, concurrent capture
// operations may be serialized depending on the backend.
type Tools struct {
	capture capture.ScreenCapture
	vision  *vision.Manager
}

// NewTools creates a new Tools instance with the given capture implementation.
//
// The capture argument is the platform-specific ScreenCapture implementation
// that handles the actual screen capture operations. This allows the tools
// to work identically regardless of the underlying desktop environment.
//
// Example:
//
//	capt, err := x11.NewX11Capture()
//	if err != nil {
//	    return nil, err
//	}
//	tools := NewTools(capt)
func NewTools(c capture.ScreenCapture) *Tools {
	return &Tools{
		capture: c,
	}
}

// SetVisionManager sets the vision provider manager for image analysis tools.
//
// If mgr is nil, vision tools will return an error indicating that no
// providers are configured.
func (t *Tools) SetVisionManager(mgr *vision.Manager) {
	t.vision = mgr
}

// ListMonitors returns a list of all available monitors.
//
// This function delegates to the underlying ScreenCapture implementation.
// The monitor list includes each monitor's name, aliases, position, and
// dimensions in the virtual screen coordinate space.
//
// Returns a slice of Monitor structs, or an error if the operation fails.
// The slice may be empty if no monitors are detected.
func (t *Tools) ListMonitors(ctx context.Context) ([]capture.Monitor, error) {
	return t.capture.ListMonitors()
}

// ListWindows returns a list of all open windows.
//
// This function delegates to the underlying ScreenCapture implementation.
// The window list includes each window's ID, title, position, and dimensions.
//
// Returns a slice of Window structs, or an error if the operation fails.
// The slice is empty if no windows are detected or if the window manager is not available.
// Some implementations may not support window enumeration and will return an error.
func (t *Tools) ListWindows(ctx context.Context) ([]capture.Window, error) {
	return t.capture.ListWindows()
}

// CaptureScreen captures the full screen or a specific monitor.
//
// The monitor argument specifies which monitor to capture. If empty, all monitors
// are captured as a single combined image. If a monitor name or alias is
// provided, only that monitor is captured.
//
// Monitor matching is case-insensitive. The function tries exact name match
// first, then alias match.
//
// Returns PNG-encoded image data, or an error if the capture fails.
// The returned data is ready for transmission as MCP ImageContent.
func (t *Tools) CaptureScreen(ctx context.Context, monitor string) ([]byte, error) {
	img, err := t.capture.CaptureScreen(monitor)
	if err != nil {
		return nil, fmt.Errorf("capture screen failed: %w", err)
	}
	return encodeImage(img)
}

// CaptureWindow captures a window by its title.
//
// The title argument specifies the window to capture. Matching is case-
// insensitive and uses substring matching - if the window title contains
// the specified string, it is considered a match.
//
// Returns an error if multiple windows match (to prevent ambiguity) or if no
// window matches. The error message includes the partial title used.
//
// Returns PNG-encoded image data, or an error if the capture fails.
func (t *Tools) CaptureWindow(ctx context.Context, title string) ([]byte, error) {
	img, err := t.capture.CaptureWindow(title)
	if err != nil {
		return nil, fmt.Errorf("capture windows failed: %w", err)
	}
	return encodeImage(img)
}

// CaptureRegion captures an arbitrary rectangular region of the screen.
//
// The x and y arguments specify the top-left corner coordinates.
// The w and h arguments specify the width and height.
// Coordinates are relative to the virtual screen origin (0, 0).
//
// If the specified region extends beyond the screen bounds, it is clipped
// to the valid screen area.
//
// Returns PNG-encoded image data, or an error if the capture fails.
func (t *Tools) CaptureRegion(ctx context.Context, x, y, w, h int) ([]byte, error) {
	img, err := t.capture.CaptureRegion(x, y, w, h)
	if err != nil {
		return nil, fmt.Errorf("capture region failed: %w", err)
	}
	return encodeImage(img)
}

// ListVisionProviders returns metadata for all configured vision providers.
//
// Returns a list of provider names, models, and which one is the default.
// Returns an error if no vision providers are configured.
func (t *Tools) ListVisionProviders(ctx context.Context) ([]vision.ProviderInfo, error) {
	if t.vision == nil {
		return nil, fmt.Errorf("no vision providers configured")
	}
	return t.vision.Providers(), nil
}

// AnalyzeImage sends an image to a vision provider for analysis.
//
// The image argument should be PNG-encoded bytes. The prompt argument
// specifies what analysis to perform. If provider is empty, the default
// provider is used. If timeout is non-zero, it overrides the provider's
// configured timeout for this call (in seconds).
//
// Returns the text response from the AI model.
func (t *Tools) AnalyzeImage(ctx context.Context, image []byte, prompt string, provider string, timeout int) (string, error) {
	if t.vision == nil {
		return "", fmt.Errorf("no vision providers configured")
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}
	return t.vision.AnalyzeWith(ctx, provider, image, prompt)
}

// ExtractText performs OCR on an image and returns structured markdown text.
//
// The image argument should be PNG-encoded bytes. If provider is empty,
// the default provider is used. If timeout is non-zero, it overrides the
// provider's configured timeout for this call (in seconds).
//
// The prompt instructs the model to extract text and format it as markdown
// preserving structure (headings, bold, lists, tables, code blocks).
func (t *Tools) ExtractText(ctx context.Context, image []byte, provider string, timeout int) (string, error) {
	if t.vision == nil {
		return "", fmt.Errorf("no vision providers configured")
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}

	const extractTextPrompt = `Extract all text from this image.
Format the output as markdown, preserving the visual structure:
- Use headings for titles and section headers
- Use bold for emphasized text
- Use lists for bullet points
- Use tables for tabular data
- Use code blocks for code
Maintain the original language.`

	return t.vision.AnalyzeWith(ctx, provider, image, extractTextPrompt)
}

// FindRegion finds the bounding box coordinates of a described element in an image.
//
// The image argument should be PNG-encoded bytes. The description argument
// specifies what element to find (e.g., "the search button", "the logo").
// If provider is empty, the default provider is used. If timeout is non-zero,
// it overrides the provider's configured timeout for this call (in seconds).
//
// Returns the coordinates as x, y, width, height in pixels.
func (t *Tools) FindRegion(ctx context.Context, image []byte, description string, provider string, timeout int) (string, error) {
	if t.vision == nil {
		return "", fmt.Errorf("no vision providers configured")
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}

	prompt := fmt.Sprintf(`Find the bounding box coordinates of: %s

Return ONLY a JSON object with this exact structure, no other text:
{"x": <number>, "y": <number>, "width": <number>, "height": <number>}

Coordinates are in pixels relative to the top-left corner of the image.
Do not include any explanation or markdown formatting.`, description)

	return t.vision.AnalyzeWith(ctx, provider, image, prompt)
}

// CompareImages sends two images to a vision provider for comparison.
//
// Both image arguments should be PNG-encoded bytes. The prompt argument
// specifies what comparison to perform. If provider is empty, the default
// provider is used. If timeout is non-zero, it overrides the provider's
// configured timeout for this call (in seconds).
//
// Returns the text response from the AI model describing the comparison.
func (t *Tools) CompareImages(ctx context.Context, image1 []byte, image2 []byte, prompt string, provider string, timeout int) (string, error) {
	if t.vision == nil {
		return "", fmt.Errorf("no vision providers configured")
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}
	return t.vision.CompareImages(ctx, provider, image1, image2, prompt)
}

// encodeImage converts an image.Image to PNG-encoded bytes.
//
// This is an internal helper function used by all capture methods.
// It encodes the image to PNG format, which is the standard format for
// MCP ImageContent transmissions.
//
// Returns the PNG-encoded data, or an error if encoding fails.
func encodeImage(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}
	return buf.Bytes(), nil
}

// ToolResult represents the result of an MCP tool execution.
//
// This struct provides a standardized format for tool results,
// with separate fields for success status, data, and error information.
// It is used internally by the tools package for consistent result formatting.
//
// The JSON structure allows easy parsing by MCP clients:
//   - On success: Success is true, Data contains JSON-encoded result
//   - On error: Success is false, Error contains error message
type ToolResult struct {
	// Success indicates whether the tool executed successfully.
	// True for successful execution, false for errors.
	Success bool `json:"success"`

	// Data contains the JSON-encoded result data.
	// Only populated when Success is true.
	Data json.RawMessage `json:"data,omitempty"`

	// Error contains an error message string.
	// Only populated when Success is false.
	Error string `json:"error,omitempty"`
}

// SuccessResult creates a successful tool result.
//
// This helper function wraps arbitrary data into a ToolResult struct
// with Success set to true. The data is JSON-encoded and stored
// in the Data field.
//
// Returns an error if the data cannot be JSON-encoded.
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

// ErrorResult creates an error tool result.
//
// This helper function creates a ToolResult struct with Success set to
// false and the Error field set to the provided error message.
//
// The error parameter should be a human-readable error message
// suitable for display to the user.
func ErrorResult(err string) *ToolResult {
	return &ToolResult{
		Success: false,
		Error:   err,
	}
}
