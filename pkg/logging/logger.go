package logging

import (
	"log/slog"
	"os"
)

// InitLogger initializes the global logger with a JSON handler.
func InitLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}
