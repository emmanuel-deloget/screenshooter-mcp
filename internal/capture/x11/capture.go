package x11

import (
	"fmt"
	"image"
	"strings"

	capture "github.com/emmanuel-deloget/screenshooter-mcp/internal/capture"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/randr"
	"github.com/jezek/xgb/xproto"
	"github.com/nskaggs/perfuncted/screen"
	"github.com/nskaggs/perfuncted/window"
)

type X11Capture struct {
	screenBackend screen.Screenshotter
	windowBackend window.Manager
}

func NewX11Capture() (*X11Capture, error) {
	logging.Debug().Msg("X11Capture creating screen backend")
	sc, err := screen.Open()
	if err != nil {
		logging.Error().Err(err).Msg("Failed to open screen backend")
		return nil, fmt.Errorf("failed to open screen backend: %w", err)
	}

	logging.Debug().Msg("X11Capture creating window backend")
	win, err := window.Open()
	if err != nil {
		logging.Error().Err(err).Msg("Failed to open window backend")
		sc.Close()
		return nil, fmt.Errorf("failed to open window backend: %w", err)
	}

	logging.Debug().Msg("X11Capture created successfully")
	return &X11Capture{
		screenBackend: sc,
		windowBackend: win,
	}, nil
}

func (c *X11Capture) Close() {
	if c.screenBackend != nil {
		c.screenBackend.Close()
	}
	if c.windowBackend != nil {
		c.windowBackend.Close()
	}
}

func (c *X11Capture) ListMonitors() ([]capture.Monitor, error) {
	logging.Debug().Msg("ListMonitors called")

	conn, err := newX11Conn("")
	if err != nil {
		logging.Warn().Err(err).Msg("Failed to connect to X11, using fallback")
		return c.fallbackMonitorList()
	}
	defer conn.Close()

	monitors, err := conn.getMonitors()
	if err != nil {
		logging.Warn().Err(err).Msg("Failed to get monitors from X11, using fallback")
		return c.fallbackMonitorList()
	}

	return monitors, nil
}

func (c *X11Capture) fallbackMonitorList() ([]capture.Monitor, error) {
	width, height, err := screen.Resolution(c.screenBackend)
	if err != nil {
		return nil, fmt.Errorf("failed to get screen resolution: %w", err)
	}

	return []capture.Monitor{
		{
			Name:    fmt.Sprintf("%dx%d-1", width, height),
			Aliases: []string{"1", "primary"},
			X:       0,
			Y:       0,
			Width:   width,
			Height:  height,
		},
	}, nil
}

func (c *X11Capture) ListWindows() ([]capture.Window, error) {
	logging.Debug().Msg("ListWindows called")

	windowList, err := c.windowBackend.List()
	if err != nil {
		logging.Error().Err(err).Msg("Failed to list windows")
		return nil, fmt.Errorf("failed to list windows: %w", err)
	}

	windows := make([]capture.Window, 0, len(windowList))
	for _, w := range windowList {
		windows = append(windows, capture.Window{
			ID:     capture.WindowID(w.ID),
			Name:   w.Title,
			X:      w.X,
			Y:      w.Y,
			Width:  w.W,
			Height: w.H,
		})
		logging.Debug().Uint64("id", w.ID).Str("title", w.Title).Int("x", w.X).Int("y", w.Y).Int("width", w.W).Int("height", w.H).Msg("Found window")
	}

	return windows, nil
}

func (c *X11Capture) CaptureScreen(monitor string) (image.Image, error) {
	logging.Debug().Str("monitor", monitor).Msg("CaptureScreen called")

	monitors, err := c.ListMonitors()
	if err != nil {
		return nil, fmt.Errorf("failed to list monitors: %w", err)
	}

	var targetMonitor *capture.Monitor
	for _, m := range monitors {
		if m.Name == monitor || containsAlias(m.Aliases, monitor) {
			targetMonitor = &m
			break
		}
	}

	if targetMonitor == nil {
		return nil, fmt.Errorf("monitor not found: %s", monitor)
	}

	rect := image.Rect(targetMonitor.X, targetMonitor.Y, targetMonitor.X+targetMonitor.Width, targetMonitor.Y+targetMonitor.Height)
	logging.Debug().Interface("rect", rect).Msg("Capturing screen region")

	img, err := c.screenBackend.Grab(rect)
	if err != nil {
		logging.Error().Err(err).Msg("Failed to capture screen")
		return nil, fmt.Errorf("failed to capture screen: %w", err)
	}

	return img, nil
}

func containsAlias(aliases []string, target string) bool {
	for _, a := range aliases {
		if a == target {
			return true
		}
	}
	return false
}

