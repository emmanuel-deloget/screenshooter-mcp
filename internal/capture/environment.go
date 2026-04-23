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
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
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

// commandExecutor is an interface for executing commands, used for testing.
type commandExecutor interface {
	LookPath(file string) (string, error)
	Run(name string, args ...string) ([]byte, error)
}

// realExecutor implements commandExecutor using os/exec.
type realExecutor struct{}

func (r *realExecutor) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (r *realExecutor) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	return stdout.Bytes(), err
}

// EnvironmentDetector detects the current desktop environment.
//
// EnvironmentDetector provides functionality to identify whether the system
// is running X11 or Wayland. This is used to select the appropriate
// capture implementation for the current session.
type EnvironmentDetector struct {
	executor commandExecutor
}

// NewEnvironmentDetector creates a new EnvironmentDetector.
//
// No initialization is required - the detector is stateless and
// can be created with the default values.
func NewEnvironmentDetector() *EnvironmentDetector {
	return &EnvironmentDetector{
		executor: &realExecutor{},
	}
}

// detectViaLoginctl queries systemd-logind for active sessions to determine
// the desktop environment type. This is useful when running as a system service
// that doesn't have access to session environment variables.
//
// The function lists all sessions via loginctl, then checks the Type property
// of each active session. It returns the first active session's type.
//
// Returns:
//   - EnvironmentX11: Active X11 session found
//   - EnvironmentWayland: Active Wayland session found
//   - EnvironmentUnknown: No active graphical session or loginctl unavailable
//
// Returns an error if loginctl execution fails or session parsing encounters issues.
func (d *EnvironmentDetector) detectViaLoginctl() (Environment, error) {
	loginctl, err := d.executor.LookPath("loginctl")
	if err != nil {
		return EnvironmentUnknown, fmt.Errorf("loginctl not found: %w", err)
	}

	sessionsOutput, err := d.executor.Run(loginctl, "--no-legend", "list-sessions")
	if err != nil {
		return EnvironmentUnknown, fmt.Errorf("failed to list sessions: %w", err)
	}

	sessions := strings.Fields(string(sessionsOutput))
	if len(sessions) == 0 {
		return EnvironmentUnknown, fmt.Errorf("no sessions found")
	}

	for _, sessionID := range sessions {
		typeOutput, err := d.executor.Run(loginctl, "show-session", sessionID, "-p", "State", "-p", "Type")
		if err != nil {
			continue
		}

		properties := string(typeOutput)
		if !strings.Contains(properties, "State=active") {
			continue
		}

		if strings.Contains(properties, "Type=x11") {
			return EnvironmentX11, nil
		}
		if strings.Contains(properties, "Type=wayland") {
			return EnvironmentWayland, nil
		}
	}

	return EnvironmentUnknown, fmt.Errorf("no active graphical session found")
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
//  3. Final fallback: loginctl
//     If no environment variables are set, queries systemd-logind for
//     active sessions to determine the session type.
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

	env, err := d.detectViaLoginctl()
	if err == nil && env != EnvironmentUnknown {
		return env, nil
	}

	return EnvironmentUnknown, fmt.Errorf("could not detect desktop environment (no DISPLAY or WAYLAND_DISPLAY set, loginctl fallback failed: %v)", err)
}
