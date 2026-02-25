package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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

func TestTraceHandler(t *testing.T) {
	var buf bytes.Buffer
	InitLoggerWithWriter(slog.LevelInfo, &buf)

	ctx := context.Background()

	// Create a trace
	tp := sdktrace.NewTracerProvider()
	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(ctx, "test-span")
	defer span.End()

	slog.InfoContext(ctx, "trace message")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	assert.Equal(t, "trace message", logEntry["msg"])
	assert.Equal(t, span.SpanContext().TraceID().String(), logEntry["trace_id"])
	assert.Equal(t, span.SpanContext().SpanID().String(), logEntry["span_id"])
}

func TestTraceHandler_With(t *testing.T) {
	var buf bytes.Buffer
	InitLoggerWithWriter(slog.LevelInfo, &buf)

	ctx := context.Background()
	tp := sdktrace.NewTracerProvider()
	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(ctx, "test-span")
	defer span.End()

	// Use With to create a child logger
	logger := slog.Default().With("component", "test")
	logger.InfoContext(ctx, "child message")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	assert.Equal(t, "child message", logEntry["msg"])
	assert.Equal(t, "test", logEntry["component"])
	assert.Equal(t, span.SpanContext().TraceID().String(), logEntry["trace_id"])
}