func (c *X11Capture) CaptureWindow(title string) (image.Image, error) {
	logging.Debug().Str("title", title).Msg("CaptureWindow called")

	windowList, err := c.windowBackend.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list windows: %w", err)
	}

	var matchedWindow *window.Info
	for _, w := range windowList {
		if strings.Contains(strings.ToLower(w.Title), strings.ToLower(title)) {
			if matchedWindow != nil {
				return nil, fmt.Errorf("multiple windows found matching '%s'", title)
			}
			matchedWindow = &w
		}
	}

	if matchedWindow == nil {
		return nil, fmt.Errorf("window not found: %s", title)
	}

	logging.Debug().Uint64("id", matchedWindow.ID).Str("title", matchedWindow.Title).Msg("Window matched")

	rect := image.Rect(matchedWindow.X, matchedWindow.Y, matchedWindow.X+matchedWindow.W, matchedWindow.Y+matchedWindow.H)
	img, err := c.screenBackend.Grab(rect)
	if err != nil {
		return nil, fmt.Errorf("failed to capture window: %w", err)
	}

	return img, nil
}

func (c *X11Capture) CaptureRegion(x, y, w, h int) (image.Image, error) {
	logging.Debug().Int("x", x).Int("y", y).Int("width", w).Int("height", h).Msg("CaptureRegion called")

	rect := image.Rect(x, y, x+w, y+h)
	img, err := c.screenBackend.Grab(rect)
	if err != nil {
		logging.Error().Err(err).Msg("Failed to capture region")
		return nil, fmt.Errorf("failed to capture region: %w", err)
	}

	return img, nil
}

func (c *X11Capture) CaptureAllScreens() (image.Image, error) {
	logging.Debug().Msg("CaptureAllScreens called")

	monitors, err := c.ListMonitors()
	if err != nil {
		return nil, fmt.Errorf("failed to list monitors: %w", err)
	}

	if len(monitors) == 0 {
		return nil, fmt.Errorf("no monitors found")
	}

	minX := monitors[0].X
	minY := monitors[0].Y
	maxX := monitors[0].X + monitors[0].Width
	maxY := monitors[0].Y + monitors[0].Height

	for _, m := range monitors[1:] {
		if m.X < minX {
			minX = m.X
		}
		if m.Y < minY {
			minY = m.Y
		}
		if m.X+m.Width > maxX {
			maxX = m.X + m.Width
		}
		if m.Y+m.Height > maxY {
			maxY = m.Y + m.Height
		}
	}

	boundingRect := image.Rect(minX, minY, maxX, maxY)
	logging.Debug().Interface("bounding_rect", boundingRect).Int("monitor_count", len(monitors)).Msg("Capturing all screens")

	img, err := c.screenBackend.Grab(boundingRect)
	if err != nil {
		logging.Error().Err(err).Msg("Failed to capture all screens")
		return nil, fmt.Errorf("failed to capture all screens: %w", err)
	}

	return img, nil
}

type x11Conn struct {
	conn *xgb.Conn
}

func newX11Conn(displayName string) (*x11Conn, error) {
	var conn *xgb.Conn
	var err error
	if displayName == "" {
		conn, err = xgb.NewConn()
	} else {
		conn, err = xgb.NewConnDisplay(displayName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to X11: %w", err)
	}
	return &x11Conn{conn: conn}, nil
}

func (c *x11Conn) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

func (c *x11Conn) getMonitors() ([]capture.Monitor, error) {
	if err := randr.Init(c.conn); err != nil {
		return nil, fmt.Errorf("randr not available: %w", err)
	}

	root := xproto.Setup(c.conn).DefaultScreen(c.conn).Root
	resources, err := randr.GetScreenResources(c.conn, root).Reply()
	if err != nil {
		return nil, fmt.Errorf("failed to get screen resources: %w", err)
	}

	monitors := make([]capture.Monitor, 0, 8)
	for i, output := range resources.Outputs {
		info, err := randr.GetOutputInfo(c.conn, output, resources.ConfigTimestamp).Reply()
		if err != nil {
			continue
		}

		if info.Crtc == 0 {
			continue
		}

		crtcInfo, err := randr.GetCrtcInfo(c.conn, info.Crtc, resources.ConfigTimestamp).Reply()
		if err != nil {
			continue
		}

		position := "left"
		if i > 0 {
			position = fmt.Sprintf("%d", i+1)
		}

		monitors = append(monitors, capture.Monitor{
			Name:    fmt.Sprintf("%dx%d-%s", int(crtcInfo.Width), int(crtcInfo.Height), position),
			Aliases: []string{string(info.Name), fmt.Sprintf("%d", i+1)},
			X:       int(crtcInfo.X),
			Y:       int(crtcInfo.Y),
			Width:   int(crtcInfo.Width),
			Height:  int(crtcInfo.Height),
		})
	}

	if len(monitors) == 0 {
		return nil, fmt.Errorf("no monitors found")
	}

	return monitors, nil
}
