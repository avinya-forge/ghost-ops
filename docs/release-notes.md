# Release Notes

## Unreleased
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
