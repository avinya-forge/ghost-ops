package logging

import (
	"context"
	"io"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

var logLevel = new(slog.LevelVar)

// TraceHandler wraps a slog.Handler to inject trace_id and span_id.
type TraceHandler struct {
	handler slog.Handler
}

// Enabled implements slog.Handler.
func (h *TraceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle implements slog.Handler.
func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if spanCtx.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		)
	}
	return h.handler.Handle(ctx, r)
}

// WithAttrs implements slog.Handler.
func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceHandler{handler: h.handler.WithAttrs(attrs)}
}

// WithGroup implements slog.Handler.
func (h *TraceHandler) WithGroup(name string) slog.Handler {
	return &TraceHandler{handler: h.handler.WithGroup(name)}
}

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
	traceHandler := &TraceHandler{handler: jsonHandler}
	logger := slog.New(traceHandler)
	slog.SetDefault(logger)
}

// SetLevel updates the logging level dynamically.
func SetLevel(level slog.Level) {
	logLevel.Set(level)
}

// Audit records an immutable state change with an "event_type": "audit" field.
func Audit(ctx context.Context, action string, details map[string]interface{}) {
	slog.InfoContext(ctx, action,
		slog.String("event_type", "audit"),
		slog.Any("details", details),
	)
}
