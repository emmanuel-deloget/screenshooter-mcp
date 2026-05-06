package logging

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name   string
		level  string
		color  string
		format string
	}{
		{"debug level", "debug", "never", "text"},
		{"info level", "info", "never", "text"},
		{"warn level", "warn", "never", "text"},
		{"error level", "error", "never", "text"},
		{"default level", "invalid", "never", "text"},
		{"color always", "info", "always", "text"},
		{"color never", "info", "never", "text"},
		{"color auto", "info", "auto", "text"},
		{"json format", "info", "never", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			Init(tt.level, tt.color, tt.format)
		})
	}
}

func TestLogger(t *testing.T) {
	Init("info", "never", "text")

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

	Init("info", "never", "text")
	Info().Str("key", "value").Msg("test message")

	w.Close()
	os.Stderr = oldStderr

	// Read captured output
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "test message") {
		t.Errorf("expected log output to contain 'test message', got: %s", output)
	}
	if !strings.Contains(output, "INF") && !strings.Contains(output, "info") {
		t.Errorf("expected log output to contain 'INF' or 'info' level, got: %s", output)
	}
}

func TestDebugLogging(t *testing.T) {
	var buf bytes.Buffer

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Init("debug", "never", "text")
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

	Init("warn", "never", "text")
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

	Init("error", "never", "text")
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

	Init("info", "never", "text")
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
