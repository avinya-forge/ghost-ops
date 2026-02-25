package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetLevel(t *testing.T) {
	InitLogger(slog.LevelInfo)

	// Check initial level
	assert.True(t, slog.Default().Enabled(context.Background(), slog.LevelInfo))
	assert.False(t, slog.Default().Enabled(context.Background(), slog.LevelDebug))

	// Change level to Debug
	SetLevel(slog.LevelDebug)
	assert.True(t, slog.Default().Enabled(context.Background(), slog.LevelDebug))

	// Change level to Warn
	SetLevel(slog.LevelWarn)
	assert.False(t, slog.Default().Enabled(context.Background(), slog.LevelInfo))
	assert.True(t, slog.Default().Enabled(context.Background(), slog.LevelWarn))
}

func TestInitLoggerWithWriter(t *testing.T) {
	var buf bytes.Buffer
	InitLoggerWithWriter(slog.LevelInfo, &buf)

	slog.Info("test message", "key", "value")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	assert.Equal(t, "INFO", logEntry["level"])
	assert.Equal(t, "test message", logEntry["msg"])
	assert.Equal(t, "value", logEntry["key"])
	assert.Contains(t, logEntry, "time")
}

func TestInitLoggerWithWriter_Level(t *testing.T) {
	var buf bytes.Buffer
	InitLoggerWithWriter(slog.LevelWarn, &buf)

	slog.Info("should not appear")
	slog.Warn("should appear")

	output := buf.String()
	assert.NotContains(t, output, "should not appear")
	assert.Contains(t, output, "should appear")
}
