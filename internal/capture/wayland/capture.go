// Package wayland provides screen capture functionality for Wayland.
//
// This package implements the capture.ScreenCapture interface for Wayland desktop environments.
// It uses the perfuncted library, which provides portal-based screen capture
// that works with modern Wayland compositors.
//
// The implementation supports:
//   - Screen capture via libportal (portal-based, compositor-agnostic)
//   - Window enumeration via perfuncted
//   - Monitor detection (limited depending on compositor support)
//
// Note: This implementation requires the XDG desktop portal to be available
// and properly configured for screen capture.
package wayland

import (
	"context"
	"fmt"
	"image"
	"strings"

	capture "github.com/emmanuel-deloget/screenshooter-mcp/internal/capture"
	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
	"github.com/nskaggs/perfuncted/screen"
	"github.com/nskaggs/perfuncted/window"
)

// WaylandCapture implements capture.ScreenCapture for Wayland environments.
//
// WaylandCapture provides screen capture functionality using the perfuncted
// library, which uses the XDG desktop portal for screen capture. This approach
// delegates to the compositor via DBus, providing a consistent API across
// different Wayland compositors (GNOME Shell, KDE Plasma, Sway, etc.).
//
// The capture maintains references to both backends and is responsible for
// closing them when no longer needed.
//
// When the window backend cannot be opened (e.g., due to missing portals or
// permissions), the capture operates in degraded mode where window-related
// operations (ListWindows, CaptureWindow) return an error. Screen capture
// and monitor enumeration continue to work in this mode.
type WaylandCapture struct {
	// screenBackend handles actual screen capture operations.
	screenBackend screen.Screenshotter
	// windowBackend handles window enumeration and management.
	windowBackend window.Manager
	// windowAvailable indicates whether the window backend is available.
	// When false, window-related operations return an error.
	windowAvailable bool
}

// NewWaylandCapture creates a new WaylandCapture instance.
//
// This function initializes the screen backend required for capture
// operations. The window backend is optional - if it fails to initialize
// (e.g., due to missing portals or permissions), the capture operates
// in degraded mode where window-related operations (ListWindows,
// CaptureWindow) return an error.
//
// The screen backend uses the desktop portal via DBus, so it may prompt
// the user for permission on first use depending on the compositor's
// permission settings.
//
// Returns the created WaylandCapture instance, or an error if the
// screen backend fails to initialize. On error, any successfully
// opened backends are closed before returning the error.
//
// When operating in degraded mode, a warning is logged but no
// error is returned - the capture still functions for screen capture.
func NewWaylandCapture() (*WaylandCapture, error) {
	logging.Debug().Msg("WaylandCapture creating screen backend")
	sc, err := screen.Open()
	if err != nil {
		logging.Error().Err(err).Msg("Failed to open screen backend")
		return nil, fmt.Errorf("failed to open screen backend: %w", err)
	}

	logging.Debug().Msg("WaylandCapture creating window backend")
	win, err := window.Open()
	if err != nil {
		logging.Warn().Err(err).Msg("Window backend unavailable, try to use our Gnome extension")
		win, err = NewGnomeManager()
		if err != nil {
			logging.Warn().Err(err).Msg("Window backend unavailable, operating in degraded mode")
			return &WaylandCapture{
				screenBackend:   sc,
				windowBackend:   nil,
				windowAvailable: false,
			}, nil
		}
	}
	logging.Debug().Msg("WaylandCapture created successfully")
	return &WaylandCapture{
		screenBackend:   sc,
		windowBackend:   win,
		windowAvailable: true,
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
func (c *WaylandCapture) Close() {
	if c.screenBackend != nil {
		c.screenBackend.Close()
	}
	if c.windowBackend != nil {
		c.windowBackend.Close()
	}
}

// ListMonitors returns all available monitors.
//
// This method attempts to get monitor information from Wayland, but due
// to limited compositor support in Wayland, it often falls back to single-monitor
// detection based on screen resolution.
//
// The portal-based capture does not provide reliable multi-monitor enumeration,
// so this implementation returns a single monitor matching the screen resolution.
// This is generally sufficient for most capture operations.
//
// Returns a slice of Monitor structs. The slice contains a single monitor
// representing the entire screen.
func (c *WaylandCapture) ListMonitors() ([]capture.Monitor, error) {
	logging.Debug().Msg("ListMonitors called")

	monitors, err := c.getMonitorsFromWayland()
	if err != nil {
		logging.Warn().Err(err).Msg("Failed to get monitors from Wayland, using fallback")
		return c.fallbackMonitorList()
	}

	return monitors, nil
}

// getMonitorsFromWayland queries the screen backend for monitor information.
//
// Currently, this function returns a single monitor based on the screen
// resolution. Comprehensive multi-monitor support via Wayland is not
// available through the portal API.
//
// Returns a single Monitor with standard aliases.
func (c *WaylandCapture) getMonitorsFromWayland() ([]capture.Monitor, error) {
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

// fallbackMonitorList provides a single-monitor fallback.
//
// This is identical to getMonitorsFromWayland and exists for API compatibility.
// Returns a single Monitor based on screen resolution.
func (c *WaylandCapture) fallbackMonitorList() ([]capture.Monitor, error) {
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
//
// If the window backend is unavailable (e.g., missing portals or
// permissions), returns an error indicating the feature is unavailable.
func (c *WaylandCapture) ListWindows() ([]capture.Window, error) {
	logging.Debug().Msg("ListWindows called")

	if !c.windowAvailable {
		return nil, fmt.Errorf("window enumeration unavailable: window backend not initialized (likely missing XDG portal permissions)")
	}

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
// Note: Due to limited multi-monitor support in Wayland, specifying a
// non-existent monitor may succeed with unexpected results depending on
// the compositor.
//
// Returns the captured image, or an error if the monitor is not found
// or the capture fails.
func (c *WaylandCapture) CaptureScreen(monitor string) (image.Image, error) {
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
// The captured image includes the window's entire visible area.
// Coordinates are relative to the virtual screen.
//
// If the window backend is unavailable (e.g., missing portals or
// permissions), returns an error indicating the feature is unavailable.
func (c *WaylandCapture) CaptureWindow(title string) (image.Image, error) {
	logging.Debug().Str("title", title).Msg("CaptureWindow called")

	if !c.windowAvailable {
		return nil, fmt.Errorf("window capture unavailable: window backend not initialized (likely missing XDG portal permissions)")
	}

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
func (c *WaylandCapture) CaptureRegion(x, y, w, h int) (image.Image, error) {
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
func (c *WaylandCapture) CaptureAllScreens() (image.Image, error) {
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
