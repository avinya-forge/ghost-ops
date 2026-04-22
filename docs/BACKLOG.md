# Ghost Ops — Backlog

> **SSOT Owner:** Lead LLM Architect | **Last Updated:** 2026-04-22 | **Version:** 0.5.1  
> **Format:** User Story | Token-Impact: 1 (minimal) – 5 (max) | Target Engine per task  
> **State File:** `/docs/.system_state` | **Roadmap Ref:** `/docs/ROADMAP.md`

---

## Token-Impact Scale

| Score | Meaning |
|---|---|
| 1 | Snippet / single function — safe for GitHub Copilot or Jules |
| 2 | Small feature (<100 LOC) — Jules or Antigravity |
| 3 | Mid-size feature (100–500 LOC + tests) — Jules or Claude Pro |
| 4 | Large feature / system design — Claude Pro |
| 5 | High-risk / large-context (>8k tokens, cross-cutting) — Claude Pro only |

---

## PHASE-2 — Scale & Distribution

### MILESTONE 2.1: Distributed Consensus

---

**TASK-800 | Etcd Integration Strategy**  
`[BLOCKED]` `Token-Impact: 5` `Target: Claude Pro`

> **As a** cluster operator,  
> **I want** a formal evaluation of Etcd vs Redis for distributed consensus,  
> **so that** the system can run a single active reconciler across multiple nodes without split-brain.

- Acceptance: ADR written to `/docs/architecture/decisions.md`. Trade-offs documented. Decision locked.
- Risk: HIGH — architecture decision with long-tail consequences.
- Depends on: none.

---

**TASK-801 | Implement Etcd Client Setup**  
`[PENDING]` `Token-Impact: 2` `Target: Jules`

> **As a** Go host process,  
> **I want** a stable Etcd client connection with retry and TLS,  
> **so that** downstream state operations never silently fail.

- Acceptance: `pkg/store/etcd_client.go` · unit tests ≥95% · 0 lint.
- Risk: LOW.
- Depends on: TASK-800.

---

**TASK-802 | Etcd Statestore Adapter**  
`[PENDING]` `Token-Impact: 2` `Target: Jules`

> **As a** service registry,  
> **I want** CRUD operations backed by Etcd,  
> **so that** state survives node restarts.

- Acceptance: Implements `protocol.StateStore` interface · `etcd_store_test.go` ≥95% coverage.
- Risk: LOW.
- Depends on: TASK-801.

---

**TASK-803 | Distributed Leader Election**  
`[PENDING]` `Token-Impact: 4` `Target: Claude Pro`

> **As a** multi-node deployment,  
> **I want** exactly one active reconciler elected at any time,  
> **so that** duplicate reconciliations and race conditions are eliminated.

- Acceptance: Leader election via Etcd lease. Failover <5 s verified under partition simulation.
- Risk: HIGH — correctness-critical.
- Depends on: TASK-802.

---

**TASK-804 | State Synchronization Protocol**  
`[PENDING]` `Token-Impact: 4` `Target: Claude Pro`

> **As a** worker node,  
> **I want** to receive state-change events from the leader via watch streams,  
> **so that** all nodes reflect the current service state within 1 s.

- Acceptance: End-to-end test: leader change → follower reflects within 1 s.
- Risk: HIGH.
- Depends on: TASK-803.

---

**TASK-805 | Partition Tolerance Testing**  
`[PENDING]` `Token-Impact: 3` `Target: Jules`

> **As a** QA engineer,  
> **I want** automated network-partition simulation tests,  
> **so that** the cluster's CAP trade-offs are proven and documented.

- Acceptance: Chaos test in `pkg/store/partition_test.go` using tc/iptables simulation.
- Risk: MEDIUM.
- Depends on: TASK-804.

---

**TASK-807 | Node Auto-Discovery**  
`[PENDING]` `Token-Impact: 3` `Target: Jules`

