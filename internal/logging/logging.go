// Package logging provides structured logging for the MCP server.
//
// This package wraps the zerolog library to provide structured logging
// with configurable log levels and optional color support.
//
// The logger is initialized with Init() before any logging calls.
// Log levels can be set to debug, info, warn, or error.
// Color output can be set to always, never, or auto (colored if terminal).
//
// Log messages are written to stderr. The console writer provides
// human-readable output with timestamps and color highlighting.
package logging

import (
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
)

// log is the global logger instance.
//
// This logger is initialized by Init() and used by all log level
// functions. It is a package-level variable to simplify
// logging calls throughout the application.
var log zerolog.Logger

// Init initializes the logging system.
//
// The level argument controls which log messages are output:
//   - debug: All messages (debug, info, warn, error)
//   - info: Info, warn, and error messages
//   - warn: Warn and error messages
//   - error: Only error messages
//
// The color argument controls ANSI color codes in output:
//   - always: Always use colors
//   - never: Never use colors
//   - auto: Use colors if writing to a terminal
//
// This function must be called before any logging functions.
// It is typically called at application startup.
func Init(level string, color string) {
	var output io.Writer = os.Stderr

	useColor := false
	switch color {
	case "always":
		useColor = true
	case "never":
		useColor = false
	case "auto":
		useColor = isatty.IsTerminal(os.Stderr.Fd())
	}

	if useColor {
		output = zerolog.ConsoleWriter{Out: os.Stderr}
	}

	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"
	log = zerolog.New(output).With().Timestamp().Logger()

	switch level {
	case "debug":
		log = log.Level(zerolog.DebugLevel)
	case "info":
		log = log.Level(zerolog.InfoLevel)
	case "warn":
		log = log.Level(zerolog.WarnLevel)
	case "error":
		log = log.Level(zerolog.ErrorLevel)
	default:
		log = log.Level(zerolog.InfoLevel)
	}
}

// Logger returns the underlying zerolog.Logger.
//
// This can be used for advanced logging operations not
// covered by the helper functions.
//
// Returns a pointer to the internal logger instance.
func Logger() *zerolog.Logger {
	return &log
}

// Info logs a message at the info level.
//
// Info is used for normal operational messages that indicate
// the progress or status of the application.
//
// Returns an event for chaining (e.g., .Str("key", "value").Msg("message")).
func Info() *zerolog.Event {
	return log.Info()
}

// Debug logs a message at the debug level.
//
// Debug is used for detailed diagnostic information useful for
// troubleshooting. These messages are only output when log level
// is set to "debug".
//
// Returns an event for chaining.
func Debug() *zerolog.Event {
	return log.Debug()
}

// Warn logs a message at the warn level.
//
// Warn is used for warning messages that indicate potential
// issues but don't prevent operation.
//
// Returns an event for chaining.
func Warn() *zerolog.Event {
	return log.Warn()
}

// Error logs a message at the error level.
//
// Error is used for error messages that indicate failures
// in operations. These are always output regardless
// of log level.
//
// Returns an event for chaining.
func Error() *zerolog.Event {
	return log.Error()
}
