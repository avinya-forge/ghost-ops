package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ghost-ops/pkg/api"
	"ghost-ops/pkg/evolution"
	"ghost-ops/pkg/intent"
	"ghost-ops/pkg/llm"
	"ghost-ops/pkg/logging"
	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/registry"
	"ghost-ops/pkg/runtime"
	"ghost-ops/pkg/store"
)

func main() {
	blueprintsPath := flag.String("blueprints", "blueprints.json", "Path to blueprints file")
	wasmPath := flag.String("wasm", "", "Path to mock WASM binary (optional)")
	storePath := flag.String("store", "store.json", "Path to state store file")
	httpAddr := flag.String("http", ":8080", "HTTP server address")
	engineType := flag.String("engine", "mock", "Evolution engine to use (mock, compiler, ai)")
	llmProvider := flag.String("llm", "mock", "LLM provider to use (mock)")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Initialize structured logging
	logLevel := slog.LevelInfo
	if *debug {
		logLevel = slog.LevelDebug
	}
	logging.InitLogger(logLevel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slog.Info("GhostOps starting...")

	// Initialize Intent Source
	source, err := intent.NewFileIntentSource(*blueprintsPath)
	if err != nil {
		slog.Error("Failed to initialize intent source", "path", *blueprintsPath, "error", err)
		os.Exit(1)
	}

	// Initialize Evolution Engine
	var engine protocol.EvolutionEngine
	switch *engineType {
	case "compiler":
		engine = evolution.NewGoCompilerEngine()
		slog.Info("Using Go Compiler Evolution Engine")
	case "mock":
		engine = evolution.NewMockEvolutionEngine(*wasmPath)
		slog.Info("Using Mock Evolution Engine", "wasm_path", *wasmPath)
	case "ai":
		var provider protocol.LLMProvider
		switch *llmProvider {
		case "mock":
			provider = &llm.MockLLMProvider{}
			slog.Info("Using Mock LLM Provider")
		default:
			slog.Error("Invalid LLM provider", "provider", *llmProvider)
			os.Exit(1)
		}
		engine = evolution.NewAIEvolutionEngine(provider)
		slog.Info("Using AI Evolution Engine")
	default:
		slog.Error("Invalid engine type", "type", *engineType)
		os.Exit(1)
	}

	// Initialize State Store
	stateStore, err := store.NewJSONFileStore(*storePath)
	if err != nil {
		slog.Error("Failed to initialize state store", "path", *storePath, "error", err)
		os.Exit(1)
	}

	// Initialize Runtime Host
	host, err := runtime.NewWazeroRuntimeHost(ctx, stateStore)
	if err != nil {
		slog.Error("Failed to initialize runtime", "error", err)
		os.Exit(1)
	}
	defer host.Close(ctx)

	// Initialize Registry
	reg := registry.NewRegistry(stateStore, engine, source, host)

	// Initialize API Server
	srv := api.NewServer(reg)
	httpServer := &http.Server{
		Addr:    *httpAddr,
		Handler: srv,
	}

	// Start API Server
	go func() {
		slog.Info("Starting HTTP server", "addr", *httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Initial Reconciliation
	for {
		processed, err := reg.Reconcile(ctx)
		if err != nil {
			slog.Error("Initial reconciliation failed", "error", err)
			// Decide whether to exit or continue. For now, we continue but log error.
			// Maybe break if critical?
			break
		}
		if !processed {
			break
		}
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.Info("Shutting down...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server shutdown failed", "error", err)
	}

	slog.Info("GhostOps stopped")
}
