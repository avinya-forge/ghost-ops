# Release Notes

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
