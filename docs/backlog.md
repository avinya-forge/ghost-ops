# Backlog

## Todo
- [ ] Set up basic CI/CD pipeline (GitHub Actions) <!-- id: 9 -->
- [ ] Create a "Hello World" WASM module to test runtime <!-- id: 10 -->

## In-Progress
- [ ] Set up basic CI/CD pipeline (GitHub Actions) <!-- id: 9 --> - [Jules-01]

## Done
- [x] Initialize Go module `go mod init ghost-ops` <!-- id: 1 -->
- [x] Create project structure: `cmd/`, `internal/`, `pkg/`, `generated/` <!-- id: 2 -->
- [x] Define `IntentSource` interface in `pkg/protocol/intent.go` <!-- id: 3 -->
- [x] Define `EvolutionEngine` interface in `pkg/protocol/evolution.go` <!-- id: 4 -->
- [x] Define `StateStore` interface in `pkg/protocol/store.go` <!-- id: 5 -->
- [x] Define `RuntimeHost` interface in `pkg/protocol/runtime.go` <!-- id: 6 -->
- [x] Implement `MockIntentSource` for testing <!-- id: 7 -->
- [x] Implement `InMemoryStateStore` for testing <!-- id: 8 -->
