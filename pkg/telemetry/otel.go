package telemetry

import (
	"context"
	"fmt"
	"io"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"ghost-ops/pkg/config"
)

// InitTracer initializes the OpenTelemetry tracer provider.
// It returns a shutdown function that should be called on application exit.
func InitTracer(ctx context.Context, serviceName string, cfg config.TelemetryConfig, w io.Writer) (func(context.Context) error, error) {
	var exporter sdktrace.SpanExporter
	var err error

	switch cfg.Exporter {
	case "otlp-http":
		opts := []otlptracehttp.Option{}
		if cfg.Endpoint != "" {
			opts = append(opts, otlptracehttp.WithEndpoint(cfg.Endpoint))
		}
		// For simplicity, we assume insecure for now if no scheme is provided or we can add Insecure option based on config
		// But usually Endpoint includes scheme or we just follow defaults.
		// otlptracehttp defaults to https://localhost:4318
		// If endpoint doesn't have scheme, we might need WithInsecure() if testing locally.
		// Let's assume standard behavior for now.
		exporter, err = otlptracehttp.New(ctx, opts...)
	case "otlp-grpc":
		opts := []otlptracegrpc.Option{}
		if cfg.Endpoint != "" {
			opts = append(opts, otlptracegrpc.WithEndpoint(cfg.Endpoint))
		}
		exporter, err = otlptracegrpc.New(ctx, opts...)
	default:
		// Default to stdout
		exporter, err = stdouttrace.New(
			stdouttrace.WithWriter(w),
			stdouttrace.WithPrettyPrint(),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create exporter %s: %w", cfg.Exporter, err)
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown, nil
}
