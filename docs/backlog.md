# Backlog

## Phase 1: MVP (The Self-Healing Loop)

### Epic: Developer Experience
- [ ] [274] | [DOCS] Create Architecture Diagrams (PlantUML/Mermaid) | [INDEPENDENT] | [TODO]
    - [DOC] Visualize system components and flow.

### Epic: AI Evolution Enhancements
- [x] [284] | [TEST] Add Evals for Code Generation Quality | [INDEPENDENT] | [DONE]
    - [TEST] Automated evaluation of generated code.

## Phase 2: Scale (Distributed & Resilient)

### Epic: Observability (OpenTelemetry)
- [x] [312] | [FEAT] Add Trace Propagation to WASM | [BLOCKS-311] | [DONE]
    - Pass trace context to guest.
- [ ] [313] | [FEAT] Configure OTLP Exporter | [BLOCKS-310] | [TODO]
    - Export to Jaeger/Tempo.
- [x] [314] | [FEAT] Correlate Logs with Traces | [BLOCKS-311] | [DONE]
    - Add TraceID to logs.

### Epic: Service Mesh Lite
- [ ] [320] | [FEAT] Implement Sidecar Proxy Pattern | [INDEPENDENT] | [TODO]
    - [ARCH] Deploy sidecar for network interception.
- [ ] [321] | [FEAT] Implement mTLS between Services | [BLOCKS-320] | [TODO]
    - [SEC] Mutual TLS for service-to-service auth.

### Epic: Advanced Scheduling
- [x] [330] | [FEAT] Implement Priority Queues for Evolution | [INDEPENDENT] | [DONE]
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

### Epic: Autonomous Optimization Loop
- [ ] [500] | [PROPOSAL] Define Metric Thresholds for Optimization | [INDEPENDENT] | [TODO]
    - [STRAT] Establish baseline metrics.
- [ ] [501] | [FEAT] Implement Observer Agent | [BLOCKS-500] | [TODO]
    - [OBS] Continuously monitor runtime state.
- [ ] [502] | [FEAT] Trigger Re-Prompt on Latency Spike | [BLOCKS-501] | [TODO]
    - [REL] Auto-trigger LLM re-prompt if P99 > 500ms.
- [ ] [503] | [FEAT] Trigger Re-Prompt on Error Rate Spike | [BLOCKS-501] | [TODO]
    - [REL] Auto-trigger LLM re-prompt if 5xx > 1%.
- [ ] [504] | [FEAT] Validate Synthesized Code in Shadow Mode | [BLOCKS-502] | [TODO]
    - [TEST] Run new code against mirrored traffic.
- [ ] [505] | [FEAT] Compare Shadow and Primary Metrics | [BLOCKS-504] | [TODO]
    - [OPT] Ensure new code is actually better.
- [ ] [506] | [FEAT] Implement Hot-Swap Promotion | [BLOCKS-505] | [TODO]
    - [REL] Promote shadow to primary gracefully.
- [ ] [507] | [FEAT] Implement Auto-Rollback | [BLOCKS-506] | [TODO]
    - [REL] Revert if new version degrades.
- [ ] [508] | [DOCS] Document Optimization Loop | [INDEPENDENT] | [TODO]
    - [DOC] Detail the ZHO feedback cycle.
- [ ] [509] | [TEST] End-to-End Optimization Test | [BLOCKS-507] | [TODO]
    - [TEST] Full simulation of failure and self-healing.

### Epic: Multi-Language Expansion
- [ ] [600] | [PROPOSAL] Rust Guest SDK Design | [INDEPENDENT] | [TODO]
    - [STRAT] Map host functions to Rust interfaces.
- [ ] [601] | [FEAT] Implement Rust Guest SDK Base | [BLOCKS-600] | [TODO]
    - [FEAT] Basic memory sharing for Rust.
- [ ] [602] | [FEAT] Implement Rust Guest SDK Logger | [BLOCKS-601] | [TODO]
    - [FEAT] Hook up structured logging.
- [ ] [603] | [FEAT] Rust Compiler Evolution Engine | [BLOCKS-601] | [TODO]
    - [FEAT] Support `cargo build --target wasm32-wasi`.
- [ ] [604] | [TEST] Test Rust Compiler Engine | [BLOCKS-603] | [TODO]
    - [TEST] Validate WASM output from Rust source.
- [ ] [605] | [PROPOSAL] Python (Wasm) Guest SDK Design | [INDEPENDENT] | [TODO]
    - [STRAT] Evaluate CPython vs MicroPython for WASM.
- [ ] [606] | [FEAT] Implement Python Guest SDK Base | [BLOCKS-605] | [TODO]
    - [FEAT] Bootstrapping Python environment in WASM.
- [ ] [607] | [FEAT] Python Evolution Engine | [BLOCKS-606] | [TODO]
    - [FEAT] Bundle Python scripts into WASM modules.
- [ ] [608] | [DOCS] Update Examples with Rust/Python | [BLOCKS-604] | [TODO]
    - [DOC] Add basic examples.
