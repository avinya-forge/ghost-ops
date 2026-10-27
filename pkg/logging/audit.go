package logging

import (
	"context"
	"log/slog"
	"strings"
)

// Audit logs an immutable state change or security event.
func Audit(ctx context.Context, action string, details map[string]interface{}) {
	args := []any{
		slog.String("event_type", "audit"),
		slog.String("action", action),
	}
	for k, v := range details {
		lowerKey := strings.ToLower(k)
		if strings.Contains(lowerKey, "secret") || strings.Contains(lowerKey, "token") || strings.Contains(lowerKey, "password") || strings.Contains(lowerKey, "key") {
			args = append(args, slog.String(k, "***MASKED***"))
		} else {
			args = append(args, slog.Any(k, v))
		}
	}
	slog.InfoContext(ctx, "Audit Event", args...)
}
