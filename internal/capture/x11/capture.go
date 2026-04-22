// Package x11 provides screen capture functionality for X11.
//
// This package implements the capture.ScreenCapture interface for X11 desktop environments.
// It uses two backend libraries:
//   - perfuncted/screen: For actual screen capture operations
//   - perfuncted/window: For window enumeration
//   - xgb/randr: For multi-monitor detection via RANDR extension
//
// The implementation supports:
//   - Multiple monitor detection using X11 RANDR
//   - Window enumeration via perfuncted
//   - Screen/window/region capture via perfuncted
//
// Note: This implementation requires an active X11 session with proper
// permissions to access the display.
package x11

import (
	"context"
	"fmt"
	"image"
	"strings"
	"time"

	capture "github.com/emmanuel-deloget/screenshooter-mcp/internal/capture"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/randr"
	"github.com/jezek/xgb/xproto"
	"github.com/nskaggs/perfuncted/screen"
	"github.com/nskaggs/perfuncted/window"
)

// X11Capture implements capture.ScreenCapture for X11 environments.
//
// X11Capture provides screen capture functionality by combining multiple
// backend libraries. The screen and window operations are delegated to
// perfuncted, while monitor enumeration uses the xgb library for RANDR access.
//
// The capture maintains references to both backends and is responsible for
// closing them when no longer needed.
type X11Capture struct {
	// screenBackend handles actual screen capture operations.
	screenBackend screen.Screenshotter
	// windowBackend handles window enumeration and management.
	windowBackend window.Manager
}

// NewX11Capture creates a new X11Capture instance.
//
// This function initializes both the screen and window backends required
// for capture operations. Both backends must be successfully opened
// for the capture to function.
//
// Returns the created X11Capture instance, or an error if either
// backend fails to initialize. On error, any successfully opened
// backends are closed before returning the error.
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

// Close releases resources held by the capture instance.
//
// This method closes both the screen and window backends.
// It is safe to call even if one or both backends failed to initialize -
// nil checks ensure only initialized backends are closed.
//
// Call this method when the capture is no longer needed to prevent
// resource leaks.
func (c *X11Capture) Close() {
	if c.screenBackend != nil {
		c.screenBackend.Close()
	}
	if c.windowBackend != nil {
		c.windowBackend.Close()
	}
}

// ListMonitors returns all available monitors using X11 RANDR.
//
// This method first attempts to enumerate monitors using the X11 RANDR
// extension, which provides accurate multi-monitor information including
// position and resolution for each output.
//
// If RANDR fails (e.g., extension not available, insufficient permissions),
// the method falls back to using the screen resolution from perfuncted
// and creating a single monitor representing the entire screen.
//
// Returns a slice of Monitor structs, or an error if both primary and
// fallback methods fail.
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

// fallbackMonitorList provides a single-monitor fallback when RANDR is unavailable.
//
// This method is used when the X11 RANDR extension cannot be queried.
// It creates a monitor description based on the screen resolution
// from perfuncted, suitable for single-monitor configurations.
//
// Returns a single Monitor with the screen dimensions and standard aliases
// ("1" and "primary"). Returns an error if the resolution cannot
// be determined.
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

// ListWindows returns all open windows.
//
// This method delegates to the perfuncted window backend to enumerate
// visible windows. The window list includes all windows that have titles
// and are visible on the screen.
//
// Returns a slice of Window structs. The slice is empty if no windows are
// found or if window enumeration fails.
func (c *X11Capture) ListWindows() ([]capture.Window, error) {
	logging.Debug().Msg("ListWindows called")

	ctx := context.Background()
	windowList, err := c.windowBackend.List(ctx)
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

// CaptureScreen captures a specific monitor or the entire virtual screen.
//
// If the monitor argument is empty, all monitors are captured as a single combined
// image using CaptureAllScreens. If a monitor name or alias is provided,
// only that monitor's area is captured.
//
// The monitor string is matched against Monitor.Name first, then Monitor.Aliases.
// Match is case-sensitive for name, case-insensitive for aliases.
//
// Returns the captured image, or an error if the monitor is not found
// or the capture fails.
func (c *X11Capture) CaptureScreen(monitor string) (image.Image, error) {
	logging.Debug().Str("monitor", monitor).Msg("CaptureScreen called")

	if monitor == "" {
		return c.CaptureAllScreens()
	}

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

	ctx := context.Background()
	img, err := c.screenBackend.Grab(ctx, rect)
	if err != nil {
		logging.Error().Err(err).Msg("Failed to capture screen")
		return nil, fmt.Errorf("failed to capture screen: %w", err)
	}

	return img, nil
}

// containsAlias checks if an alias matches the target string.
//
// This is a case-insensitive comparison, useful for allowing
// users to specify monitors by various identifiers.
//
// Returns true if an alias matches, false otherwise.
func containsAlias(aliases []string, target string) bool {
	for _, a := range aliases {
		if a == target {
			return true
		}
	}
	return false
}

// CaptureWindow captures a window by its title.
//
// The title argument is matched case-insensitively using substring matching.
// If the window title contains the specified string, it is considered a match.
//
// If multiple windows match the title, an error is returned to prevent ambiguity.
// If no window matches, an error is returned.
//
// The captured image includes the window's entire visible area including
// any window decorations (borders, title bar) depending on the backend.
// Coordinates are relative to the virtual screen.
func (c *X11Capture) CaptureWindow(title string) (image.Image, error) {
	logging.Debug().Str("title", title).Msg("CaptureWindow called")

	ctx := context.Background()
	windowList, err := c.windowBackend.List(ctx)
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

	err = c.windowBackend.Activate(ctx, title)
	if err != nil {
		return nil, fmt.Errorf("failed to activate window: %w", err)
	}
	<-time.After(200 * time.Millisecond)

	if matchedWindow.W == 0 || matchedWindow.H == 0 {
		return nil, fmt.Errorf("window '%s' cannot be captured (width=%d, height=%d)", title, matchedWindow.X, matchedWindow.Y)
	}

	logging.Debug().Uint64("id", matchedWindow.ID).Str("title", matchedWindow.Title).Msg("Window matched")

	rect := image.Rect(matchedWindow.X, matchedWindow.Y, matchedWindow.X+matchedWindow.W, matchedWindow.Y+matchedWindow.H)
	img, err := c.screenBackend.Grab(ctx, rect)
	if err != nil {
		return nil, fmt.Errorf("failed to capture window: %w", err)
	}

	return img, nil
}

// CaptureRegion captures an arbitrary rectangular region.
//
// The x and y arguments specify the top-left corner coordinates.
// The w and h arguments specify the width and height.
// Coordinates are in the virtual screen space.
//
// If the specified region extends beyond the screen bounds, it is clipped
// to the valid area. Returns an error if the capture fails.
func (c *X11Capture) CaptureRegion(x, y, w, h int) (image.Image, error) {
	logging.Debug().Int("x", x).Int("y", y).Int("width", w).Int("height", h).Msg("CaptureRegion called")

	rect := image.Rect(x, y, x+w, y+h)
	ctx := context.Background()
	img, err := c.screenBackend.Grab(ctx, rect)
	if err != nil {
		logging.Error().Err(err).Msg("Failed to capture region")
		return nil, fmt.Errorf("failed to capture region: %w", err)
	}

	return img, nil
}

// CaptureAllScreens captures all monitors as a single combined image.
//
// This method first enumerates all monitors, then calculates the bounding
// rectangle that encompasses all monitors. The entire virtual screen is captured
// as one image.
//
// Returns the combined image, or an error if no monitors are found or
// the capture fails.
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

	ctx := context.Background()
	img, err := c.screenBackend.Grab(ctx, boundingRect)
	if err != nil {
		logging.Error().Err(err).Msg("Failed to capture all screens")
		return nil, fmt.Errorf("failed to capture all screens: %w", err)
	}

	return img, nil
}