- [ ] [609] | [TEST] Cross-Language Interop Testing | [BLOCKS-607] | [TODO]
    - [TEST] Verify Go host can invoke Rust/Python guests uniformly.

### Epic: Advanced Security Hardening
- [ ] [700] | [PROPOSAL] Wasm Sandboxing Review | [INDEPENDENT] | [TODO]
    - [STRAT] Identify potential host escapes.
- [ ] [701] | [FEAT] Implement Capability-based Security | [BLOCKS-700] | [TODO]
    - [SEC] Fine-grained permissions per module.
- [ ] [702] | [FEAT] Enforce Network Egress Policies | [BLOCKS-701] | [TODO]
    - [SEC] Block unauthorized outgoing calls from WASM.
- [ ] [703] | [FEAT] Implement File System Jails | [BLOCKS-701] | [TODO]
    - [SEC] Restrict WASM disk access strictly to allowed dirs.
- [ ] [704] | [FEAT] Automated Vulnerability Scanning | [INDEPENDENT] | [TODO]
    - [SEC] Scan generated code for common CVEs.
- [x] [705] | [FEAT] LLM Prompt Injection Defenses | [INDEPENDENT] | [DONE]
    - [SEC] Sanitize user intents before LLM processing.
- [x] [706] | [TEST] Security Chaos Engineering | [BLOCKS-705] | [DONE]
    - [TEST] Inject malicious payloads into intents.
- [x] [707] | [FEAT] Audit Logging for State Changes | [INDEPENDENT] | [DONE]
    - [SEC] Immutable logs for registry modifications.
- [ ] [708] | [FEAT] Secret Management Integration | [INDEPENDENT] | [TODO]
    - [SEC] Fetch secrets securely (Vault/AWS SM).
- [ ] [709] | [DOCS] Security Architecture Guide | [BLOCKS-708] | [TODO]
    - [DOC] Document trust boundaries.

### Epic: Cluster State Management
- [ ] [800] | [PROPOSAL] Etcd Integration Strategy | [INDEPENDENT] | [TODO]
    - [STRAT] Evaluate Etcd vs Redis for consensus.
- [ ] [801] | [FEAT] Implement Etcd Client Setup | [BLOCKS-800] | [TODO]
    - [FEAT] Basic connection handling.
- [ ] [802] | [FEAT] Etcd StateStore Adapter | [BLOCKS-801] | [TODO]
    - [FEAT] CRUD operations using Etcd.
- [ ] [803] | [FEAT] Distributed Leader Election | [BLOCKS-802] | [TODO]
    - [REL] Ensure single active reconciler loop.
- [ ] [804] | [FEAT] State Synchronization Protocol | [BLOCKS-803] | [TODO]
    - [REL] Propagate state changes to worker nodes.
- [ ] [805] | [TEST] Partition Tolerance Testing | [BLOCKS-804] | [TODO]
    - [TEST] Simulate network splits.
- [x] [806] | [FEAT] Cluster Health Dashboard Data | [INDEPENDENT] | [DONE]
    - [OBS] Aggregate node status.
- [ ] [807] | [FEAT] Node Auto-Discovery | [BLOCKS-804] | [TODO]
    - [REL] Dynamic scaling of worker pool.
- [ ] [808] | [FEAT] Graceful Node Draining | [BLOCKS-807] | [TODO]
    - [REL] Safely evict services on shutdown.
- [ ] [809] | [DOCS] Cluster Setup Guide | [BLOCKS-808] | [TODO]
    - [DOC] Steps to run a multi-node deployment.

### Epic: Dynamic Routing
- [ ] [900] | [PROPOSAL] Layer 7 Gateway Design | [INDEPENDENT] | [TODO]
    - [STRAT] Define routing rules format.
- [ ] [901] | [FEAT] Implement HTTP Gateway | [BLOCKS-900] | [TODO]
    - [FEAT] Map external routes to internal services.
- [ ] [902] | [FEAT] Dynamic Route Reconfiguration | [BLOCKS-901] | [TODO]
    - [REL] Update routes without dropping connections.
- [ ] [903] | [FEAT] Blue/Green Deployment Support | [BLOCKS-902] | [TODO]
    - [REL] Route traffic weights (e.g., 90/10).
- [ ] [904] | [FEAT] Path-based Routing | [BLOCKS-901] | [TODO]
    - [FEAT] e.g., /api/v1/auth -> auth-service.
- [ ] [905] | [FEAT] Header-based Routing | [BLOCKS-901] | [TODO]
    - [FEAT] e.g., X-Beta: true -> beta-service.
- [ ] [906] | [TEST] Gateway Load Testing | [BLOCKS-905] | [TODO]
    - [TEST] Ensure minimal overhead (<2ms).
- [ ] [907] | [FEAT] Websocket Support in Gateway | [BLOCKS-901] | [TODO]
    - [FEAT] Proxy WS connections to WASM.
- [ ] [908] | [FEAT] Gateway Rate Limiting | [BLOCKS-901] | [TODO]
    - [SEC] Global limits per IP.
- [ ] [909] | [DOCS] Routing Configuration Guide | [BLOCKS-908] | [TODO]
    - [DOC] Document gateway usage.