> **As a** cluster admin,  
> **I want** new worker nodes to register themselves automatically,  
> **so that** scaling requires zero manual configuration.

- Acceptance: Node joins cluster within 30 s of startup with zero config change on leader.
- Risk: MEDIUM.
- Depends on: TASK-804.

---

**TASK-808 | Graceful Node Draining**  
`[PENDING]` `Token-Impact: 2` `Target: Jules`

> **As a** node operator,  
> **I want** SIGTERM to drain active WASM invocations before shutdown,  
> **so that** in-flight requests complete without error.

- Acceptance: No in-flight invocations dropped during graceful shutdown test. Drain timeout: 30 s.
- Risk: LOW.
- Depends on: TASK-807.

---

**TASK-809 | Cluster Setup Guide**  
`[PENDING]` `Token-Impact: 1` `Target: Antigravity / Doc Generator`

> **As a** platform engineer,  
> **I want** a step-by-step guide for deploying a multi-node Ghost Ops cluster,  
> **so that** the system can be reproduced from scratch without tribal knowledge.

- Acceptance: `/docs/cluster-setup.md` created. Runbook verified on clean VM.
- Risk: LOW.
- Depends on: TASK-808.

---

### MILESTONE 2.2: Security Hardening

---

**TASK-702 | Enforce Network Egress Policies**  
`[PENDING]` `Token-Impact: 3` `Target: Claude Pro`

> **As a** security auditor,  
> **I want** WASM modules to be blocked from unauthorized outbound network calls,  
> **so that** compromised guest code cannot exfiltrate data.

- Acceptance: Modules blocked unless URL matches `config.capabilities.network.allowed_hosts`. Test: unauthorized call → returns error, no packet leaves host.
- Risk: HIGH — bypass = data exfiltration.
- Depends on: EPIC 701 (DONE).

---

**TASK-703 | Implement File System Jails**  
`[PENDING]` `Token-Impact: 3` `Target: Claude Pro`

> **As a** security auditor,  
> **I want** WASM disk access strictly limited to declared allowed directories,  
> **so that** guest modules cannot read host secrets or system files.

- Acceptance: Any path outside allowed dirs returns `EACCES`. Fuzz test with path-traversal inputs.
- Risk: HIGH.
- Depends on: EPIC 701 (DONE).

---

**TASK-704 | Automated Vulnerability Scanning**  
`[PENDING]` `Token-Impact: 2` `Target: Jules`

> **As a** CI pipeline,  
> **I want** generated Go source code scanned for CVEs before compilation,  
> **so that** the evolution engine cannot unknowingly ship vulnerable code.

- Acceptance: `gosec` + `govulncheck` run on every synthesized file. High-severity findings block deployment.
- Risk: MEDIUM.
- Depends on: none.

---

**TASK-708 | Secret Management Integration**  
`[PENDING]` `Token-Impact: 4` `Target: Claude Pro`

> **As a** service operator,  
> **I want** runtime secrets fetched from Vault or AWS Secrets Manager,  
> **so that** credentials never appear in source code, blueprints, or logs.

- Acceptance: `pkg/secrets/` package with `Get(key) (string, error)`. Secrets masked in all log output.
- Risk: HIGH — EU AI Act Article 15 (data protection) applies.
- Depends on: TASK-702.

---

**TASK-709 | Security Architecture Guide**  
`[BLOCKED]` `Token-Impact: 1` `Target: Antigravity / Doc Generator`

> **As a** new contributor,  
> **I want** a documented trust-boundary diagram and threat model,  
> **so that** security decisions are reproducible without consulting the original author.

- Acceptance: `/docs/security-architecture.md` with Mermaid trust-boundary diagram.
- Risk: LOW.
- Depends on: TASK-702, TASK-703, TASK-708.

---

## PHASE-3a — Multi-Language Autonomy

### MILESTONE 3.1: Rust Guest SDK

---

