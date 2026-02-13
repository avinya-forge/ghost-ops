package main

import (
	"context"
	"flag"
	"log"

	"ghost-ops/pkg/evolution"
	"ghost-ops/pkg/intent"
	"ghost-ops/pkg/runtime"
)

func main() {
	blueprintsPath := flag.String("blueprints", "blueprints.json", "Path to blueprints file")
	wasmPath := flag.String("wasm", "", "Path to mock WASM binary (optional)")
	flag.Parse()

	ctx := context.Background()

	log.Println("GhostOps starting...")

	// Initialize Intent Source
	source, err := intent.NewFileIntentSource(*blueprintsPath)
	if err != nil {
		log.Fatalf("Failed to initialize intent source from %s: %v", *blueprintsPath, err)
	}

	// Initialize Evolution Engine
	engine := evolution.NewMockEvolutionEngine(*wasmPath)

	// Initialize Runtime Host
	host, err := runtime.NewWazeroRuntimeHost(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize runtime: %v", err)
	}
	defer host.Close(ctx)

	// Processing loop
	count := 0
	for {
		bp, err := source.GetNextBlueprint(ctx)
		if err != nil {
			log.Printf("Error getting blueprint: %v", err)
			continue
		}
		if bp == nil {
			log.Println("No more blueprints.")
			break
		}

		log.Printf("Processing blueprint for service: %s", bp.ServiceID)

		// Evolve
		wasmBytes, err := engine.Evolve(ctx, *bp)
		if err != nil {
			log.Printf("Failed to evolve service %s: %v", bp.ServiceID, err)
			continue
		}

		// Deploy/Load
		if err := host.LoadModule(ctx, bp.ServiceID, wasmBytes); err != nil {
			log.Printf("Failed to load module for service %s: %v", bp.ServiceID, err)
			continue
		}

		log.Printf("Service %s loaded successfully.", bp.ServiceID)
		count++
	}

	log.Printf("GhostOps finished. Processed %d blueprints.", count)
}
