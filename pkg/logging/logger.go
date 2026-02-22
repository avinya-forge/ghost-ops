package logging

import (
	"log/slog"
	"os"
)

var logLevel = new(slog.LevelVar)

// InitLogger initializes the global logger with a JSON handler and the specified level.
func InitLogger(level slog.Level) {
	logLevel.Set(level)
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)
}

// SetLevel updates the logging level dynamically.
func SetLevel(level slog.Level) {
	logLevel.Set(level)
}
