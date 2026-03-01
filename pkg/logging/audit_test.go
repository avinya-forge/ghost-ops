package logging

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestAudit(t *testing.T) {
	var buf bytes.Buffer
	InitLoggerWithWriter(slog.LevelInfo, &buf)

	ctx := context.Background()
	Audit(ctx, "update_service", map[string]interface{}{
		"service_id": "test-svc",
		"version":    2,
	})

	output := buf.String()
	if !strings.Contains(output, `"event_type":"audit"`) {
		t.Errorf("Expected event_type: audit, got: %s", output)
	}
	if !strings.Contains(output, `"action":"update_service"`) {
		t.Errorf("Expected action: update_service, got: %s", output)
	}
	if !strings.Contains(output, `"service_id":"test-svc"`) {
		t.Errorf("Expected service_id: test-svc, got: %s", output)
	}
	if !strings.Contains(output, `"version":2`) {
		t.Errorf("Expected version: 2, got: %s", output)
	}
}
