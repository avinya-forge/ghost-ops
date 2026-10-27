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
		"api_key":    "super-secret-123",
		"my_token":   "tok_abc123",
		"PASSWORD":   "hunter2",
		"is_active":  true,
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
	if !strings.Contains(output, `"is_active":true`) {
		t.Errorf("Expected is_active: true, got: %s", output)
	}

	// Verify masking
	if strings.Contains(output, "super-secret-123") {
		t.Errorf("Expected api_key to be masked, got: %s", output)
	}
	if !strings.Contains(output, `"api_key":"***MASKED***"`) {
		t.Errorf("Expected api_key: ***MASKED***, got: %s", output)
	}

	if strings.Contains(output, "tok_abc123") {
		t.Errorf("Expected my_token to be masked, got: %s", output)
	}
	if !strings.Contains(output, `"my_token":"***MASKED***"`) {
		t.Errorf("Expected my_token: ***MASKED***, got: %s", output)
	}

	if strings.Contains(output, "hunter2") {
		t.Errorf("Expected PASSWORD to be masked, got: %s", output)
	}
	if !strings.Contains(output, `"PASSWORD":"***MASKED***"`) {
		t.Errorf("Expected PASSWORD: ***MASKED***, got: %s", output)
	}
}
