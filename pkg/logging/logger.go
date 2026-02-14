package logging

import (
	"log/slog"
	"os"
)

// InitLogger initializes the global logger with a JSON handler and the specified level.
func InitLogger(level slog.Level) {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)
}
