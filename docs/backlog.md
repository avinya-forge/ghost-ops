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
- [x] [200] | [FEAT] Add Viper Dependency & Setup | [INDEPENDENT] | [DONE]
    - [SEC] Ensure secure defaults. [TEST] Verify precedence. [OPT] Lazy load.
- [x] [201] | [FEAT] Define `Config` Struct | [BLOCKS-200] | [DONE]
    - [LINT] Add struct tags. [TEST] JSON/YAML serialization.
- [x] [202] | [FEAT] Migrate CLI Flags to Viper Config | [BLOCKS-201] | [DONE]
    - [TEST] Verify flag overrides. [SEC] No secrets in flags.
- [x] [203] | [FEAT] Add Config Validation Logic | [BLOCKS-201] | [DONE]
    - [SEC] Sanitize inputs. [TEST] Invalid config cases.
- [x] [204] | [FEAT] Add Hot-Reload for Config | [BLOCKS-200] | [DONE]
    - [OPT] Debounce events. [TEST] Verify dynamic update.
- [x] [205] | [FEAT] Add Environment Variable Overrides | [BLOCKS-200] | [DONE]
    - [SEC] Mask secrets in logs. [TEST] Env var mapping.

### Epic: Error Handling
- [x] [210] | [FEAT] Define `AppError` Interface in `pkg/protocol` | [INDEPENDENT] | [DONE]
    - [LINT] GoDoc. [TEST] Interface compliance. [OPT] O(1) error wrapping.
- [x] [211] | [FEAT] Implement Sentinel Errors | [BLOCKS-210] | [DONE]
    - [TEST] Error type assertions. [LINT] Constant naming.
- [ ] [212] | [FEAT] Refactor `Registry` to use `AppError` | [BLOCKS-211] | [TODO]
    - [TEST] Verify error propagation. [OPT] Minimize allocation.
- [ ] [213] | [FEAT] Refactor `Runtime` to use `AppError` | [BLOCKS-211] | [TODO]
    - [TEST] Map Wazero exit codes. [SEC] Mask internal details.
- [ ] [214] | [FEAT] Add Error Codes to API Responses | [BLOCKS-210] | [TODO]
    - [TEST] HTTP 4xx/5xx mappings. [SEC] No stack traces in production.
- [ ] [215] | [FEAT] Add Stack Trace to Internal Errors | [BLOCKS-210] | [TODO]
    - [OPT] Capture only on debug. [SEC] Redact sensitive paths.

### Epic: Validation
- [x] [220] | [FEAT] Define JSON Schema for `Blueprint` | [INDEPENDENT] | [DONE]
    - [LINT] Valid JSON. [TEST] Schema validation tests.
- [x] [221] | [FEAT] Implement `Blueprint` Validation Logic | [BLOCKS-220] | [DONE]
    - [SEC] Sanitize inputs. [TEST] Boundary checks.
- [x] [222] | [FEAT] Validate `service_id` format | [BLOCKS-220] | [DONE]
    - [SEC] Regex `^[a-zA-Z0-9-]+$`. [TEST] Inject special chars.
- [x] [223] | [FEAT] Validate `constraints` size limits | [BLOCKS-220] | [DONE]
    - [SEC] Max size 1MB. [TEST] Large payload DoS.
- [ ] [224] | [FEAT] Add Validation Middleware to API | [BLOCKS-221] | [TODO]
    - [OPT] Fail fast. [TEST] Middleware chain order.

### Epic: CLI Enhancements
- [x] [230] | [FEAT] Add `ghost-ops version` command | [INDEPENDENT] | [DONE]
    - [TEST] Compare output with runtime. [LINT] Flag parsing.
- [x] [231] | [FEAT] Add `ghost-ops service list` command | [INDEPENDENT] | [DONE]
    - [OPT] Pagination support. [TEST] Empty/Large list.
- [x] [232] | [FEAT] Add `ghost-ops service inspect` command | [BLOCKS-231] | [DONE]
    - [SEC] Redact secrets. [TEST] Not found case.
- [ ] [233] | [FEAT] Add `ghost-ops service logs` command | [BLOCKS-231] | [TODO]
    - [OPT] Stream buffer size. [TEST] Connection drop handling.
- [x] [234] | [FEAT] Add `ghost-ops config show` command | [BLOCKS-201] | [DONE]
    - [SEC] Mask sensitive keys. [TEST] Verify output matches loaded config.

### Epic: Runtime Hardening
- [x] [240] | [FEAT] Configure Wazero Memory Limits | [INDEPENDENT] | [DONE]
    - [SEC] Max 128MB/guest. [TEST] OOM handling.
- [ ] [241] | [FEAT] Configure Wazero CPU Limits | [INDEPENDENT] | [TODO]
    - [SEC] Max instructions/call. [TEST] Infinite loop break.
- [x] [242] | [FEAT] Implement Execution Timeouts | [INDEPENDENT] | [DONE]
    - [REL] 5s default. [TEST] Context cancellation.
- [ ] [243] | [FEAT] Verify Async Module Instantiation | [INDEPENDENT] | [TODO]
    - [OPT] Non-blocking I/O. [TEST] Load module under load.
- [ ] [244] | [FEAT] Implement Module Cache Eviction | [INDEPENDENT] | [TODO]
    - [OPT] LRU policy. [TEST] Cache hit/miss ratio.

### Epic: Registry Enhancements
- [ ] [250] | [FEAT] Implement Event Bus for Service Updates | [INDEPENDENT] | [TODO]
    - [OPT] Async channel. [TEST] Subscriber notification.
