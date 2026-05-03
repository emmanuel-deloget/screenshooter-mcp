# Unit Testing Approach for Tools

## Overview

The `internal/tools` package provides MCP tool implementations that wrap the `capture.ScreenCapture` interface and `vision.Manager`. To unit test these tools without requiring an actual display server (X11/Wayland), we need to create mock implementations of these dependencies.

## Key Interfaces to Mock

### 1. capture.ScreenCapture Interface

```go
type ScreenCapture interface {
    ListMonitors() ([]Monitor, error)
    ListWindows() ([]Window, error)
    CaptureScreen(monitor string) (image.Image, error)
    CaptureWindow(title string) (image.Image, error)
    CaptureRegion(x, y, w, h int) (image.Image, error)
    CaptureAllScreens() (image.Image, error)
}
```

### 2. vision.Manager (for vision tool tests)

The vision manager needs to be mockable for testing vision-dependent tools.

## Creating Mock Implementations

### Mock ScreenCapture

```go
// mockCapture implements capture.ScreenCapture for testing
type mockCapture struct {
    monitors []capture.Monitor
    windows  []capture.Window
    images   map[string]image.Image  // key: monitor name or "all"
    err      error                  // force error if set
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
    if img, ok := m.images[fmt.Sprintf("region:%d,%d,%d,%d", x, y, w, h)]; ok {
        return img, nil
    }
    // Return a blank image
    return image.NewRGBA(image.Rect(0, 0, w, h)), nil
}

func (m *mockCapture) CaptureAllScreens() (image.Image, error) {
    if m.err != nil {
        return nil, m.err
    }
    if img, ok := m.images["all"]; ok {
        return img, nil
    }
    return nil, fmt.Errorf("no image available")
}
```

## Test Structure

Create `internal/tools/tools_test.go`:

```go
package tools

import (
    "context"
    "image"
    "image/color"
    "testing"

    "github.com/emmanuel-deloget/screenshooter-mcp/internal/capture"
)

func TestListMonitors(t *testing.T) {
    tests := []struct {
        name     string
        mock     *mockCapture
        wantLen  int
        wantErr  bool
    }{
        {
            name: "success",
            mock: &mockCapture{
                monitors: []capture.Monitor{
                    {Name: "DP-1", Aliases: []string{"1"}, X: 0, Y: 0, Width: 1920, Height: 1080},
                },
            },
            wantLen: 1,
            wantErr: false,
        },
        {
            name: "error",
            mock: &mockCapture{
                err: fmt.Errorf("enumeration failed"),
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

func TestCaptureScreen(t *testing.T) {
    // Create a test image
    img := image.NewRGBA(image.Rect(0, 0, 100, 100))
    for i := range img.Pix {
        img.Pix[i] = 0xFF
    }

    tests := []struct {
        name    string
        monitor string
        mock    *mockCapture
        wantErr bool
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
                images: map[string]image.Image{"": img},
            },
            wantErr: false,
        },
        {
            name:    "monitor not found",
            monitor: "nonexistent",
            mock: &mockCapture{},
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
            // Verify the data is valid PNG
            if !tt.wantErr {
                _, err := png.Decode(bytes.NewReader(data))
                if err != nil {
                    t.Errorf("CaptureScreen() returned invalid PNG: %v", err)
                }
            }
        })
    }
}

func TestCaptureWindow(t *testing.T) {
    img := image.NewRGBA(image.Rect(0, 0, 100, 100))

    tests := []struct {
        name  string
        title string
        mock  *mockCapture
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
            mock: &mockCapture{},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tools := NewTools(tt.mock)
            data, err := tools.CaptureWindow(context.Background(), tt.title)

            if (err != nil) != tt.wantErr {
                t.Errorf("CaptureWindow() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func TestEncodeImage(t *testing.T) {
    tests := []struct {
        name string
        img  image.Image
        wantErr bool
    }{
        {
            name: "valid image",
            img:  image.NewRGBA(image.Rect(0, 0, 10, 10)),
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            data, err := encodeImage(tt.img)
            if (err != nil) != tt.wantErr {
                t.Errorf("encodeImage() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !tt.wantErr && len(data) == 0 {
                t.Error("encodeImage() returned empty data")
            }
        })
    }
}
```

## Vision Tools Testing

For vision tools (`AnalyzeImage`, `ExtractText`, `FindRegion`, `CompareImages`), create a mock vision.Manager:

```go
type mockVisionManager struct {
    providers []vision.ProviderInfo
    response  string
    err       error
}

func (m *mockVisionManager) Providers() []vision.ProviderInfo {
    return m.providers
}

func (m *mockVisionManager) AnalyzeWith(ctx context.Context, provider string, image []byte, prompt string) (string, error) {
    return m.response, m.err
}

func (m *mockVisionManager) CompareImages(ctx context.Context, provider string, image1, image2 []byte, prompt string) (string, error) {
    return m.response, m.err
}
```

## Test Coverage Goals

Aim to test:
1. **Happy path**: Normal operation with valid inputs
2. **Error handling**: Capture backend errors, missing windows/monitors
3. **Edge cases**: Empty strings, invalid parameters
4. **Image encoding**: Verify returned data is valid PNG
5. **Vision tools**: With and without configured providers, timeout handling

## Running Tests

```bash
# Run tools tests
go test ./internal/tools/...

# Run with coverage
go test -cover ./internal/tools/...

# Run all tests
go test ./...
```