**TASK-600 | Rust Guest SDK Design**  
`[BLOCKED]` `Token-Impact: 5` `Target: Claude Pro`

> **As an** evolution engine,  
> **I want** a formal ABI mapping from Go host functions to Rust `extern` declarations,  
> **so that** Rust WASM modules can call `kv_get`, `kv_set`, `log`, and `rpc` identically to Go guests.

- Acceptance: RFC written to `/docs/architecture/decisions.md`. Approved before TASK-601 begins.
- Risk: HIGH — ABI mistakes require full SDK rewrite.
- Depends on: none.

---

**TASK-601 | Implement Rust Guest SDK Base**  
`[PENDING]` `Token-Impact: 3` `Target: Jules`

> **As a** Rust WASM author,  
> **I want** a `ghost_ops_sdk` crate providing host function bindings,  
> **so that** I can write WASM services in Rust without manual `extern` declarations.

- Acceptance: `pkg/sdk/rust/` crate compiles to `wasm32-wasi`. Basic memory sharing works.
- Risk: MEDIUM.
- Depends on: TASK-600.

---

**TASK-602 | Implement Rust Guest SDK Logger**  
`[PENDING]` `Token-Impact: 1` `Target: Jules`

> **As a** Rust WASM module,  
> **I want** structured log calls that surface in the host's log stream,  
> **so that** debugging Rust guests is identical to debugging Go guests.

- Acceptance: `ghost_ops_sdk::log!(level, msg)` appears in `pkg/logging` output. JSON format.
- Risk: LOW.
- Depends on: TASK-601.

---

**TASK-603 | Rust Compiler Evolution Engine**  
`[PENDING]` `Token-Impact: 3` `Target: Jules`

> **As an** evolution engine,  
> **I want** to compile Rust source code to WASM using `cargo build --target wasm32-wasi`,  
> **so that** the AI can synthesize and deploy Rust services the same way it handles Go.

- Acceptance: `pkg/evolution/rust_compiler.go` implements `Compiler` interface. `rust_compiler_test.go` ≥95%.
- Risk: MEDIUM.
- Depends on: TASK-601.

---

**TASK-604 | Test Rust Compiler Engine**  
`[PENDING]` `Token-Impact: 2` `Target: Jules`

> **As a** QA engineer,  
> **I want** end-to-end tests that synthesize, compile, and invoke a Rust WASM module,  
> **so that** Rust evolution is proven before Q3 gate.

- Acceptance: `pkg/evolution/rust_compiler_test.go` covers valid + invalid Rust source. WASM output invocable.
- Risk: LOW.
- Depends on: TASK-603.

---

**TASK-605 | Python (WASM) Guest SDK Design**  
`[BLOCKED]` `Token-Impact: 5` `Target: Claude Pro`

> **As an** evolution engine architect,  
> **I want** a formal evaluation of CPython vs MicroPython for WASM,  
> **so that** the Python SDK choice is justified and irreversible footguns are avoided.

- Acceptance: Decision record in `/docs/architecture/decisions.md`. Bundle size, startup latency, and ABI coverage compared.
- Risk: HIGH.
- Depends on: TASK-604.

---

**TASK-606 | Implement Python Guest SDK Base**  
`[PENDING]` `Token-Impact: 3` `Target: Jules`

> **As a** Python WASM author,  
> **I want** a bootstrapped Python WASM environment with host bindings,  
> **so that** I can write WASM services in Python without embedded C.

- Acceptance: Python WASM module boots and calls `kv_get`. Cold-start <50 ms.
- Risk: MEDIUM.
- Depends on: TASK-605.

---

**TASK-607 | Python Evolution Engine**  
`[PENDING]` `Token-Impact: 3` `Target: Jules`

> **As an** evolution engine,  
> **I want** to bundle a Python script into a WASM module,  
> **so that** AI-synthesized Python code deploys identically to Go and Rust.

