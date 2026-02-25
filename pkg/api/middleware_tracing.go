package api

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// TracingMiddleware wraps a handler to start a span for each request.
func TracingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracer := otel.Tracer("ghost-ops/api")

		// Extract propagation context
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		ctx, span := tracer.Start(ctx, spanName)
		defer span.End()

		// Serve with new context
		next(w, r.WithContext(ctx))
	}
}
