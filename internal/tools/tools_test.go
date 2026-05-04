package tools

import (
	"context"
	"fmt"
	"image"
	"testing"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/capture"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/vision"
)

// mockCapture implements capture.ScreenCapture for testing
type mockCapture struct {
	monitors []capture.Monitor
	windows  []capture.Window
	images   map[string]image.Image
	err      error
}

func (m *mockCapture) ListMonitors() ([]capture.Monitor, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.monitors, nil
}

func (m *mockCapture) ListWindows() ([]capture.Window, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.windows, nil
}

func (m *mockCapture) CaptureScreen(monitor string) (image.Image, error) {
	if m.err != nil {
		return nil, m.err
	}
	if img, ok := m.images[monitor]; ok {
		return img, nil
	}
	if monitor == "" {
		if img, ok := m.images["all"]; ok {
			return img, nil
		}
	}
	return nil, fmt.Errorf("monitor not found: %s", monitor)
}

func (m *mockCapture) CaptureWindow(title string) (image.Image, error) {
	if m.err != nil {
		return nil, m.err
	}
	if img, ok := m.images["window:"+title]; ok {
		return img, nil
	}
	return nil, fmt.Errorf("window not found: %s", title)
}

func (m *mockCapture) CaptureRegion(x, y, w, h int) (image.Image, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := "region"
	if img, ok := m.images[key]; ok {
		return img, nil
	}
	return image.NewRGBA(image.Rect(0, 0, w, h)), nil
}

func (m *mockCapture) CaptureAllScreens() (image.Image, error) {
	if m.err != nil {
		return nil, m.err
	}
	if img, ok := m.images["all"]; ok {
		return img, nil
	}
	return nil, fmt.Errorf("capture failed")
}

func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for i := range img.Pix {
		img.Pix[i] = 0xFF
	}
	return img
}

// mockVisionManager wraps *vision.Manager to allow testing
type mockVisionManager struct {
	mgr *vision.Manager
}

func (m *mockVisionManager) Providers() []vision.ProviderInfo {
	if m.mgr == nil {
		return nil
	}
	return m.mgr.Providers()
}

func (m *mockVisionManager) AnalyzeWith(ctx context.Context, provider string, image []byte, prompt string) (string, error) {
	if m.mgr == nil {
		return "", fmt.Errorf("no vision providers configured")
	}
	return m.mgr.AnalyzeWith(ctx, provider, image, prompt)
}

func (m *mockVisionManager) CompareImages(ctx context.Context, provider string, image1, image2 []byte, prompt string) (string, error) {
	if m.mgr == nil {
		return "", fmt.Errorf("no vision providers configured")
	}
	return m.mgr.CompareImages(ctx, provider, image1, image2, prompt)
}

func TestSetVisionManager(t *testing.T) {
	tools := NewTools(&mockCapture{})
	if tools.vision != nil {
		t.Error("vision manager should be nil initially")
	}

	mgr := &vision.Manager{}
	tools.SetVisionManager(mgr)

	if tools.vision != mgr {
		t.Error("SetVisionManager() did not set the vision manager")
	}
}

func TestListVisionProvidersWithManager(t *testing.T) {
	tools := NewTools(&mockCapture{})
	tools.SetVisionManager(&vision.Manager{})

	providers, err := tools.ListVisionProviders(context.Background())
	if err != nil {
		t.Errorf("ListVisionProviders() with manager error = %v", err)
	}
	// Empty manager returns empty slice, not nil
	if len(providers) != 0 {
		t.Errorf("ListVisionProviders() should return empty slice for empty manager, got %v", providers)
	}
}

func TestAnalyzeImage(t *testing.T) {
	tools := NewTools(&mockCapture{})
	tools.SetVisionManager(&vision.Manager{})

	_, err := tools.AnalyzeImage(context.Background(), []byte("imagedata"), "prompt", "", 0)
	if err == nil {
		t.Error("AnalyzeImage() should fail with empty manager")
	}
}

func TestExtractText(t *testing.T) {
	tools := NewTools(&mockCapture{})
	tools.SetVisionManager(&vision.Manager{})

	_, err := tools.ExtractText(context.Background(), []byte("imagedata"), "", 0)
	if err == nil {
		t.Error("ExtractText() should fail with empty manager")
	}
}

func TestFindRegion(t *testing.T) {
	tools := NewTools(&mockCapture{})
	tools.SetVisionManager(&vision.Manager{})

	_, err := tools.FindRegion(context.Background(), []byte("imagedata"), "description", "", 0)
	if err == nil {
		t.Error("FindRegion() should fail with empty manager")
	}
}

func TestCompareImages(t *testing.T) {
	tools := NewTools(&mockCapture{})
	tools.SetVisionManager(&vision.Manager{})

	_, err := tools.CompareImages(context.Background(), []byte("image1"), []byte("image2"), "", "", 0)
	if err == nil {
		t.Error("CompareImages() should fail with empty manager")
	}
}

