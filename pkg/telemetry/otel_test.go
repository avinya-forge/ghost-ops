package telemetry

import (
	"context"
	"io"
	"testing"

	"ghost-ops/pkg/config"

	"github.com/stretchr/testify/assert"
)

func TestInitTracer_Stdout(t *testing.T) {
	cfg := config.TelemetryConfig{
		Exporter: "stdout",
	}
	shutdown, err := InitTracer(context.Background(), "test-service", cfg, io.Discard)
	assert.NoError(t, err)
	assert.NotNil(t, shutdown)
	_ = shutdown(context.Background())
}

func TestInitTracer_OTLP_HTTP(t *testing.T) {
	// Just check if it initializes without error (it might warn about connection if we tried to send spans, but init should be fine)
	cfg := config.TelemetryConfig{
		Exporter: "otlp-http",
		Endpoint: "localhost:4318",
	}
	shutdown, err := InitTracer(context.Background(), "test-service", cfg, io.Discard)
	assert.NoError(t, err)
	assert.NotNil(t, shutdown)
	_ = shutdown(context.Background())
}
