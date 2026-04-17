package x11

import (
	"fmt"
	"image"

	capture "github.com/emmanuel-deloget/screenshooter-mcp/internal/capture"
)

type X11Capture struct{}

func NewX11Capture() *X11Capture {
	return &X11Capture{}
}

func (c *X11Capture) ListMonitors() ([]capture.Monitor, error) {
	return nil, fmt.Errorf("X11 capture not implemented")
}

func (c *X11Capture) ListWindows() ([]capture.Window, error) {
	return nil, fmt.Errorf("X11 capture not implemented")
}

func (c *X11Capture) CaptureScreen(monitor string) (image.Image, error) {
	return nil, fmt.Errorf("X11 capture not implemented")
}

func (c *X11Capture) CaptureWindow(id capture.WindowID) (image.Image, error) {
	return nil, fmt.Errorf("X11 capture not implemented")
}

func (c *X11Capture) CaptureRegion(x, y, w, h int) (image.Image, error) {
	return nil, fmt.Errorf("X11 capture not implemented")
}

func (c *X11Capture) CaptureAllScreens() (image.Image, error) {
	return nil, fmt.Errorf("X11 capture not implemented")
}
