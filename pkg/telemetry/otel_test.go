package telemetry

import (
	"context"
	"io"
	"testing"

	"ghost-ops/pkg/config"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

func TestInitTracer_Stdout(t *testing.T) {
	cfg := config.TelemetryConfig{
		Exporter: "stdout",
	}
	shutdown, err := InitTracer(context.Background(), "test-service", cfg, io.Discard)
	assert.NoError(t, err)
	defer shutdown(context.Background())

	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	span.End()

	assert.NotNil(t, otel.GetTracerProvider())
}

func TestInitTracer_OTLP_HTTP(t *testing.T) {
	// We can't easily test the actual connection without a mock server,
	// but we can verify it doesn't error on initialization with a dummy endpoint.
	cfg := config.TelemetryConfig{
		Exporter: "otlp-http",
		Endpoint: "localhost:4318",
	}
	// OTLP exporters might try to connect immediately or in background.
	// New returns an error if it fails to create the exporter.
	shutdown, err := InitTracer(context.Background(), "test-service", cfg, io.Discard)
	assert.NoError(t, err)
	if shutdown != nil {
		defer shutdown(context.Background())
	}
}

func TestInitTracer_OTLP_GRPC(t *testing.T) {
	cfg := config.TelemetryConfig{
		Exporter: "otlp-grpc",
		Endpoint: "localhost:4317",
	}
	shutdown, err := InitTracer(context.Background(), "test-service", cfg, io.Discard)
	assert.NoError(t, err)
	if shutdown != nil {
		defer shutdown(context.Background())
	}
}