- [ ] [251] | [FEAT] Add Metrics for Service Lifecycle | [INDEPENDENT] | [TODO]
    - [OBS] Gauge/Histogram. [TEST] Metric emission.
- [ ] [252] | [FEAT] Implement Dependency Graph for Services | [INDEPENDENT] | [TODO]
    - [ARCH] DAG traversal. [TEST] Cycle detection.
- [ ] [253] | [FEAT] Add Service Health Check Logic | [INDEPENDENT] | [TODO]
    - [REL] 10s interval. [TEST] Unhealthy service purge.

### Epic: Hardening (Observability & Security)
- [x] [260] | [TEST] Add Unit Tests for `pkg/telemetry` | [INDEPENDENT] | [DONE]
    - [TEST] Achieve >95% coverage for metrics collector.
- [ ] [261] | [TEST] Add Unit Tests for `pkg/logging` | [INDEPENDENT] | [TODO]
    - [TEST] Achieve >95% coverage for structured logger.
- [ ] [262] | [TEST] Add Unit Tests for `pkg/sdk/guest` | [INDEPENDENT] | [TODO]
    - [TEST] Achieve >95% coverage for guest SDK.
- [ ] [263] | [SEC] Run `gosec` Security Scan | [INDEPENDENT] | [TODO]
    - [SEC] Audit and fix all high-severity issues.
- [ ] [264] | [OPT] Benchmark `RuntimeHost` Performance | [INDEPENDENT] | [TODO]
    - [OPT] Ensure overhead < 5ms per invoke.
- [ ] [265] | [TEST] Add Integration Tests for API | [INDEPENDENT] | [TODO]
    - [TEST] End-to-end flow coverage.

### Epic: Developer Experience
- [ ] [270] | [FEAT] Add Makefile `dev` target (hot reload) | [INDEPENDENT] | [TODO]
    - [DX] Auto-restart server on source change.
- [ ] [271] | [DOCS] Generate API Swagger/OpenAPI Spec | [INDEPENDENT] | [TODO]
    - [DOC] Auto-generate spec from code/comments.
- [ ] [272] | [FEAT] Add Pre-commit Hooks (Git) | [INDEPENDENT] | [TODO]
    - [DX] Enforce lint/test before commit.
- [ ] [273] | [FEAT] Add `ghost-ops init` command | [INDEPENDENT] | [TODO]
    - [DX] Bootstrap new project structure.
- [ ] [274] | [DOCS] Create Architecture Diagrams (PlantUML/Mermaid) | [INDEPENDENT] | [TODO]
    - [DOC] Visualize system components and flow.
- [ ] [275] | [DOCS] Update `README.md` with new features | [INDEPENDENT] | [TODO]
    - [DOC] Reflect latest capabilities.

### Epic: AI Evolution Enhancements
- [ ] [280] | [FEAT] Support Custom System Prompts | [INDEPENDENT] | [TODO]
    - [FEAT] Allow overriding default LLM prompt.
- [ ] [281] | [FEAT] Implement Prompt Caching | [INDEPENDENT] | [TODO]
    - [OPT] Cache prompts to save tokens/cost.
- [ ] [282] | [FEAT] Add Token Usage Metrics | [INDEPENDENT] | [TODO]
    - [OBS] Track input/output tokens per evolution.
- [ ] [283] | [FEAT] Support Ollama Provider | [INDEPENDENT] | [TODO]
    - [FEAT] Add local LLM support via Ollama.
- [ ] [284] | [TEST] Add Evals for Code Generation Quality | [INDEPENDENT] | [TODO]
    - [TEST] Automated evaluation of generated code.

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

### Epic: Service Mesh Lite
- [ ] [320] | [FEAT] Implement Sidecar Proxy Pattern | [INDEPENDENT] | [TODO]
    - [ARCH] Deploy sidecar for network interception.
- [ ] [321] | [FEAT] Implement mTLS between Services | [BLOCKS-320] | [TODO]
    - [SEC] Mutual TLS for service-to-service auth.
- [ ] [322] | [FEAT] Implement Circuit Breaker | [INDEPENDENT] | [TODO]
    - [REL] Fail fast on downstream failures.
- [ ] [323] | [FEAT] Implement Retry Logic with Backoff | [INDEPENDENT] | [TODO]
    - [REL] Exponential backoff for transient errors.
- [ ] [324] | [FEAT] Implement Rate Limiting (Token Bucket) | [INDEPENDENT] | [TODO]
    - [REL] Protect services from overload.

### Epic: Advanced Scheduling
- [ ] [330] | [FEAT] Implement Priority Queues for Evolution | [INDEPENDENT] | [TODO]
    - [ALG] Min-heap for task priority. [TEST] Priority order.
- [ ] [331] | [FEAT] Implement Resource-Aware Scheduling | [BLOCKS-330] | [TODO]
    - [OPT] Bin packing algorithm. [TEST] Maximize density.

## Phase 3: Future (Autonomous Evolution)
- [ ] [400] | [PROPOSAL] Autonomous Feedback Loop Architecture | [INDEPENDENT] | [TODO]
    - [STRAT] Design self-optimizing loop.
- [ ] [401] | [PROPOSAL] Multi-Language Guest SDK Support | [INDEPENDENT] | [TODO]
    - [STRAT] Design support for Rust/Python.
- [ ] [402] | [PROPOSAL] Automated Vulnerability Scanning | [INDEPENDENT] | [TODO]
    - [STRAT] Design security scanner for generated code.
