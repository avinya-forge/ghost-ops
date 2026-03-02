package telemetry

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"

	"ghost-ops/pkg/config"
)

func TestInitTracer_Stdout(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer

	cfg := config.TelemetryConfig{
		Exporter: "stdout",
	}

	shutdown, err := InitTracer(ctx, "test-service", cfg, &buf)
	assert.NoError(t, err)
	assert.NotNil(t, shutdown)

	// Verify tracer provider is set
	tp := otel.GetTracerProvider()
	assert.NotNil(t, tp)

	// Clean up
	err = shutdown(ctx)
	assert.NoError(t, err)
}

func TestInitTracer_OtlpHttp(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer

	cfg := config.TelemetryConfig{
		Exporter: "otlp-http",
		Endpoint: "localhost:4318",
	}

	shutdown, err := InitTracer(ctx, "test-service", cfg, &buf)
	assert.NoError(t, err)
	assert.NotNil(t, shutdown)

	// Clean up
	err = shutdown(ctx)
	assert.NoError(t, err)
}

func TestInitTracer_OtlpGrpc(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer

	cfg := config.TelemetryConfig{
		Exporter: "otlp-grpc",
		Endpoint: "localhost:4317",
	}

	shutdown, err := InitTracer(ctx, "test-service", cfg, &buf)
	assert.NoError(t, err)
	assert.NotNil(t, shutdown)

	// Clean up
	err = shutdown(ctx)
	assert.NoError(t, err)
}

func TestInitTracer_Default(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer

	// empty exporter should default to stdout
	cfg := config.TelemetryConfig{
		Exporter: "",
	}

	shutdown, err := InitTracer(ctx, "test-service", cfg, &buf)
	assert.NoError(t, err)
	assert.NotNil(t, shutdown)

	// Clean up
	err = shutdown(ctx)
	assert.NoError(t, err)
}
