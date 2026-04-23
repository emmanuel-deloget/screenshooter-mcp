package capture

import (
	"fmt"
	"testing"
)

type mockExecutor struct {
	lookPathFunc func(file string) (string, error)
	runFunc      func(name string, args ...string) ([]byte, error)
}

func (m *mockExecutor) LookPath(file string) (string, error) {
	return m.lookPathFunc(file)
}

func (m *mockExecutor) Run(name string, args ...string) ([]byte, error) {
	return m.runFunc(name, args...)
}

func TestDetectViaLoginctl(t *testing.T) {
	tests := []struct {
		name        string
		lookPath    func(file string) (string, error)
		run         func(name string, args ...string) ([]byte, error)
		expectedEnv Environment
		expectError bool
	}{
		{
			name: "loginctl not found",
			lookPath: func(file string) (string, error) {
				return "", fmt.Errorf("not found")
			},
			run:         nil,
			expectedEnv: EnvironmentUnknown,
			expectError: true,
		},
		{
			name: "list sessions fails",
			lookPath: func(file string) (string, error) {
				return "/usr/bin/loginctl", nil
			},
			run: func(name string, args ...string) ([]byte, error) {
				if len(args) >= 2 && args[0] == "--no-legend" && args[1] == "list-sessions" {
					return nil, fmt.Errorf("failed")
				}
				return nil, nil
			},
			expectedEnv: EnvironmentUnknown,
			expectError: true,
		},
		{
			name: "no sessions",
			lookPath: func(file string) (string, error) {
				return "/usr/bin/loginctl", nil
			},
			run: func(name string, args ...string) ([]byte, error) {
				if len(args) >= 2 && args[0] == "--no-legend" && args[1] == "list-sessions" {
					return []byte(""), nil
				}
				return nil, nil
			},
			expectedEnv: EnvironmentUnknown,
			expectError: true,
		},
		{
			name: "active X11 session",
			lookPath: func(file string) (string, error) {
				return "/usr/bin/loginctl", nil
			},
			run: func(name string, args ...string) ([]byte, error) {
				if len(args) >= 2 && args[0] == "--no-legend" && args[1] == "list-sessions" {
					return []byte("1 seat0 user"), nil
				}
				if len(args) >= 4 && args[0] == "show-session" && args[1] == "1" {
					return []byte("State=active\nType=x11\n"), nil
				}
				return nil, nil
			},
			expectedEnv: EnvironmentX11,
			expectError: false,
		},
		{
			name: "active Wayland session",
			lookPath: func(file string) (string, error) {
				return "/usr/bin/loginctl", nil
			},
			run: func(name string, args ...string) ([]byte, error) {
				if len(args) >= 2 && args[0] == "--no-legend" && args[1] == "list-sessions" {
					return []byte("2 seat0 user"), nil
				}
				if len(args) >= 4 && args[0] == "show-session" && args[1] == "2" {
					return []byte("State=active\nType=wayland\n"), nil
				}
				return nil, nil
			},
			expectedEnv: EnvironmentWayland,
			expectError: false,
		},
		{
			name: "skip inactive session",
			lookPath: func(file string) (string, error) {
				return "/usr/bin/loginctl", nil
			},
			run: func(name string, args ...string) ([]byte, error) {
				if len(args) >= 2 && args[0] == "--no-legend" && args[1] == "list-sessions" {
					return []byte("1 seat0 user 2 seat0 user"), nil
				}
				if len(args) >= 4 && args[0] == "show-session" {
					if args[1] == "1" {
						return []byte("State=closing\nType=x11\n"), nil
					}
					if args[1] == "2" {
						return []byte("State=active\nType=wayland\n"), nil
					}
				}
				return nil, nil
			},
			expectedEnv: EnvironmentWayland,
			expectError: false,
		},
		{
			name: "no graphical sessions",
			lookPath: func(file string) (string, error) {
				return "/usr/bin/loginctl", nil
			},
			run: func(name string, args ...string) ([]byte, error) {
				if len(args) >= 2 && args[0] == "--no-legend" && args[1] == "list-sessions" {
					return []byte("1 seat0 user"), nil
				}
				if len(args) >= 4 && args[0] == "show-session" && args[1] == "1" {
					return []byte("State=active\nType=tty\n"), nil
				}
				return nil, nil
			},
			expectedEnv: EnvironmentUnknown,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &EnvironmentDetector{
				executor: &mockExecutor{
					lookPathFunc: tt.lookPath,
					runFunc:      tt.run,
				},
			}

			env, err := detector.detectViaLoginctl()
			if env != tt.expectedEnv {
				t.Errorf("detectViaLoginctl() environment = %v, want %v", env, tt.expectedEnv)
			}
			if (err != nil) != tt.expectError {
				t.Errorf("detectViaLoginctl() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

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
