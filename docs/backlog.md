# Backlog

## Phase 0: Hygiene (Foundations)
- [x] [100] | [LINT] Add `.golangci.yml` configuration | [INDEPENDENT] | [DONE]
    - [LINT] Enable `staticcheck`, `gofmt`, `govet`.
- [x] [101] | [LINT] Fix `staticcheck` issues (SA4006 unused variables) | [BLOCKS-100] | [DONE]
    - [LINT] Run linter and fix reported issues.
- [x] [102] | [TEST] Add Unit Tests for `Registry` Error Paths | [INDEPENDENT] | [DONE]
    - [UNIT] Coverage for `Reconcile` failure scenarios.
- [x] [103] | [TEST] Add Unit Tests for `WazeroRuntimeHost` Edge Cases | [INDEPENDENT] | [DONE]
    - [UNIT] Coverage for `LoadModule` with invalid WASM.
- [x] [104] | [TEST] Add Unit Tests for `IntentSource` Edge Cases | [INDEPENDENT] | [DONE]
    - [UNIT] Coverage for empty file, invalid JSON.
- [x] [105] | [REFACTOR] Restructure `examples/` directory | [INDEPENDENT] | [DONE]
    - Move `hello-world` to `examples/basic/`.
- [x] [106] | [DOCS] Create `CONTRIBUTING.md` | [INDEPENDENT] | [DONE]
    - [DOC] detailed guide on how to contribute.

## Phase 1: MVP (The Self-Healing Loop)

### Epic: Configuration Management
- [ ] [200] | [FEAT] Add Viper Dependency & Setup | [INDEPENDENT] | [TODO]
    - [SEC] Ensure secure defaults.
- [ ] [201] | [FEAT] Define `Config` Struct | [BLOCKS-200] | [TODO]
    - Define fields for `Server`, `Log`, `Store`, `Runtime`.
- [ ] [202] | [FEAT] Migrate CLI Flags to Viper Config | [BLOCKS-201] | [TODO]
    - Replace `flag` package usage in `main.go`.
- [ ] [203] | [FEAT] Add Config Validation Logic | [BLOCKS-201] | [TODO]
    - Validate required fields.
- [ ] [204] | [FEAT] Add Hot-Reload for Config | [BLOCKS-200] | [TODO]
    - Watch config file for changes.
- [ ] [205] | [FEAT] Add Environment Variable Overrides | [BLOCKS-200] | [TODO]
    - Map env vars to config keys.

### Epic: Error Handling
- [ ] [210] | [FEAT] Define `AppError` Interface in `pkg/protocol` | [INDEPENDENT] | [TODO]
    - Define `Code` and `Message`.
- [ ] [211] | [FEAT] Implement Sentinel Errors | [BLOCKS-210] | [TODO]
    - `ErrNotFound`, `ErrInvalidInput`, `ErrInternal`.
- [ ] [212] | [FEAT] Refactor `Registry` to use `AppError` | [BLOCKS-211] | [TODO]
    - Replace `fmt.Errorf`.
- [ ] [213] | [FEAT] Refactor `Runtime` to use `AppError` | [BLOCKS-211] | [TODO]
    - Map Wazero errors.
- [ ] [214] | [FEAT] Add Error Codes to API Responses | [BLOCKS-210] | [TODO]
    - Return structured JSON error responses.
- [ ] [215] | [FEAT] Add Stack Trace to Internal Errors | [BLOCKS-210] | [TODO]
    - Capture stack on error creation.

### Epic: Validation
- [ ] [220] | [FEAT] Define JSON Schema for `Blueprint` | [INDEPENDENT] | [TODO]
    - Define required fields.
- [ ] [221] | [FEAT] Implement `Blueprint` Validation Logic | [BLOCKS-220] | [TODO]
    - [SEC] Validate `service_id`.
- [ ] [222] | [FEAT] Validate `service_id` format | [BLOCKS-220] | [TODO]
    - Alphanumeric check.
- [ ] [223] | [FEAT] Validate `constraints` size limits | [BLOCKS-220] | [TODO]
    - Prevent DoS via large constraints.
- [ ] [224] | [FEAT] Add Validation Middleware to API | [BLOCKS-221] | [TODO]
    - Validate inputs before processing.

### Epic: CLI Enhancements
- [ ] [230] | [FEAT] Add `ghost-ops version` command | [INDEPENDENT] | [TODO]
    - Print build version.
