package x11

import (
	"fmt"
	"image"

	capture "github.com/emmanuel-deloget/screenshooter-mcp/internal/capture"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
)

type X11Capture struct{}

func NewX11Capture() *X11Capture {
	logging.Debug().Msg("X11Capture created")
	return &X11Capture{}
}

func (c *X11Capture) ListMonitors() ([]capture.Monitor, error) {
	logging.Debug().Msg("ListMonitors called")
	err := fmt.Errorf("X11 capture not implemented")
	logging.Warn().Err(err).Msg("Monitor listing not implemented")
	return nil, err
}

func (c *X11Capture) ListWindows() ([]capture.Window, error) {
	logging.Debug().Msg("ListWindows called")
	err := fmt.Errorf("X11 capture not implemented")
	logging.Warn().Err(err).Msg("Window listing not implemented")
	return nil, err
}

func (c *X11Capture) CaptureScreen(monitor string) (image.Image, error) {
	logging.Debug().Str("monitor", monitor).Msg("CaptureScreen called")
	err := fmt.Errorf("X11 capture not implemented")
	logging.Error().Err(err).Msg("Screen capture not implemented")
	return nil, err
}

func (c *X11Capture) CaptureWindow(id capture.WindowID) (image.Image, error) {
	logging.Debug().Int64("window_id", int64(id)).Msg("CaptureWindow called")
	err := fmt.Errorf("X11 capture not implemented")
	logging.Error().Err(err).Msg("Window capture not implemented")
	return nil, err
}

func (c *X11Capture) CaptureRegion(x, y, w, h int) (image.Image, error) {
	logging.Debug().Int("x", x).Int("y", y).Int("width", w).Int("height", h).Msg("CaptureRegion called")
	err := fmt.Errorf("X11 capture not implemented")
	logging.Error().Err(err).Msg("Region capture not implemented")
	return nil, err
}

func (c *X11Capture) CaptureAllScreens() (image.Image, error) {
	logging.Debug().Msg("CaptureAllScreens called")
	err := fmt.Errorf("X11 capture not implemented")
	logging.Error().Err(err).Msg("All screens capture not implemented")
	return nil, err
}
