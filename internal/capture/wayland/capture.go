package wayland

import (
	"fmt"
	"image"

	capture "github.com/emmanuel-deloget/screenshooter-mcp/internal/capture"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
)

type WaylandCapture struct{}

func NewWaylandCapture() *WaylandCapture {
	logging.Debug().Msg("WaylandCapture created")
	return &WaylandCapture{}
}

func (c *WaylandCapture) ListMonitors() ([]capture.Monitor, error) {
	logging.Debug().Msg("ListMonitors called")
	err := fmt.Errorf("Wayland capture not implemented")
	logging.Warn().Err(err).Msg("Monitor listing not implemented")
	return nil, err
}

func (c *WaylandCapture) ListWindows() ([]capture.Window, error) {
	logging.Debug().Msg("ListWindows called")
	err := fmt.Errorf("Wayland capture not implemented")
	logging.Warn().Err(err).Msg("Window listing not implemented")
	return nil, err
}

func (c *WaylandCapture) CaptureScreen(monitor string) (image.Image, error) {
	logging.Debug().Str("monitor", monitor).Msg("CaptureScreen called")
	err := fmt.Errorf("Wayland capture not implemented")
	logging.Error().Err(err).Msg("Screen capture not implemented")
	return nil, err
}

func (c *WaylandCapture) CaptureWindow(id capture.WindowID) (image.Image, error) {
	logging.Debug().Int64("window_id", int64(id)).Msg("CaptureWindow called")
	err := fmt.Errorf("Wayland capture not implemented")
	logging.Error().Err(err).Msg("Window capture not implemented")
	return nil, err
}

func (c *WaylandCapture) CaptureRegion(x, y, w, h int) (image.Image, error) {
	logging.Debug().Int("x", x).Int("y", y).Int("width", w).Int("height", h).Msg("CaptureRegion called")
	err := fmt.Errorf("Wayland capture not implemented")
	logging.Error().Err(err).Msg("Region capture not implemented")
	return nil, err
}

func (c *WaylandCapture) CaptureAllScreens() (image.Image, error) {
	logging.Debug().Msg("CaptureAllScreens called")
	err := fmt.Errorf("Wayland capture not implemented")
	logging.Error().Err(err).Msg("All screens capture not implemented")
	return nil, err
}
