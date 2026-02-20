# Backlog

## Phase 0: Hygiene (Foundations)
- [ ] [LINT] Add `.golangci.yml` configuration to root <!-- id: 100 --> [INDEPENDENT]
    - [LINT] Enable `staticcheck`, `gofmt`, `govet`.
- [ ] [LINT] Fix `staticcheck` issues (SA4006 unused variables) <!-- id: 101 --> [BLOCKS-100]
    - [LINT] Run linter and fix reported issues.
- [ ] [TEST] Add Unit Tests for `Registry` Error Paths <!-- id: 102 --> [INDEPENDENT]
    - [UNIT] Coverage for `Reconcile` failure scenarios (store error, runtime error).
- [ ] [TEST] Add Unit Tests for `WazeroRuntimeHost` Edge Cases <!-- id: 103 --> [INDEPENDENT]
    - [UNIT] Coverage for `LoadModule` with invalid WASM.
    - [UNIT] Coverage for `Close` with active modules.
- [ ] [TEST] Add Unit Tests for `IntentSource` <!-- id: 104 --> [INDEPENDENT]
    - [UNIT] Coverage for empty file, invalid JSON, missing file.
- [ ] [REFACTOR] Restructure `examples/` directory <!-- id: 105 --> [INDEPENDENT]
    - Move `hello-world` to `examples/basic/`.
    - Create `examples/advanced/` placeholder.

## Phase 1: MVP (The Self-Healing Loop)
### Configuration Management
- [ ] [FEAT] Add Viper Dependency & Setup <!-- id: 200 --> [INDEPENDENT]
    - [SEC] Ensure secure defaults.
- [ ] [FEAT] Define `Config` Struct <!-- id: 201 --> [BLOCKS-200]
    - Define fields for `Server`, `Log`, `Store`, `Runtime`.
- [ ] [FEAT] Migrate CLI Flags to Viper Config <!-- id: 202 --> [BLOCKS-201]
    - Replace `flag` package usage in `main.go`.

### Error Handling
- [ ] [FEAT] Define `AppError` Interface in `pkg/protocol` <!-- id: 210 --> [INDEPENDENT]
    - Define `Code` (string) and `Message` (string).
- [ ] [FEAT] Implement Sentinel Errors <!-- id: 211 --> [BLOCKS-210]
    - `ErrNotFound`, `ErrInvalidInput`, `ErrInternal`.
- [ ] [FEAT] Refactor `Registry` to use `AppError` <!-- id: 212 --> [BLOCKS-211]
    - Replace `fmt.Errorf` with typed errors.
- [ ] [FEAT] Refactor `Runtime` to use `AppError` <!-- id: 213 --> [BLOCKS-211]
    - Map Wazero errors to `AppError`.

### Validation
- [ ] [FEAT] Define JSON Schema for `Blueprint` <!-- id: 220 --> [INDEPENDENT]
    - Define required fields: `service_id`, `intent`.
- [ ] [FEAT] Implement `Blueprint` Validation Logic <!-- id: 221 --> [BLOCKS-220]
    - [SEC] Validate `service_id` format (alphanumeric).
    - [SEC] Validate `constraints` size limits.

### CLI Enhancements
- [ ] [FEAT] Add `ghost-ops version` command <!-- id: 230 --> [INDEPENDENT]
    - Print build version and commit hash.
- [ ] [FEAT] Add `ghost-ops service list` command <!-- id: 231 --> [INDEPENDENT]
    - [OPT] Use `Registry.ListServices`.
    - Format output as table.

## Phase 2: Scale (Distributed & Resilient)
### Distributed Store (Redis)
- [ ] [FEAT] Implement Redis Client Setup <!-- id: 300 --> [INDEPENDENT]
    - Use `go-redis/v9`.
- [ ] [FEAT] Implement `RedisStateStore` Adapter <!-- id: 301 --> [BLOCKS-300]
    - Implement `Get`, `Set`, `ListServices`.
- [ ] [FEAT] Implement Distributed Lock for `Reconcile` <!-- id: 302 --> [BLOCKS-301]
    - Prevent multiple instances from reconciling same blueprint.

### Observability (OpenTelemetry)
- [ ] [FEAT] Setup OpenTelemetry SDK <!-- id: 310 --> [INDEPENDENT]
    - Configure Exporter (stdout or OTLP).
- [ ] [FEAT] Add Tracing Middleware to API <!-- id: 311 --> [BLOCKS-310]
    - Trace HTTP requests.

## Phase 3: Future (Autonomous Evolution)
- [ ] [ARCH] Design Autonomous Feedback Loop Architecture <!-- id: 400 -->
- [ ] [FEAT] Implement Runtime Metric Analysis Engine <!-- id: 401 -->
- [ ] [FEAT] Implement Python Guest SDK <!-- id: 402 -->
- [ ] [FEAT] Implement Rust Guest SDK <!-- id: 403 -->
- [ ] [SEC] Implement Vulnerability Scanner for WASM <!-- id: 404 -->
