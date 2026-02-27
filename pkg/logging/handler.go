package logging

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// TraceHandler wraps a slog.Handler to inject trace IDs.
type TraceHandler struct {
	handler slog.Handler
}

// NewTraceHandler creates a new TraceHandler.
func NewTraceHandler(h slog.Handler) slog.Handler {
	return &TraceHandler{handler: h}
}

// Enabled reports whether the handler handles records at the given level.
func (h *TraceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle handles the Record.
func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if spanCtx.IsValid() {
		// Create a new record with the trace attributes
		r.AddAttrs(
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		)
	}
	return h.handler.Handle(ctx, r)
}

// WithAttrs returns a new Handler with the given attributes added.
func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewTraceHandler(h.handler.WithAttrs(attrs))
}

// WithGroup returns a new Handler with the given group name.
func (h *TraceHandler) WithGroup(name string) slog.Handler {
	return NewTraceHandler(h.handler.WithGroup(name))
}