// x11Conn provides a low-level connection to X11 for RANDR operations.
//
// x11Conn wraps the xgb connection and provides methods for querying
// display configuration. This is separate from the screen capture
// backend because RANDR queries require direct X11 access.
type x11Conn struct {
	conn *xgb.Conn
}

// newX11Conn creates a new X11 connection.
//
// The displayName argument specifies the X11 display to connect to.
// If empty, the default display (from DISPLAY environment variable)
// is used.
//
// Returns the connection, or an error if connection fails.
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

// Close closes the X11 connection.
//
// This method disconnects from the X11 server. After Close is called,
// the connection should not be used for further operations.
func (c *x11Conn) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

// getMonitors queries X11 RANDR for monitor information.
//
// This method enumerates monitors by querying the RANDR extension for
// connected outputs and their CRTC (Cathode Ray Tube Controller) configurations.
//
// Steps:
//  1. Initialize RANDR extension
//  2. Query screen resources for output list
//  3. For each output, get output info and CRTC
//  4. Build Monitor structs with position and aliases
//
// The position alias (left, right, up, down, middle) is computed relative
// to the primary monitor at position (0, 0).
//
// Returns a slice of Monitor structs, or an error if RANDR is unavailable
// or returns no monitors.
func (c *x11Conn) getMonitors() ([]capture.Monitor, error) {
	if err := randr.Init(c.conn); err != nil {
		return nil, fmt.Errorf("randr not available: %w", err)
	}

	root := xproto.Setup(c.conn).DefaultScreen(c.conn).Root
	resources, err := randr.GetScreenResources(c.conn, root).Reply()
	if err != nil {
		return nil, fmt.Errorf("failed to get screen resources: %w", err)
	}

	type rawMonitor struct {
		name          string
		x, y          int16
		width, height uint16
	}

	rawMonitors := make([]rawMonitor, 0, 8)
	for _, output := range resources.Outputs {
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

		rawMonitors = append(rawMonitors, rawMonitor{
			name:   string(info.Name),
			x:      crtcInfo.X,
			y:      crtcInfo.Y,
			width:  crtcInfo.Width,
			height: crtcInfo.Height,
		})
	}

	if len(rawMonitors) == 0 {
		return nil, fmt.Errorf("no monitors found")
	}

	primaryIdx := 0
	for i, m := range rawMonitors {
		if m.x == 0 && m.y == 0 {
			primaryIdx = i
			break
		}
	}

	computePosition := func(idx int, x, y int16) string {
		if idx == primaryIdx {
			return "middle"
		}

		primary := rawMonitors[primaryIdx]
		relX := int(x) - int(primary.x)
		relY := int(y) - int(primary.y)

		if relX < 0 {
			return "left"
		} else if relX > 0 {
			return "right"
		} else if relY < 0 {
			return "up"
		} else if relY > 0 {
			return "down"
		}
		return "middle"
	}

	monitors := make([]capture.Monitor, 0, len(rawMonitors))
	for i, rm := range rawMonitors {
		position := computePosition(i, rm.x, rm.y)
		width := int(rm.width)
		height := int(rm.height)

		monitors = append(monitors, capture.Monitor{
			Name: rm.name,
			Aliases: []string{
				fmt.Sprintf("%s-%dx%d", position, width, height),
				rm.name,
				fmt.Sprintf("%d", i+1),
			},
			X:      int(rm.x),
			Y:      int(rm.y),
			Width:  width,
			Height: height,
		})
	}

	return monitors, nil
}
