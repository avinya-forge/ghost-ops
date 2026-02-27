package logging

import (
	"io"
	"log/slog"
	"os"
)

var logLevel = new(slog.LevelVar)

// InitLogger initializes the global logger with a JSON handler and the specified level.
func InitLogger(level slog.Level) {
	InitLoggerWithWriter(level, os.Stdout)
}

// InitLoggerWithWriter initializes the global logger with a JSON handler, the specified level, and output writer.
func InitLoggerWithWriter(level slog.Level, w io.Writer) {
	logLevel.Set(level)
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	jsonHandler := slog.NewJSONHandler(w, opts)
	traceHandler := NewTraceHandler(jsonHandler)
	logger := slog.New(traceHandler)
	slog.SetDefault(logger)
}

// SetLevel updates the logging level dynamically.
func SetLevel(level slog.Level) {
	logLevel.Set(level)
}
