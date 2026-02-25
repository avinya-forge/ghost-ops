package logging

import (
	"context"
	"io"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

var logLevel = new(slog.LevelVar)

// TraceHandler is a slog.Handler that adds trace IDs to the record.
type TraceHandler struct {
	slog.Handler
}

// Handle handles the record.
func (h TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		r.AddAttrs(
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs returns a new TraceHandler with the given attributes.
func (h TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return TraceHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup returns a new TraceHandler with the given group.
func (h TraceHandler) WithGroup(name string) slog.Handler {
	return TraceHandler{Handler: h.Handler.WithGroup(name)}
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
	handler := &TraceHandler{Handler: slog.NewJSONHandler(w, opts)}
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// SetLevel updates the logging level dynamically.
func SetLevel(level slog.Level) {
	logLevel.Set(level)
}
