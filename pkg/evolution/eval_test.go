package evolution

import (
	"context"
	"testing"

	"ghost-ops/pkg/llm"
	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestCodeGenerationQuality_E2E(t *testing.T) {
	// 1. Valid Code Case
	validCode := `package main
import "fmt"
func main() { fmt.Println("Hello") }
`
	provider := &llm.MockLLMProvider{
		Response: validCode,
	}

	collector := telemetry.NewInMemoryCollector()
	engine := NewAIEvolutionEngine(provider, collector)

	bp := protocol.Blueprint{
		ServiceID: "eval-service",
		Intent:    "hello world",
	}

	wasm, err := engine.Evolve(context.Background(), bp)
	assert.NoError(t, err)
	assert.NotEmpty(t, wasm)

	// 2. Invalid Code Case (Compilation Error)
	provider.Response = `package main; func main() { INVALID_SYNTAX }`
	_, err = engine.Evolve(context.Background(), bp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "compilation failed")
}
