package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

func TestTraceHandler_Handle(t *testing.T) {
	// Setup Tracer
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	tracer := tp.Tracer("test-tracer")

	// Setup Logger
	var buf bytes.Buffer
	InitLoggerWithWriter(slog.LevelInfo, &buf)

	// Start Span
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	// Log
	slog.InfoContext(ctx, "test log message")

	// Verify Output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to unmarshal log entry: %v", err)
	}

	if msg, ok := logEntry["msg"].(string); !ok || msg != "test log message" {
		t.Errorf("Expected message 'test log message', got %v", msg)
	}

	if traceID, ok := logEntry["trace_id"].(string); !ok || traceID == "" {
		t.Error("Expected trace_id in log entry")
	}

	if spanID, ok := logEntry["span_id"].(string); !ok || spanID == "" {
		t.Error("Expected span_id in log entry")
	}
}