- Acceptance: `pkg/evolution/python_compiler.go` implements `Compiler`. `python_compiler_test.go` ≥95%.
- Risk: MEDIUM.
- Depends on: TASK-606.

---

**TASK-608 | Update Examples with Rust/Python**  
`[PENDING]` `Token-Impact: 1` `Target: Antigravity`

> **As a** new developer,  
> **I want** working Rust and Python "hello-world" examples,  
> **so that** onboarding to multi-language development takes <30 minutes.

- Acceptance: `examples/basic/hello-world-rust/` and `examples/basic/hello-world-python/` build and run.
- Risk: LOW.
- Depends on: TASK-604, TASK-607.

---

**TASK-609 | Cross-Language Interop Testing**  
`[PENDING]` `Token-Impact: 3` `Target: Jules`

> **As a** platform engineer,  
> **I want** a single integration test suite that invokes Go, Rust, and Python WASM modules via the same API,  
> **so that** language parity is continuously verified in CI.

- Acceptance: `pkg/api/integration_multilang_test.go`. All three languages pass identical invocation tests. CI green.
- Risk: MEDIUM.
- Depends on: TASK-608.

---

### MILESTONE 3.3: Gateway Enhancements

---

**TASK-902 | Dynamic Route Reconfiguration**  
`[PENDING]` `Token-Impact: 3` `Target: Jules`

> **As a** platform operator,  
> **I want** gateway routes to update without dropping active connections,  
> **so that** routing changes are zero-downtime.

- Acceptance: Atomic route swap under load. No 5xx during reconfiguration. Test: 1000 concurrent RPS during swap.
- Risk: MEDIUM.
- Depends on: EPIC 901 (DONE).

---

**TASK-903 | Blue/Green Deployment Support**  
`[PENDING]` `Token-Impact: 3` `Target: Jules`

> **As a** deployment engineer,  
> **I want** to split traffic between two service versions by weight (e.g., 90/10),  
> **so that** new versions can be validated on live traffic before full cutover.

- Acceptance: `POST /routes/{id}/weights` accepted. Traffic distribution within ±2% of target over 10k requests.
- Risk: MEDIUM.
- Depends on: TASK-902.

---

**TASK-906 | Gateway Load Testing**  
`[PENDING]` `Token-Impact: 2` `Target: Jules`

> **As a** performance engineer,  
> **I want** an automated load test that validates gateway overhead <2 ms p99,  
> **so that** the SLA is provably met before Q3 gate.

- Acceptance: `k6` or `vegeta` test script in `tests/load/`. P99 latency <2 ms at 10k RPS.
- Risk: LOW.
- Depends on: TASK-903.

---

**TASK-907 | WebSocket Support in Gateway**  
`[PENDING]` `Token-Impact: 3` `Target: Jules`

> **As a** client application,  
> **I want** the gateway to proxy WebSocket connections to WASM services,  
> **so that** real-time bidirectional communication is supported without a separate path.

- Acceptance: WS upgrade handshake proxied correctly. 1000-message round-trip test passes.
- Risk: MEDIUM.
- Depends on: TASK-902.

---

**TASK-908 | Gateway Rate Limiting**  
`[PENDING]` `Token-Impact: 2` `Target: GitHub Copilot`

> **As a** platform operator,  
> **I want** configurable per-IP rate limits enforced at the gateway,  
> **so that** individual clients cannot exhaust host resources.

- Acceptance: Requests beyond limit receive `429 Too Many Requests`. Limit configurable in `config.yaml`. Test: burst + sustained load.
- Risk: LOW.
- Depends on: TASK-902.

---

**TASK-909 | Routing Configuration Guide**  
`[BLOCKED]` `Token-Impact: 1` `Target: Antigravity / Doc Generator`

> **As a** new platform engineer,  
> **I want** documentation covering path routing, header routing, blue/green, and WebSocket configuration,  
> **so that** the gateway is operable without reading source code.

