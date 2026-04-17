package logging

import (
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
)

var log zerolog.Logger

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

func Logger() *zerolog.Logger {
	return &log
}

func Info() *zerolog.Event {
	return log.Info()
}

func Debug() *zerolog.Event {
	return log.Debug()
}

func Warn() *zerolog.Event {
	return log.Warn()
}

func Error() *zerolog.Event {
	return log.Error()
}
