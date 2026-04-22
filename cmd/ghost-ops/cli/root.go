package cli

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"ghost-ops/pkg/api"
	"ghost-ops/pkg/config"
	"ghost-ops/pkg/event"
	"ghost-ops/pkg/evolution"
	"ghost-ops/pkg/intent"
	"ghost-ops/pkg/llm"
	"ghost-ops/pkg/logging"
	"ghost-ops/pkg/observer"
	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/registry"
	"ghost-ops/pkg/runtime"
	"ghost-ops/pkg/store"
	"ghost-ops/pkg/telemetry"
)

// Execute handles the main CLI entry point.
func Execute(version string) {
	// Define flags using pflag
	pflag.String("blueprints", "blueprints.json", "Path to blueprints file")
	pflag.String("wasm", "", "Path to mock WASM binary (optional)")
	pflag.String("store", "store.json", "Path to state store file")
	pflag.Int("port", 8080, "HTTP server port")
	pflag.String("engine", "mock", "Evolution engine to use (mock, compiler, ai)")
	pflag.String("llm", "mock", "LLM provider to use (mock)")
	pflag.Bool("debug", false, "Enable debug logging")

	pflag.Parse()

	// Bind flags to Viper
	viper.BindPFlag("paths.blueprints", pflag.Lookup("blueprints"))
	viper.BindPFlag("paths.wasm", pflag.Lookup("wasm"))
	viper.BindPFlag("paths.store", pflag.Lookup("store"))
	viper.BindPFlag("server.port", pflag.Lookup("port"))
	viper.BindPFlag("engine.type", pflag.Lookup("engine"))
	viper.BindPFlag("llm.provider", pflag.Lookup("llm"))

	// Load configuration early to support CLI commands. Non-server commands
	// (e.g. version) don't require a valid config file, so we do not exit on
	// error — but we do warn so that a corrupted config is immediately visible.
	if _, err := config.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load config: %v\n", err)
	}

	args := pflag.Args()
	if len(args) > 0 {
		switch args[0] {
		case "version":
			fmt.Printf("ghost-ops version %s\n", version)
			return
		case "service":
			if len(args) > 1 {
				handleServiceCommand(args[1:])
				return
			}
			fmt.Println("Usage: ghost-ops service [list|inspect <id>]")
			os.Exit(1)
		case "config":
			if len(args) > 1 {
				handleConfigCommand(args[1:])
				return
			}
			fmt.Println("Usage: ghost-ops config [show]")
			os.Exit(1)
		case "init":
			runInit()
			return
		}
	}

	runServer(version)
}

