# Backlog

## Phase 0: Hygiene (Current Focus)
- [x] Create `README.md` with project overview and architecture diagram <!-- id: 40 -->
- [x] Create `Makefile` with targets: `build`, `test`, `lint`, `run` <!-- id: 41 -->
- [x] Add `/healthz` endpoint to API Server <!-- id: 42 -->
- [ ] Add linter config (`golangci-lint`) <!-- id: 43 -->
- [ ] Refactor `examples/` directory structure for clarity <!-- id: 44 -->
- [ ] Add unit tests for `Registry` error cases <!-- id: 45 -->
- [ ] Add unit tests for `WazeroRuntimeHost` edge cases <!-- id: 46 -->
- [ ] Document `docs/development-standards.md` with PR guidelines <!-- id: 47 -->
- [ ] Audit and ensure structured logging in all components <!-- id: 48 -->
- [ ] Ensure graceful shutdown in `main.go` (context cancellation) <!-- id: 49 -->

## Phase 1: MVP (The Self-Healing Loop)
- [ ] Implement Configuration Management (replace flags with Viper/koanf) <!-- id: 50 -->
- [ ] Implement better error handling (custom error types) <!-- id: 51 -->
- [ ] Implement `BlueprintSource` interface for HTTP/S3 sources <!-- id: 52 -->
- [ ] Add integration tests for full end-to-end flow <!-- id: 53 -->
- [ ] Implement CLI flag for runtime `log-level` adjustment <!-- id: 54 -->
- [ ] Implement authentication for `/metrics` endpoint <!-- id: 55 -->
- [ ] Add `version` command to CLI <!-- id: 56 -->
- [ ] Implement basic TUI Dashboard for Service Registry <!-- id: 57 -->
- [ ] Add support for complex constraints in Blueprint schema <!-- id: 58 -->
- [ ] Implement `Blueprint` validation logic before processing <!-- id: 59 -->
- [ ] Verify and Document Shadow Mode implementation <!-- id: 80 -->
- [ ] Verify and Document Zero-Downtime Deployment implementation <!-- id: 81 -->

## Phase 2: Scale (Distributed & Resilient)
- [ ] Design Distributed Store Interface <!-- id: 60 -->
- [ ] Implement Redis Store Adapter <!-- id: 61 -->
- [ ] Implement Etcd Store Adapter <!-- id: 62 -->
- [ ] Implement Remote WASM Registry Client (OCI compliant) <!-- id: 63 -->
- [ ] Implement Leader Election for Registry high availability <!-- id: 64 -->
- [ ] Implement Horizontal Pod Autoscaling (HPA) metrics exposure <!-- id: 65 -->
- [ ] Add Distributed Tracing (OpenTelemetry) <!-- id: 66 -->
- [ ] Implement Circuit Breaker for Evolution Engine (prevent loops) <!-- id: 67 -->
- [ ] Implement Rate Limiting for API endpoints <!-- id: 68 -->
- [ ] Implement Multi-Tenant support (Namespaces/Tenants) <!-- id: 69 -->

## Phase 3: Future (Autonomous Evolution)
- [ ] Design Autonomous Feedback Loop Architecture <!-- id: 70 -->
- [ ] Implement Runtime Metric Analysis Engine <!-- id: 71 -->
- [ ] Implement Python Guest SDK <!-- id: 72 -->
- [ ] Implement Rust Guest SDK <!-- id: 73 -->
- [ ] Implement Vulnerability Scanner for generated WASM code <!-- id: 74 -->
- [ ] Implement "Dark Mode" autonomous operation toggle <!-- id: 75 -->