- [ ] [231] | [FEAT] Add `ghost-ops service list` command | [INDEPENDENT] | [TODO]
    - Table output of services.
- [ ] [232] | [FEAT] Add `ghost-ops service inspect` command | [BLOCKS-231] | [TODO]
    - Detailed JSON view of a service.
- [ ] [233] | [FEAT] Add `ghost-ops service logs` command | [BLOCKS-231] | [TODO]
    - Stream logs (future).
- [ ] [234] | [FEAT] Add `ghost-ops config show` command | [BLOCKS-201] | [TODO]
    - Print loaded config.

### Epic: Runtime Hardening
- [ ] [240] | [FEAT] Configure Wazero Memory Limits | [INDEPENDENT] | [TODO]
    - Prevent guest from consuming all host memory.
- [ ] [241] | [FEAT] Configure Wazero CPU Limits | [INDEPENDENT] | [TODO]
    - Time slicing or instruction limits.
- [ ] [242] | [FEAT] Implement Execution Timeouts | [INDEPENDENT] | [TODO]
    - Context timeout for Invoke.
- [ ] [243] | [FEAT] Verify Async Module Instantiation | [INDEPENDENT] | [TODO]
    - Ensure `Compile` doesn't block main loop.
- [ ] [244] | [FEAT] Implement Module Cache Eviction | [INDEPENDENT] | [TODO]
    - LRU cache for compiled modules.

### Epic: Registry Enhancements
- [ ] [250] | [FEAT] Implement Event Bus for Service Updates | [INDEPENDENT] | [TODO]
    - Internal pub/sub.
- [ ] [251] | [FEAT] Add Metrics for Service Lifecycle | [INDEPENDENT] | [TODO]
    - Track evolution time.
- [ ] [252] | [FEAT] Implement Dependency Graph for Services | [INDEPENDENT] | [TODO]
    - Track service-to-service calls.
- [ ] [253] | [FEAT] Add Service Health Check Logic | [INDEPENDENT] | [TODO]
    - Periodic liveness probes.

## Phase 2: Scale (Distributed & Resilient)

### Epic: Distributed Store (Redis)
- [ ] [300] | [FEAT] Implement Redis Client Setup | [INDEPENDENT] | [TODO]
    - Use `go-redis/v9`.
- [ ] [301] | [FEAT] Implement `RedisStateStore` Adapter | [BLOCKS-300] | [TODO]
    - `Get`, `Set`, `List`.
- [ ] [302] | [FEAT] Implement Distributed Lock for `Reconcile` | [BLOCKS-301] | [TODO]
    - Redis Lock.
- [ ] [303] | [FEAT] Implement Redis Pub/Sub for Events | [BLOCKS-300] | [TODO]
    - Broadcast updates across nodes.
- [ ] [304] | [FEAT] Implement State Compression | [BLOCKS-301] | [TODO]
    - Compress large payloads.

### Epic: Observability (OpenTelemetry)
- [ ] [310] | [FEAT] Setup OpenTelemetry SDK | [INDEPENDENT] | [TODO]
    - Configure Exporter.
- [ ] [311] | [FEAT] Add Tracing Middleware to API | [BLOCKS-310] | [TODO]
    - Trace HTTP requests.
- [ ] [312] | [FEAT] Add Trace Propagation to WASM | [BLOCKS-311] | [TODO]
    - Pass trace context to guest.
- [ ] [313] | [FEAT] Configure OTLP Exporter | [BLOCKS-310] | [TODO]
    - Export to Jaeger/Tempo.
- [ ] [314] | [FEAT] Correlate Logs with Traces | [BLOCKS-311] | [TODO]
    - Add TraceID to logs.

## Phase 3: Future (Autonomous Evolution)
- [ ] [400] | [PROPOSAL] Autonomous Feedback Loop Architecture | [INDEPENDENT] | [TODO]
    - [STRAT] Design self-optimizing loop.
- [ ] [401] | [PROPOSAL] Multi-Language Guest SDK Support | [INDEPENDENT] | [TODO]
    - [STRAT] Design support for Rust/Python.
- [ ] [402] | [PROPOSAL] Automated Vulnerability Scanning | [INDEPENDENT] | [TODO]
    - [STRAT] Design security scanner for generated code.