func TestListMonitors(t *testing.T) {
	tests := []struct {
		name     string
		mock     *mockCapture
		wantLen  int
		wantErr  bool
	}{
		{
			name: "success with monitors",
			mock: &mockCapture{
				monitors: []capture.Monitor{
					{Name: "DP-1", Aliases: []string{"1"}, X: 0, Y: 0, Width: 1920, Height: 1080},
					{Name: "DP-2", Aliases: []string{"2"}, X: 1920, Y: 0, Width: 1920, Height: 1080},
				},
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "success with no monitors",
			mock: &mockCapture{
				monitors: []capture.Monitor{},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "error from capture backend",
			mock: &mockCapture{
				err: fmt.Errorf("capture failed"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := NewTools(tt.mock)
			monitors, err := tools.ListMonitors(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListMonitors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(monitors) != tt.wantLen {
				t.Errorf("ListMonitors() got %v monitors, want %v", len(monitors), tt.wantLen)
			}
		})
	}
}

func TestListWindows(t *testing.T) {
	tests := []struct {
		name     string
		mock     *mockCapture
		wantLen  int
		wantErr  bool
	}{
		{
			name: "success with windows",
			mock: &mockCapture{
				windows: []capture.Window{
					{ID: 1, Name: "Terminal", X: 0, Y: 0, Width: 800, Height: 600, Active: true},
					{ID: 2, Name: "Firefox", X: 100, Y: 100, Width: 1200, Height: 800},
				},
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "success with no windows",
			mock: &mockCapture{
				windows: []capture.Window{},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "error from capture backend",
			mock: &mockCapture{
				err: fmt.Errorf("capture failed"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := NewTools(tt.mock)
			windows, err := tools.ListWindows(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListWindows() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(windows) != tt.wantLen {
				t.Errorf("ListWindows() got %v windows, want %v", len(windows), tt.wantLen)
			}
		})
	}
}

func TestCaptureScreen(t *testing.T) {
	img := createTestImage(100, 100)

	tests := []struct {
		name     string
		monitor  string
		mock     *mockCapture
		wantErr  bool
	}{
		{
			name:    "capture specific monitor",
			monitor: "DP-1",
			mock: &mockCapture{
				images: map[string]image.Image{"DP-1": img},
			},
			wantErr: false,
		},
		{
			name:    "capture all screens",
			monitor: "",
			mock: &mockCapture{
				images: map[string]image.Image{"all": img},
			},
			wantErr: false,
		},
		{
			name:    "monitor not found",
			monitor: "nonexistent",
			mock:    &mockCapture{},
			wantErr: true,
		},
		{
			name:    "error from capture backend",
			monitor: "DP-1",
			mock: &mockCapture{
				err: fmt.Errorf("capture failed"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := NewTools(tt.mock)
			data, err := tools.CaptureScreen(context.Background(), tt.monitor)

			if (err != nil) != tt.wantErr {
				t.Errorf("CaptureScreen() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && data == nil {
				t.Error("CaptureScreen() returned nil data without error")
			}
		})
	}
}

func TestCaptureWindow(t *testing.T) {
	img := createTestImage(100, 100)

	tests := []struct {
		name    string
		title   string
		mock    *mockCapture
		wantErr bool
	}{
		{
			name:  "capture existing window",
			title: "Terminal",
			mock: &mockCapture{
				images: map[string]image.Image{"window:Terminal": img},
			},
			wantErr: false,
		},
		{
			name:  "window not found",
			title: "Nonexistent",
			mock:  &mockCapture{},
			wantErr: true,
		},
		{
			name:  "error from capture backend",
			title: "Terminal",
			mock: &mockCapture{
				err: fmt.Errorf("capture failed"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := NewTools(tt.mock)
			_, err := tools.CaptureWindow(context.Background(), tt.title)

			if (err != nil) != tt.wantErr {
				t.Errorf("CaptureWindow() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCaptureRegion(t *testing.T) {
	img := createTestImage(100, 100)

	tests := []struct {
		name    string
		x, y    int
		w, h    int
		mock    *mockCapture
		wantErr bool
	}{
		{
			name: "capture valid region",
			x:    0,
			y:    0,
			w:    100,
			h:    100,
			mock: &mockCapture{
				images: map[string]image.Image{"region": img},
			},
			wantErr: false,
		},
		{
			name: "error from capture backend",
			x:    0,
			y:    0,
			w:    100,
			h:    100,
			mock: &mockCapture{
				err: fmt.Errorf("capture failed"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := NewTools(tt.mock)
			data, err := tools.CaptureRegion(context.Background(), tt.x, tt.y, tt.w, tt.h)

			if (err != nil) != tt.wantErr {
				t.Errorf("CaptureRegion() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && data == nil {
				t.Error("CaptureRegion() returned nil data without error")
			}
		})
	}
}

func TestEncodeImage(t *testing.T) {
	tests := []struct {
		name    string
		img     image.Image
		wantErr bool
	}{
		{
			name:    "valid image",
			img:     image.NewRGBA(image.Rect(0, 0, 10, 10)),
			wantErr: false,
		},
		{
			name:    "nil image",
			img:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := encodeImage(tt.img)

			if tt.img == nil {
				if err == nil {
					t.Error("encodeImage(nil) should return error")
				}
				return
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("encodeImage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(data) == 0 {
				t.Error("encodeImage() returned empty data")
			}
		})
	}
}

func TestGetSkillInfo(t *testing.T) {
	tools := NewTools(&mockCapture{})
	info := tools.GetSkillInfo()

	if info == "" {
		t.Error("GetSkillInfo() returned empty string")
	}
}

func TestSuccessResult(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
	}{
		{
			name:    "string data",
			data:    "test result",
			wantErr: false,
		},
		{
			name:    "struct data",
			data:    map[string]string{"key": "value"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SuccessResult(tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("SuccessResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !result.Success {
					t.Error("SuccessResult() Success should be true")
				}
				if len(result.Data) == 0 {
					t.Error("SuccessResult() Data should not be empty")
				}
			}
		})
	}
}

func TestErrorResult(t *testing.T) {
	errMsg := "test error"
	result := ErrorResult(errMsg)

	if result.Success {
		t.Error("ErrorResult() Success should be false")
	}
	if result.Error != errMsg {
		t.Errorf("ErrorResult() Error = %v, want %v", result.Error, errMsg)
	}
}