- Acceptance: `/docs/gateway-routing.md` with examples for every routing mode. Reviewed by a non-author.
- Risk: LOW.
- Depends on: TASK-907, TASK-908.

---

## PHASE-3b — Full ZHO

---

**TASK-400 | Autonomous Feedback Loop Architecture**  
`[BLOCKED]` `Token-Impact: 5` `Target: Claude Pro`

> **As a** Ghost Ops system,  
> **I want** a formal design for the self-optimization loop (metrics → re-prompt → shadow → hot-swap),  
> **so that** the system can improve itself without human intervention.

- Acceptance: Architecture RFC in `/docs/architecture/decisions.md`. API, DB schema, and auth designed. Reviewed by two humans before implementation begins.
- Risk: HIGH — EU AI Act Article 14 (human oversight) must be accounted for in design.
- Depends on: TASK-804 (state sync), EPIC 500 (observer).

---

**TASK-1003 | Expand Log Retention Policies**  
`[PENDING]` `Token-Impact: 2` `Target: Jules`

> **As a** platform operator,  
> **I want** configurable TTL for streaming logs per service,  
> **so that** storage costs are controlled without losing audit trails.

- Acceptance: `log_retention_days` per-service config. Expiry enforced on schedule. Test: logs purged after TTL.
- Risk: LOW.
- Depends on: none.

---

**TASK-1004 | Enhance Log Filtering UI**  
`[BLOCKED]` `Token-Impact: 3` `Target: Claude Pro`

> **As a** support engineer,  
> **I want** UI components for filtering logs by level, service, and time range,  
> **so that** root-cause analysis time is reduced from minutes to seconds.

- Acceptance: Filter panel design approved. Component renders without UI framework debt.
- Risk: MEDIUM.
- Depends on: TASK-1003.

---

**TASK-1005 | Integrate External Auth Providers**  
`[PENDING]` `Token-Impact: 4` `Target: Claude Pro` `[HIGH-RISK]`

> **As a** platform administrator,  
> **I want** OAuth2/OIDC authentication for gateway API access,  
> **so that** user identity is federated and credentials are never stored locally.

- Acceptance: Auth middleware in `pkg/api/middleware_auth.go`. Tested against a real OIDC provider (e.g., Keycloak). EU AI Act Article 13 transparency requirements documented.
- Risk: HIGH — credential handling, EU AI Act applies.
- Depends on: TASK-708.

---

**TASK-1006 | Optimize Runtime Memory Allocation**  
`[PENDING]` `Token-Impact: 3` `Target: Jules`

> **As a** runtime host,  
> **I want** Wazero memory pooling and reuse across invocations,  
> **so that** per-invocation GC pressure is reduced and p99 latency improves.

- Acceptance: Benchmark in `pkg/runtime/benchmark_test.go`. Memory alloc/op reduced ≥30% vs baseline. No regression in existing tests.
- Risk: MEDIUM.
- Depends on: none.

---

## Vaulted (Complete)

| EPIC / Task | Title | Shipped |
|---|---|---|
| EPIC 500 | Observer Agent & Metric Thresholds | v0.5.0 |
| EPIC 701 | Capability-Based Security (701.1–701.4) | v0.6.0 |
| EPIC 901 | HTTP Gateway — Path + Header Routing | v0.6.0 |
| TASK 320 | Sidecar Proxy Pattern | v0.5.0 |
| TASK 321 | mTLS Between Services | v0.5.0 |
| TASK 331 | Resource-Aware Scheduling | v0.5.0 |
| TASK 1000–1002 | Log Streaming + Anomaly Detection | v0.5.0 |
| TASK 1200 | Priority Queue for Evolution Engine | v0.5.0 |

---

*V-Score (Self-Reported): 4.6 / 5.0 — All tasks carry Token-Impact + Target Engine. Flat hierarchy maintained.*
