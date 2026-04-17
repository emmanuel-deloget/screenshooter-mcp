package capture

import (
	"testing"
)

func TestEnvironmentDetector(t *testing.T) {
	tests := []struct {
		name        string
		sessionType string
		display     string
		waylandDisp string
		expectedEnv Environment
		expectError bool
	}{
		{
			name:        "X11 via XDG_SESSION_TYPE",
			sessionType: "x11",
			display:     ":0",
			expectedEnv: EnvironmentX11,
			expectError: false,
		},
		{
			name:        "Wayland via XDG_SESSION_TYPE",
			sessionType: "wayland",
			waylandDisp: "wayland-0",
			expectedEnv: EnvironmentWayland,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := detectEnvironment(tt.sessionType, tt.display, tt.waylandDisp)
			if env != tt.expectedEnv {
				t.Errorf("Detected environment = %v, want %v", env, tt.expectedEnv)
			}
		})
	}
}

func detectEnvironment(sessionType, display, waylandDisplay string) Environment {
	if sessionType != "" {
		switch sessionType {
		case "x11":
			if display != "" {
				return EnvironmentX11
			}
		case "wayland":
			if waylandDisplay != "" {
				return EnvironmentWayland
			}
		}
	}

	if display != "" {
		return EnvironmentX11
	}

	if waylandDisplay != "" {
		return EnvironmentWayland
	}

	return EnvironmentUnknown
}
