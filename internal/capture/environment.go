// Package environment provides desktop environment detection for Linux.
//
// This package helps determine whether the system is running X11 or Wayland,
// which is essential for selecting the appropriate capture backend.
//
// Detection is performed by examining environment variables set by display
// servers and login managers. The detection algorithm checks multiple
// sources to handle various system configurations.
package capture

import (
	"fmt"
	"os"
)

// Environment represents the detected desktop environment.
//
// Environment is a string type that identifies which Linux display
// system is in use. The value is used to select the appropriate
// capture implementation.
type Environment string

const (
	// EnvironmentX11 indicates the system is running the X11 Window System.
	// X11 is the traditional Linux display server, using the DISPLAY
	// environment variable to identify the display.
	EnvironmentX11 Environment = "x11"

	// EnvironmentWayland indicates the system is running the Wayland display server.
	// Wayland is a modern display server protocol, using WAYLAND_DISPLAY
	// to identify the Wayland socket.
	EnvironmentWayland Environment = "wayland"

	// EnvironmentUnknown indicates the desktop environment could not be detected.
	// This occurs when neither X11 nor Wayland environment variables are set.
	EnvironmentUnknown Environment = "unknown"
)

// EnvironmentDetector detects the current desktop environment.
//
// EnvironmentDetector provides functionality to identify whether the system
// is running X11 or Wayland. This is used to select the appropriate
// capture implementation for the current session.
type EnvironmentDetector struct{}

// NewEnvironmentDetector creates a new EnvironmentDetector.
//
// No initialization is required - the detector is stateless and
// can be created with the default values.
func NewEnvironmentDetector() *EnvironmentDetector {
	return &EnvironmentDetector{}
}

// Detect determines the current desktop environment.
//
// The detection algorithm uses a two-stage approach:
//
//  1. Primary check: XDG_SESSION_TYPE
//     Most modern systems set this environment variable during login.
//     If set, it directly indicates "x11" or "wayland" and is
//     combined with checking for the appropriate display variable.
//
//  2. Fallback check: DISPLAY and WAYLAND_DISPLAY
//     If XDG_SESSION_TYPE is not set, the function checks for DISPLAY
//     (X11) or WAYLAND_DISPLAY (Wayland) environment variables.
//
// The function requires BOTH the session type AND the display variable
// to be present for a positive match. This prevents false positives
// when environment variables are set but not actually in use.
//
// Returns:
//   - EnvironmentX11: X11 detected
//   - EnvironmentWayland: Wayland detected
//   - EnvironmentUnknown: Could not detect
//
// Returns an error if no desktop environment can be determined.
func (d *EnvironmentDetector) Detect() (Environment, error) {
	sessionType := os.Getenv("XDG_SESSION_TYPE")
	if sessionType != "" {
		switch sessionType {
		case "x11":
			if os.Getenv("DISPLAY") != "" {
				return EnvironmentX11, nil
			}
		case "wayland":
			if os.Getenv("WAYLAND_DISPLAY") != "" {
				return EnvironmentWayland, nil
			}
		}
	}

	if os.Getenv("DISPLAY") != "" {
		return EnvironmentX11, nil
	}

	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return EnvironmentWayland, nil
	}

	return EnvironmentUnknown, fmt.Errorf("could not detect desktop environment (no DISPLAY or WAYLAND_DISPLAY set)")
}
