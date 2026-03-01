package logging

import (
	"context"
	"log/slog"
)

// Audit logs an immutable state change or security event.
func Audit(ctx context.Context, action string, details map[string]interface{}) {
	args := []any{
		slog.String("event_type", "audit"),
		slog.String("action", action),
	}
	for k, v := range details {
		args = append(args, slog.Any(k, v))
	}
	slog.InfoContext(ctx, "Audit Event", args...)
}
