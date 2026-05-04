package logging

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name  string
		level string
		color string
	}{
		{"debug level", "debug", "never"},
		{"info level", "info", "never"},
		{"warn level", "warn", "never"},
		{"error level", "error", "never"},
		{"default level", "invalid", "never"},
		{"color always", "info", "always"},
		{"color never", "info", "never"},
		{"color auto", "info", "auto"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			Init(tt.level, tt.color)
		})
	}
}

func TestLogger(t *testing.T) {
	Init("info", "never")

	logger := Logger()
	if logger == nil {
		t.Error("Logger() returned nil")
	}
}

func TestLoggingOutput(t *testing.T) {
	// Capture output
	var buf bytes.Buffer

	// Temporarily replace stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Init("info", "never")
	Info().Str("key", "value").Msg("test message")

	w.Close()
	os.Stderr = oldStderr

	// Read captured output
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "test message") {
		t.Errorf("expected log output to contain 'test message', got: %s", output)
	}
	if !strings.Contains(output, "info") {
		t.Errorf("expected log output to contain 'info' level, got: %s", output)
	}
}

func TestDebugLogging(t *testing.T) {
	var buf bytes.Buffer

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Init("debug", "never")
	Debug().Msg("debug message")

	w.Close()
	os.Stderr = oldStderr

	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "debug message") {
		t.Errorf("expected log output to contain 'debug message', got: %s", output)
	}
}

func TestWarnLogging(t *testing.T) {
	var buf bytes.Buffer

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Init("warn", "never")
	Warn().Msg("warn message")

	w.Close()
	os.Stderr = oldStderr

	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "warn message") {
		t.Errorf("expected log output to contain 'warn message', got: %s", output)
	}
}

func TestErrorLogging(t *testing.T) {
	var buf bytes.Buffer

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Init("error", "never")
	Error().Err(nil).Msg("error message")

	w.Close()
	os.Stderr = oldStderr

	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "error message") {
		t.Errorf("expected log output to contain 'error message', got: %s", output)
	}
}

func TestInfoLevelFiltersDebug(t *testing.T) {
	var buf bytes.Buffer

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Init("info", "never")
	Debug().Msg("should not appear")
	Info().Msg("should appear")

	w.Close()
	os.Stderr = oldStderr

	buf.ReadFrom(r)
	output := buf.String()

	if strings.Contains(output, "should not appear") {
		t.Error("debug message should not appear at info level")
	}
	if !strings.Contains(output, "should appear") {
		t.Error("info message should appear at info level")
	}
}
