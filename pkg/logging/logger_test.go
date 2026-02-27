package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

func TestSetLevel(t *testing.T) {
	var buf bytes.Buffer
	InitLoggerWithWriter(slog.LevelInfo, &buf)

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

	// Setup Tracer
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	tracer := tp.Tracer("test-tracer")

	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	logger := slog.Default()
	logger.InfoContext(ctx, "traced message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	if logEntry["trace_id"] != traceID {
		t.Errorf("Expected trace_id %s, got %v", traceID, logEntry["trace_id"])
	}
	if logEntry["span_id"] != spanID {
		t.Errorf("Expected span_id %s, got %v", spanID, logEntry["span_id"])
	}
}
