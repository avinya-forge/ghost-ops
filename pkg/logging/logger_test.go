package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
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

func TestInitLoggerWithWriter_TraceHandler(t *testing.T) {
	var buf bytes.Buffer
	InitLoggerWithWriter(slog.LevelInfo, &buf)

	// Create a dummy tracer and context
	ctx := context.Background()

	// Use slog.InfoContext directly to pass the context down.
	slog.InfoContext(ctx, "test message")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	assert.NotContains(t, logEntry, "trace_id")
	assert.NotContains(t, logEntry, "span_id")
}

func TestInitLoggerWithWriter_TraceHandlerWithTrace(t *testing.T) {
	var buf bytes.Buffer
	InitLoggerWithWriter(slog.LevelInfo, &buf)

	// Create a dummy trace ID and span ID
	traceID, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	spanID, _ := trace.SpanIDFromHex("0102030405060708")

	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})

	ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)

	slog.InfoContext(ctx, "test message")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	assert.Equal(t, "0102030405060708090a0b0c0d0e0f10", logEntry["trace_id"])
	assert.Equal(t, "0102030405060708", logEntry["span_id"])
}

