package wayland

import (
	"fmt"
	"image"

	capture "github.com/emmanuel-deloget/screenshooter-mcp/internal/capture"
)

type WaylandCapture struct{}

func NewWaylandCapture() *WaylandCapture {
	return &WaylandCapture{}
}

func (c *WaylandCapture) ListMonitors() ([]capture.Monitor, error) {
	return nil, fmt.Errorf("Wayland capture not implemented")
}

func (c *WaylandCapture) ListWindows() ([]capture.Window, error) {
	return nil, fmt.Errorf("Wayland capture not implemented")
}

func (c *WaylandCapture) CaptureScreen(monitor string) (image.Image, error) {
	return nil, fmt.Errorf("Wayland capture not implemented")
}

func (c *WaylandCapture) CaptureWindow(id capture.WindowID) (image.Image, error) {
	return nil, fmt.Errorf("Wayland capture not implemented")
}

func (c *WaylandCapture) CaptureRegion(x, y, w, h int) (image.Image, error) {
	return nil, fmt.Errorf("Wayland capture not implemented")
}

func (c *WaylandCapture) CaptureAllScreens() (image.Image, error) {
	return nil, fmt.Errorf("Wayland capture not implemented")
}
