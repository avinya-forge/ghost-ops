# Backlog

## Todo

## In-Progress

## Done
- [x] Create Dockerfile for multi-stage build <!-- id: 11 -->
- [x] Implement basic CLI structure in `cmd/ghost-ops/main.go` <!-- id: 12 -->
- [x] Implement `FileIntentSource` (reading blueprints from file) <!-- id: 13 -->
- [x] Implement `MockEvolutionEngine` (returning pre-compiled or dummy WASM) <!-- id: 14 -->
- [x] Implement `WazeroRuntimeHost` (using wazero for WASM execution) <!-- id: 15 -->
- [x] Initialize Go module `go mod init ghost-ops` <!-- id: 1 -->
- [x] Create project structure: `cmd/`, `internal/`, `pkg/`, `generated/` <!-- id: 2 -->
- [x] Define `IntentSource` interface in `pkg/protocol/intent.go` <!-- id: 3 -->
- [x] Define `EvolutionEngine` interface in `pkg/protocol/evolution.go` <!-- id: 4 -->
- [x] Define `StateStore` interface in `pkg/protocol/store.go` <!-- id: 5 -->
- [x] Define `RuntimeHost` interface in `pkg/protocol/runtime.go` <!-- id: 6 -->
- [x] Implement `MockIntentSource` for testing <!-- id: 7 -->
- [x] Implement `InMemoryStateStore` for testing <!-- id: 8 -->
- [x] Set up basic CI/CD pipeline (GitHub Actions) <!-- id: 9 -->
- [x] Create a "Hello World" WASM module to test runtime <!-- id: 10 -->
