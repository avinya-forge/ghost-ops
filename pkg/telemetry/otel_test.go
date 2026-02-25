package telemetry

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
)

func TestInitTracer(t *testing.T) {
	buf := &bytes.Buffer{}
	shutdown, err := InitTracer("test-service", buf)
	require.NoError(t, err)

	tracer := otel.Tracer("test-tracer")
	_, span := tracer.Start(context.Background(), "test-span")
	time.Sleep(10 * time.Millisecond)
	span.End()

	// Ensure shutdown flushes
	err = shutdown(context.Background())
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "test-span")
}
