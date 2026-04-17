package capture

import (
	"fmt"
	"os"
)

type Environment string

const (
	EnvironmentX11     Environment = "x11"
	EnvironmentWayland Environment = "wayland"
	EnvironmentUnknown Environment = "unknown"
)

type EnvironmentDetector struct{}

func NewEnvironmentDetector() *EnvironmentDetector {
	return &EnvironmentDetector{}
}

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
