package logging

import (
	"context"
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
