package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ghost-ops/pkg/api"
	"ghost-ops/pkg/evolution"
	"ghost-ops/pkg/intent"
	"ghost-ops/pkg/registry"
	"ghost-ops/pkg/runtime"
	"ghost-ops/pkg/store"
)

func main() {
	blueprintsPath := flag.String("blueprints", "blueprints.json", "Path to blueprints file")
	wasmPath := flag.String("wasm", "", "Path to mock WASM binary (optional)")
	storePath := flag.String("store", "ghost-ops-state.json", "Path to state store file")
	httpAddr := flag.String("http", "", "HTTP server address (e.g. :8080). If empty, server is not started.")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// Initialize State Store
	stateStore, err := store.NewJSONFileStore(*storePath)
	if err != nil {
		log.Fatalf("Failed to initialize state store: %v", err)
	}

	// Initialize Service Registry
	reg := registry.NewServiceRegistry(source, engine, host, stateStore)

	// Initial Reconciliation
	log.Println("Running initial reconciliation...")
	count := 0
	for {
		processed, err := reg.Reconcile(ctx)
		if err != nil {
			log.Printf("Error during reconciliation: %v", err)
			continue
		}
		if !processed {
			break
		}
		count++
	}
	log.Printf("Initial reconciliation complete. Processed %d blueprints.", count)

	// Start API Server if requested
	if *httpAddr != "" {
		server := api.NewAPIServer(reg)

		// Run server in goroutine
		go func() {
			log.Printf("Starting HTTP server on %s", *httpAddr)
			if err := server.Serve(*httpAddr); err != nil {
				log.Fatalf("HTTP server failed: %v", err)
			}
		}()

		// Wait for shutdown signal
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Println("Shutting down...")
	} else {
		log.Println("No HTTP server requested. Exiting.")
	}
}
