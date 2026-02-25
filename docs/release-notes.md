# Release Notes

## Unreleased
- **Distributed Store**: Implemented Redis Pub/Sub for event broadcasting (`Publish`/`Subscribe`).
- **Distributed Store**: Implemented GZIP compression for service records in Redis to optimize storage.
- **Observability**: Implemented trace correlation in logs by injecting `trace_id` and `span_id` via a custom `TraceHandler`.
- **Resilience**: Implemented Service Mesh Lite patterns: Retry (Exponential Backoff), Circuit Breaker, and Rate Limiting (Token Bucket).
- **Distributed Store**: Added Distributed Lock mechanism to `RedisStore` using `SET NX` for coordination.
- **Observability**: Integrated OpenTelemetry SDK for tracing and added Tracing Middleware to API.
- **Registry**: Implemented an in-memory Event Bus to publish `ServiceDeployed` and `ServiceUnhealthy` events.
- **API**: Added validation middleware (`MaxBytes`, `ContentType`, `Recover`) to harden API endpoints.
- **Developer Experience**: Added `ghost-dev` tool for hot-reloading development server (`make dev`).
- **Developer Experience**: Added pre-commit hooks (`make hooks`) to enforce lint and test pass before commit.
- **Testing**: Added end-to-end integration tests (`pkg/api/integration_test.go`) covering intent-to-invocation flow.
- **Runtime**: Implemented async module instantiation to prevent blocking on long-running guest initialization.
- **Runtime**: Implemented compiled module cache with LRU eviction (limit 50) to optimize reload performance.
- **Registry**: Enhanced service health check logic to verify active module state and context.
- **API**: Added `POST /services/{id}/invoke` endpoint for direct service invocation.
- **CLI Enhancements**: Added `ghost-ops init` to bootstrap projects and `ghost-ops service logs` to view service output.
- **Runtime Hardening**: Configured Wazero with CPU limits (CloseOnContextDone) and implemented secure log capture for guest modules.
- **Error Handling**: Added stack traces to internal errors (logged only) for better debugging.
- **Security**: Ran `gosec` and fixed high-severity issues (permissions, integer overflows, slowloris).
- **Error Handling**: Refactored `Registry`, `Runtime`, and `API` to use standardized `AppError` types and HTTP status codes.
- **Testing**: Added integration tests for API error mapping.
- **Testing**: Added unit tests for `pkg/logging` (refactored for testability) and `pkg/sdk/guest` (including RPC integration tests).
- **Observability**: Enhanced `Registry` with duration metrics for reconcile, evolve, and deploy phases.
- **CLI Enhancements**: Added `ghost-ops service list`, `ghost-ops service inspect <id>`, and `ghost-ops config show` commands.
- **Configuration Management**: Migrated CLI flags to Viper, implemented config validation, and added support for hot-reloading configuration (including dynamic logging level).
- **Validation**: Implemented strict validation for Blueprints, including `service_id` format, intent length, and constraint size limits.
- **Refactoring**: Refactored CLI structure to support subcommands using `cobra`-like pattern with `pflag`.
- Added Viper dependency and basic configuration setup (`pkg/config`).
- Implemented environment variable overrides and config file support.
- Defined strongly-typed `Config` struct in `pkg/config` with JSON/YAML support.
- Implemented `AppError` interface and sentinel errors in `pkg/protocol`.
- Configured Wazero runtime with 128MB memory limit per instance.
- Implemented 5s execution timeout for WASM invocations.
- Added unit tests for `pkg/telemetry`.
- Implemented `ghost-ops version` command.
- **AI**: Added support for Ollama LLM provider (`llm.provider: ollama`).
- **AI**: Added support for custom system prompts in configuration (`llm.system_prompt`).
- **Optimization**: Benchmarked RuntimeHost performance (~0.08ms overhead), meeting <5ms target.
- **AI**: Implemented Prompt Caching for LLM providers (`llm.cache_enabled`) to reduce costs and latency for identical intents.
- **API**: Generated OpenAPI 3.0 Specification (`docs/openapi.yaml`) for API documentation.
- **Observability**: Added Token Usage Metrics (`llm_tokens_total`) to track input/output costs.
- **Registry**: Implemented Dependency Graph validation (`pkg/registry/graph.go`) to prevent cycles in service dependencies.

## v0.2.1 (Released)
- Restructured `examples/` directory to group basic examples (`examples/basic/hello-world`).
- Added `CONTRIBUTING.md` for development guidelines.

## v0.2.0 (Released)
- Implemented Telemetry Collection (runtime metrics) via `MetricsCollector`.
- Integrated AI Engine into CLI for end-to-end intent-to-WASM flow.
- Implemented `AIEvolutionEngine` to support generating WASM from natural language intent (using Mock LLM or OpenAI).
- Added `LLMProvider` interface and `MockLLMProvider`.
- Added `OpenAIProvider` for real code generation from intents.
- Refactored `GoCompilerEngine` to use shared compilation logic.
- Implemented Service Versioning Strategy (incrementing version on evolution).
- Enhanced Reconcile Loop to support continuous evolution by reloading blueprints when the file changes.
- Enhanced AI Evolution Engine to support constraint-aware synthesis by passing full blueprints (intents + constraints) to the LLM.
- Added `/healthz` endpoint to API Server for liveness probes.
- implemented graceful shutdown handling in `main.go` using context cancellation.
- Audited and ensured structured logging across all core components.

## v0.1.0 (Foundation)
- Initial project setup.
- Core interface definitions.
- Added GitHub Actions CI workflow.
- Added "Hello World" WASM module example.
- Implemented Structured Logging.
- Implemented persistent JSON File Store.
- Implemented Service Registry for lifecycle management.
- Implemented HTTP API Server for external interaction.
- Integrated all core components in the main CLI.
- Added support for payload passing and memory management in WASM host runtime.
- Implemented `GoCompilerEvolutionEngine` to support compiling Go source code from blueprints into WASM.