func runInit() {
	configFile := "ghost-ops.yaml"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		configContent := `# GhostOps Configuration
paths:
  blueprints: "blueprints/blueprints.json"
  store: "store.json"
  wasm: ""

server:
  port: 8080

engine:
  type: "mock" # Options: mock, compiler, ai

llm:
  provider: "mock" # Options: mock, openai, ollama
  api_key: ""
  model: "gpt-4"
  base_url: ""
  system_prompt: ""

logging:
  level: "info" # Options: debug, info, warn, error
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			fmt.Printf("Failed to create config file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Created ghost-ops.yaml")
	} else {
		fmt.Println("ghost-ops.yaml already exists")
	}

	if _, err := os.Stat("blueprints"); os.IsNotExist(err) {
		if err := os.Mkdir("blueprints", 0755); err != nil {
			fmt.Printf("Failed to create blueprints directory: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Created blueprints directory")
	}

	blueprintFile := "blueprints/blueprints.json"
	if _, err := os.Stat(blueprintFile); os.IsNotExist(err) {
		sampleContent := `[
  {
    "service_id": "hello-service",
    "intent": "Create a service that responds 'Hello from GhostOps' to any request.",
    "constraints": {
      "language": "go"
    }
  }
]`
		if err := os.WriteFile(blueprintFile, []byte(sampleContent), 0644); err != nil {
			fmt.Printf("Failed to create blueprints file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Created blueprints/blueprints.json")
	}

	fmt.Println("Initialization complete. Run './ghost-ops' to start.")
}

func runServer(version string) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Handle debug flag override
	debug, _ := pflag.CommandLine.GetBool("debug")
	if debug {
		cfg.Logging.Level = "debug"
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logging
	logLevel := slog.LevelInfo
	if cfg.Logging.Level == "debug" {
		logLevel = slog.LevelDebug
	}
	logging.InitLogger(logLevel)

	// Initialize Tracer
	shutdownTracer, err := telemetry.InitTracer(context.Background(), "ghost-ops", cfg.Telemetry, os.Stdout)
	if err != nil {
		slog.Error("Failed to initialize tracer", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			slog.Error("Failed to shutdown tracer", "error", err)
		}
	}()

	// Watch for config changes
	config.Watch(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slog.Info("GhostOps starting...", "version", version)

	// Initialize Intent Source
	source, err := intent.NewFileIntentSource(cfg.Paths.Blueprints)
	if err != nil {
		slog.Error("Failed to initialize intent source", "path", cfg.Paths.Blueprints, "error", err)
		os.Exit(1)
	}

	// Initialize Telemetry Collector
	collector := telemetry.NewInMemoryCollector()

	// Initialize Evolution Engine
	var engine protocol.EvolutionEngine
	switch cfg.Engine.Type {
	case "compiler":
		engine = evolution.NewGoCompilerEngine()
		slog.Info("Using Go Compiler Evolution Engine")
	case "mock":
		engine = evolution.NewMockEvolutionEngine(cfg.Paths.WASM)
		slog.Info("Using Mock Evolution Engine", "wasm_path", cfg.Paths.WASM)
	case "ai":
		var provider protocol.LLMProvider
		switch cfg.LLM.Provider {
		case "mock":
			provider = &llm.MockLLMProvider{}
			slog.Info("Using Mock LLM Provider")
		case "openai":
			apiKey := cfg.LLM.APIKey
			if apiKey == "" {
				apiKey = os.Getenv("OPENAI_API_KEY")
			}
			if apiKey == "" {
				slog.Error("OPENAI_API_KEY environment variable or llm.api_key config is required for openai provider")
				os.Exit(1)
			}
			model := cfg.LLM.Model
			if model == "" {
				model = os.Getenv("OPENAI_MODEL")
			}
			provider = llm.NewOpenAIProvider(apiKey, model, cfg.LLM.BaseURL, cfg.LLM.SystemPrompt)
			slog.Info("Using OpenAI LLM Provider", "model", model)
		case "ollama":
			model := cfg.LLM.Model
			if model == "" {
				model = os.Getenv("OLLAMA_MODEL")
			}
			provider = llm.NewOllamaProvider(cfg.LLM.BaseURL, model, cfg.LLM.SystemPrompt)
			slog.Info("Using Ollama LLM Provider", "model", model)
		default:
			slog.Error("Invalid LLM provider", "provider", cfg.LLM.Provider)
			os.Exit(1)
		}

		if cfg.LLM.CacheEnabled {
			cache := llm.NewInMemoryCache()
			provider = llm.NewCachedLLMProvider(provider, cache)
			slog.Info("LLM Prompt Caching Enabled")
		}

		engine = evolution.NewAIEvolutionEngine(provider, collector)
		slog.Info("Using AI Evolution Engine")
	default:
		slog.Error("Invalid engine type", "type", cfg.Engine.Type)
		os.Exit(1)
	}

	// Initialize State Store
	stateStore, err := store.NewJSONFileStore(cfg.Paths.Store)
	if err != nil {
		slog.Error("Failed to initialize state store", "path", cfg.Paths.Store, "error", err)
		os.Exit(1)
	}

	// Initialize Runtime Host
	host, err := runtime.NewWazeroRuntimeHost(ctx, stateStore, collector, cfg.Capabilities)
	if err != nil {
		slog.Error("Failed to initialize runtime", "error", err)
		os.Exit(1)
	}
	defer host.Close(ctx)

	// Initialize Event Bus
	eventBus := event.NewInMemoryEventBus()

	// Initialize Registry
	reg := registry.NewRegistry(stateStore, engine, source, host, collector, eventBus, cfg.Registry)

	// Start Health Check Loop
	reg.StartHealthCheck(ctx)

	// Start internal event loop (Promotions/Rollbacks)
	reg.StartEventLoop(ctx)

	// Initialize Observers
	metricObserver := observer.NewMetricObserver(collector, eventBus)
	metricObserver.Start(ctx, 5*time.Second)

	// Initialize API Server
	srv := api.NewServer(reg, collector, cfg.Server.MaxBodySize)
	httpAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	httpServer := &http.Server{
		Addr:              httpAddr,
		Handler:           srv,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start API Server
	go func() {
		slog.Info("Starting HTTP server", "addr", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Continuous Reconciliation Loop
	go func() {
		slog.Info("Starting reconciliation loop")
		for {
			select {
			case <-ctx.Done():
				return
			default:
				processed, err := reg.Reconcile(ctx)
				if err != nil {
					slog.Error("Reconciliation error", "error", err)
					time.Sleep(5 * time.Second) // Backoff on error
					continue
				}
				if !processed {
					// No pending blueprints, sleep before checking again
					time.Sleep(2 * time.Second)
				}
			}
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.Info("Shutting down...")
	cancel() // Signal loop to stop

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server shutdown failed", "error", err)
	}

	slog.Info("GhostOps stopped")
}
