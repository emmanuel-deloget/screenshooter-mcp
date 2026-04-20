// Package capture provides screen capture functionality for X11 and Wayland desktops.
//
// This package defines the core interfaces and types for screen capture operations.
// It provides a unified API for capturing screens, windows, and regions that works
// identically across different Linux desktop environments.
//
// The package is organized into several sub-packages:
//
//   - capture/types.go: Core types and interfaces (Monitor, Window, ScreenCapture)
//   - capture/element.go: Element/region types for OCR and coordinate handling
//   - capture/environment.go: Desktop environment detection
//   - capture/x11: X11-specific capture implementation
//   - capture/wayland: Wayland-specific capture implementation
//
// The ScreenCapture interface is implemented by platform-specific packages
// (x11 and wayland) to provide capture functionality. Each implementation
// handles the details of its respective desktop environment.
package capture

import "image"

// Monitor represents a physical display monitor.
//
// A Monitor corresponds to a connected display output in the system.
// The position (X, Y) and dimensions (Width, Height) are expressed in the
// virtual screen coordinate space, where (0, 0) is the top-left corner
// of the primary monitor.
//
// The Name field contains the monitor's system name (e.g., "DP-1" for DisplayPort
// output 1). The Aliases field provides alternative identifiers that can be
// used when specifying which monitor to capture, such as "1", "primary",
// or position-based names like "middle-1920x1080".
type Monitor struct {
	// Name is the system name of the monitor.
	// This is typically the connector name (e.g., "DP-1", "HDMI-1", "eDP-1").
	Name string

	// Aliases provides alternative identifiers for the monitor.
	// Common aliases include position ("left", "right", "up", "down"),
	// index ("1", "2"), and resolution ("1920x1080").
	Aliases []string

	// X is the X coordinate of the monitor's top-left corner
	// in the virtual screen space.
	X int

	// Y is the Y coordinate of the monitor's top-left corner
	// in the virtual screen space.
	Y int

	// Width is the horizontal resolution in pixels.
	Width int

	// Height is the vertical resolution in pixels.
	Height int
}

// Window represents an open window.
//
// A Window corresponds to a top-level window managed by the window manager.
// Small/temporary windows (e.g., tooltips) may not be included
// depending on the backend implementation.
//
// The coordinates (X, Y, Width, Height) describe the window's position
// and size in the virtual screen coordinate space.
// The X and Y represent the top-left corner of the window frame
// (including decorations) if the window manager draws decorations.
type Window struct {
	// ID is a unique identifier for the window.
	// The format is backend-specific (typically an integer ID).
	ID WindowID

	// Name is the window's title string.
	// This is the text displayed in the window title bar.
	Name string

	// X is the X coordinate of the window's top-left corner
	// in the virtual screen space.
	X int

	// Y is the Y coordinate of the window's top-left corner
	// in the virtual screen space.
	Y int

	// Width is the window's width in pixels.
	Width int

	// Height is the window's height in pixels.
	Height int
}

// WindowID is the type used to identify windows.
//
// WindowID is typically an integer type, but the exact format
// and range depend on the backend implementation.
// The ID is only meaningful within a single process and
// may not persist across restarts.
type WindowID int64

// ScreenCapture defines the interface for screen capture operations.
//
// This interface is implemented by platform-specific packages
// (x11 and wayland) to provide capture functionality.
// All methods operate on the virtual screen coordinate space.
//
// Implementations should be safe for concurrent use, though
// concurrent operations may be serialized depending on the backend's
// threading model.
type ScreenCapture interface {
	// ListMonitors returns all available monitors.
	//
	// Returns a slice of Monitor structs describing each connected display.
	// Returns an error if monitor enumeration fails.
	ListMonitors() ([]Monitor, error)

	// ListWindows returns all open windows.
	//
	// Returns a slice of Window structs for all visible windows.
	// Some backends may exclude certain window types
	// (e.g., tooltips, desktop icons).
	// Returns an error if window enumeration fails.
	ListWindows() ([]Window, error)

	// CaptureScreen captures a monitor or all screens.
	//
	// If monitor is empty, captures all monitors as a single combined image.
	// If monitor is specified, captures only that monitor.
	//
	// The monitor string is matched against Monitor.Name and Monitor.Aliases.
	// Returns an error if no monitor matches or capture fails.
	CaptureScreen(monitor string) (image.Image, error)

	// CaptureWindow captures a window by its title.
	//
	// The title argument is matched case-insensitively using substring matching.
	// If multiple windows match, returns an error to prevent ambiguity.
	// If no window matches, returns an error.
	//
	// Returns the captured image, or an error if capture fails.
	CaptureWindow(title string) (image.Image, error)

	// CaptureRegion captures an arbitrary rectangular region.
	//
	// The x and y arguments specify the top-left corner.
	// The w and h arguments specify the dimensions.
	// Coordinates are in the virtual screen space.
	//
	// If the region extends beyond screen bounds, it is clipped.
	// Returns the captured image, or an error if capture fails.
	CaptureRegion(x, y, w, h int) (image.Image, error)

	// CaptureAllScreens captures all monitors as a single combined image.
	//
	// This is equivalent to calling CaptureScreen("") but explicitly
	// captures the entire virtual screen encompassing all monitors.
	//
	// Returns the combined image, or an error if capture fails.
	CaptureAllScreens() (image.Image, error)
}
