package protocol

import "context"

// EvolutionEngine converts intent into executable code.
type EvolutionEngine interface {
	// Evolve generates a WASM binary from the given blueprint.
	Evolve(ctx context.Context, blueprint Blueprint) ([]byte, error)
}
