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

## 🔴 BUG QUEUE — Fix Before All Other Work

> Sourced from automated code audit on 2026-04-22. Severity order: CRITICAL → HIGH → MEDIUM → LOW.  
> All bugs below are real defects (not style issues). Fix gate: 95% test · 0 lint · gosec clean.

---

### CRITICAL

---

**BUG-001 | Goroutine Leak in Sidecar Proxy `handleConnection()`**
`[CRITICAL]` `Token-Impact: 3` `Target: Claude Pro`
`src/sidecar/sidecar.go:104–118`

> **As a** platform operator,
> **I want** the sidecar proxy to cleanly shut down both copy goroutines before closing connections,
> **so that** panics and resource leaks do not occur under normal connection teardown.

- **Root Cause:** `handleConnection()` reads one error from `errCh`, then closes both connections and returns. The second `io.Copy` goroutine is still running and writes to the now-closed `clientConn`/`targetConn`, causing a panic or silent resource leak. Neither goroutine is joined before connections close.
- **Fix:** Drain both goroutines (wait for both `errCh` sends) before closing connections, or use `sync.WaitGroup`.
- Acceptance: No goroutine leak under connection-drop stress test. `go test -race` passes.
- Depends on: none.

---

**BUG-002 | Goroutine Leak in Shadow Invocation (Wazero Host)**
`[CRITICAL]` `Token-Impact: 3` `Target: Claude Pro`
`pkg/runtime/wazero_host.go:539–574`

> **As a** WASM runtime,
> **I want** shadow-mode invocation goroutines to be bounded by the shadow context lifetime,
> **so that** context cancellation does not leave orphaned goroutines permanently blocked.

- **Root Cause:** The goroutine spawned for shadow invocation at line 539 can be left blocked forever if the shadow context cancels before its response is read. `submitResult()` (lines 305–309) uses a non-blocking `select` with a `default` branch — the response is silently dropped and the goroutine hangs.
- **Fix:** After shadow context cancels, ensure the goroutine is either unblocked via a closed channel or collected with a `WaitGroup`.
- Acceptance: Goroutine count does not grow under repeated shadow invocations with context cancellation. `go test -race` passes.
- Depends on: none.

---

**BUG-003 | Context Leak — `context.Background()` Escapes `LoadModule()` Caller Lifetime**
`[CRITICAL]` `Token-Impact: 3` `Target: Claude Pro`
`pkg/runtime/wazero_host.go:421–446`

> **As a** service registry,
> **I want** `LoadModule()` to respect the calling context's cancellation,
> **so that** background goroutines do not outlive their parent request.

- **Root Cause:** Inside `LoadModule()`, a goroutine is spawned with `asyncCtx := context.Background()`. If the caller's context is cancelled (e.g., a timed-out reconcile), the module instantiation keeps running indefinitely. Combined with a window between the module-exists check (line 355–358) and the async goroutine completing, concurrent callers can enter a state where the module is partially registered.
- **Fix:** Derive `asyncCtx` from the caller's context (or a dedicated lifecycle context), and protect the check-then-register sequence under the same lock.
- Acceptance: `LoadModule()` cancels within 100 ms of caller context cancellation. Race detector clean.
- Depends on: none.

---

### HIGH

---

**BUG-004 | TOCTOU Race — Module Existence Check vs. Unload in Wazero Host**
`[HIGH]` `Token-Impact: 3` `Target: Claude Pro`
`pkg/runtime/wazero_host.go:229–231, 517–521`

> **As a** concurrent caller,
> **I want** module lookup and invocation to be atomic,
> **so that** a module cannot be unloaded between the existence check and the invoke call.

- **Root Cause:** `nextCommand()` checks module existence (line 229–231) under a lock. `Invoke()` reads `activeVersions`/`shadowVersions` (lines 517–521) and then releases the lock before using the channel (line 524). `UnloadVersion()` can race between the unlock and the channel read.
- **Fix:** Hold the lock through the channel acquisition, or use a reference-counted handle pattern.
- Acceptance: Concurrent invoke + unload test does not panic or return wrong data. Race detector clean.
- Depends on: none.

---

**BUG-005 | Silent Data Loss — Redis `ListServices()` Drops Records on Type Mismatch**
`[HIGH]` `Token-Impact: 2` `Target: Jules`
`pkg/store/redis_store.go:139–161`

> **As a** service registry,
> **I want** `ListServices()` to return an error when Redis returns unexpected value types,
> **so that** services are never silently dropped from the registry.

- **Root Cause:** `redis.MGet()` returns `[]interface{}`. A failed type assertion to `string` (line 147) silently `continue`s, dropping the record. If Redis storage is corrupted or a different encoding is used, services disappear from `ListServices()` with no error or log.
- **Fix:** Log a warning (with service key) and increment a metric for each skipped entry. Return a wrapped error if the skip rate exceeds a threshold.
- Acceptance: Unit test with a mock returning a non-string value verifies warning is logged and record is counted in a `skipped_records_total` metric.
- Depends on: none.

---

**BUG-006 | Panic — Event Bus Closes Channels While `Publish()` Is Writing**
`[HIGH]` `Token-Impact: 2` `Target: Jules`
`pkg/event/bus.go:60–70`

> **As an** event bus consumer,
> **I want** `Close()` to guarantee no in-flight `Publish()` call writes to a closed channel,
> **so that** graceful shutdown never panics.

- **Root Cause:** `Publish()` acquires `RLock`, reads the subscriber slice, releases the lock, then writes to subscriber channels. `Close()` acquires `Lock` and closes all channels. The window between `Publish()` releasing `RLock` and writing to channels allows `Close()` to close a channel that `Publish()` is about to send on, causing a panic.
- **Fix:** Use a `sync.WaitGroup` in `Publish()` tracked under the write lock, drained in `Close()` before closing channels. Alternatively, recover from channel-closed panics and convert to a sentinel error.
- Acceptance: Concurrent `Publish()` + `Close()` fuzz test (1000 iterations) never panics. Race detector clean.
- Depends on: none.

---

**BUG-007 | Silent HTTP Response Failures — `w.Write()` Return Values Unchecked**
`[HIGH]` `Token-Impact: 1` `Target: GitHub Copilot`
`pkg/api/server.go:99, 127, 144, 172`

> **As an** API client,
> **I want** the server to log when a response write fails,
> **so that** dropped connections are observable and not silently lost.

- **Root Cause:** Four `w.Write()` calls in `server.go` discard the `(n, err)` return. If the client disconnects mid-response, the write fails silently — the operation appears successful server-side but the client received nothing.
- **Fix:** Wrap writes in a helper that logs the error at debug level (not fatal — disconnects are normal).
- Acceptance: Unit test simulating a broken connection verifies the error is logged.
- Depends on: none.

---

**BUG-008 | Memory Leak — Evicted Compiled Modules Never Closed in Wazero Cache**
`[HIGH]` `Token-Impact: 3` `Target: Claude Pro`
`pkg/runtime/wazero_host.go:372–416`

> **As a** long-running runtime host,
> **I want** evicted compiled modules to be explicitly closed,
> **so that** memory does not grow unboundedly under cache churn.

- **Root Cause:** When the compiled-module LRU cache reaches capacity (50 entries, line 399), the oldest entry is evicted from the map (lines 404–409) but its `wazero.CompiledModule` is never closed. The wazero runtime holds internal references to the compiled module; without an explicit `Close()`, those references are never released.
- **Fix:** Call `evicted.Close(ctx)` on the evicted `CompiledModule` after removal from the map, guarded by a check that no live module instance references it.
- Acceptance: Memory profile shows stable RSS under a 200-module churn test. No wazero internal leak warnings.
- Depends on: none.

---

### MEDIUM

---

**BUG-009 | Validation Bypass — Empty Service ID Passes Middleware**
`[MEDIUM]` `Token-Impact: 1` `Target: GitHub Copilot`
`pkg/api/middleware_validation.go:12–22`

> **As a** security gate,
> **I want** an empty or whitespace-only service ID to be rejected at the middleware layer,
> **so that** downstream handlers never receive an empty identifier.

- **Root Cause:** Validation is skipped when `id == ""`. A request to `/services//logs` (double slash) routes with an empty `id`; the guard at line 14 skips regex validation and passes the empty string to handlers that do not re-validate.
- **Fix:** Treat empty string as a validation failure — return `400 Bad Request`.
- Acceptance: `GET /services//logs` returns `400`. Existing tests still pass.
- Depends on: none.

---

**BUG-010 | Implicit Contract — `json_store.go` Crashes if `load()` Returns Nil Services Map**
`[MEDIUM]` `Token-Impact: 1` `Target: GitHub Copilot`
`pkg/store/json_store.go:98–111`

> **As a** state store,
> **I want** `GetService()` to guard against a nil Services map,
> **so that** any future change to `load()` cannot silently cause a nil-map panic.

- **Root Cause:** `GetService()` accesses `sd.Services[serviceID]` (line 107) relying on `load()` always initialising the map. There is no explicit nil check. A future refactor to `load()` that returns a nil map will cause a panic at this line.
- **Fix:** Add `if sd.Services == nil { return nil, ErrNotFound }` after the `load()` call.
- Acceptance: Unit test passes a nil services map through GetService; no panic.
- Depends on: none.

---

**BUG-011 | Race Condition — Log Observer Buffer Cleared While `Observe()` Is Appending**
`[MEDIUM]` `Token-Impact: 2` `Target: Jules`
`pkg/observer/log_observer.go:68, 137–145`

> **As an** observer pipeline,
> **I want** log buffer reads and clears to be atomic with appends,
> **so that** log entries are never silently dropped during a flush.

- **Root Cause:** `flushService()` reads the buffer (line 139), releases the lock, then clears `o.buffers[serviceID] = nil` (line 144). `Observe()` can append a new entry between the read and the clear, which is then overwritten by `nil`. That entry is permanently lost.
- **Fix:** Copy-and-clear the buffer under a single held lock: read the slice, set to nil, then release the lock before flushing the copied slice to the LLM.
- Acceptance: Concurrent observe + flush test under `go test -race` with 10k events shows zero dropped entries.
- Depends on: none.

---

**BUG-012 | Race Condition — `MetricObserver.prevCounters` Map Read/Write Without Lock**
`[MEDIUM]` `Token-Impact: 2` `Target: Jules`
`pkg/observer/metric_observer.go:79–86`

> **As a** metric observer,
> **I want** all access to `prevCounters` to be mutex-protected,
> **so that** concurrent polling does not corrupt delta calculations.

- **Root Cause:** `analyzeMetrics()` reads and writes `o.prevCounters[key]` (lines 79–86) without holding a lock. `Start()` calls `Poll()` on a ticker goroutine while `analyzeMetrics` is not separately locked. Concurrent map read/write causes a fatal data race.
- **Fix:** Add a `sync.Mutex` (or reuse an existing one) guarding all `prevCounters` access.
- Acceptance: `go test -race ./pkg/observer/...` passes with concurrent start + poll calls.
- Depends on: none.

---

**BUG-013 | Integer Truncation — Large State Values Return Wrong Length in Wazero Host**
`[MEDIUM]` `Token-Impact: 2` `Target: Jules`
`pkg/runtime/wazero_host.go:156–160`

> **As a** WASM guest module,
> **I want** `kv_get_len()` to return an accurate length for all values,
> **so that** a value larger than 4 GB does not cause the guest to allocate the wrong buffer size.

- **Root Cause:** `len(val)` returns `int` (64-bit on modern platforms). The code caps at `math.MaxUint32` and silently truncates to `uint32`. A guest allocates a buffer of the returned length, then calls `kv_get()` which copies the full value — writing past the allocated buffer.
- **Fix:** Enforce a hard maximum value size (e.g., 64 MB) in `kv_set()` and return an error if exceeded, eliminating the overflow path entirely.
- Acceptance: Attempting to store a value >64 MB returns `ErrValueTooLarge`. Existing tests pass.
- Depends on: none.

---

**BUG-014 | Blocking Publisher — Slow Registry Subscriber Stalls Event Bus**
`[MEDIUM]` `Token-Impact: 2` `Target: Jules`
`pkg/registry/registry.go:178–196`

> **As a** registry,
> **I want** event bus subscriptions to use buffered channels,
> **so that** a slow registry event handler cannot block the publisher and stall all other consumers.

- **Root Cause:** `StartEventLoop()` subscribes to the event bus. If the in-memory bus's `Subscribe()` returns an unbuffered channel (line 50 of `bus.go`) and the registry's goroutine is slow (e.g., blocked on a store write), every `Publish()` call in the system blocks until this subscriber reads.
- **Fix:** Ensure `Subscribe()` always returns a buffered channel (buffer size ≥ 32). Drop events and log a warning if the buffer fills rather than blocking the publisher.
- Acceptance: Publisher benchmark shows no latency spike when the registry subscriber is artificially delayed 100 ms.
- Depends on: BUG-006.

---

### LOW

---

**BUG-015 | CPU DoS — O(n·m) Log Sanitization on Unbounded Input**
`[LOW]` `Token-Impact: 2` `Target: Jules`
`pkg/observer/log_observer.go:84–115`

> **As a** platform operator,
> **I want** log sanitization to be bounded in time,
> **so that** a malicious guest module cannot exhaust CPU by emitting adversarially large log lines.

- **Root Cause:** `sanitize()` calls `strings.Fields()` which allocates proportionally to input size, then iterates every field for multiple `strings.Contains()` checks. A log line with 1 M space-separated tokens triggers O(n·m) work per `Observe()` call.
- **Fix:** Truncate input to a hard maximum (e.g., 4096 bytes) before sanitization. Add a metric for truncation events.
- Acceptance: Benchmark with a 1 MB log line completes in <1 ms. Existing sanitization tests pass.
- Depends on: none.

---

**BUG-016 | Silently Ignored Config Load Error in CLI Root**
`[LOW]` `Token-Impact: 1` `Target: GitHub Copilot`
`cmd/ghost-ops/cli/root.go:55`

> **As a** CLI operator,
> **I want** a failed config load to surface a warning,
> **so that** corrupted or missing config is diagnosed immediately rather than producing confusing downstream errors.

- **Root Cause:** `_, _ = config.Load()` at line 55 discards both the result and the error. A corrupted YAML file, wrong permissions, or invalid fields silently result in a zero-value config — the server starts with wrong defaults and the operator has no indication why.
- **Fix:** Log the error at `WARN` level (not fatal — commands like `ghost-ops version` are valid without config).
- Acceptance: Running the CLI with a deliberately corrupted config file prints a warning to stderr.
- Depends on: none.

---

**BUG-017 | Misleading Stack Traces — Sentinel Errors Capture Init-Time Callsite**
`[LOW]` `Token-Impact: 1` `Target: GitHub Copilot`
`pkg/protocol/error.go:63–70`

> **As a** developer debugging production errors,
> **I want** stack traces in errors to point to where the error was returned, not where the sentinel was defined,
> **so that** log-based debugging does not require source cross-referencing.

- **Root Cause:** `ErrNotFound`, `ErrAlreadyExists`, etc. are constructed once at package init. Their embedded stack traces always point to `error.go:64`, regardless of where in the codebase the sentinel is returned.
- **Fix:** Wrap sentinels at the return site using `fmt.Errorf("context: %w", ErrNotFound)` rather than returning bare sentinels, and remove stack capture from sentinel construction.
- Acceptance: A returned `ErrNotFound` from `registry.go` wraps the sentinel and carries a stack frame from `registry.go`, not `error.go`.
- Depends on: none.

---

**BUG-018 | Confusing Boundary Logic — P99 Calculation Masks Empty-Slice Edge Case**
`[LOW]` `Token-Impact: 1` `Target: GitHub Copilot`
`pkg/telemetry/inmem.go:100–110`

> **As a** telemetry consumer,
> **I want** the p99 calculation to be obviously correct,
> **so that** a future refactor that removes the outer `len > 0` guard cannot introduce a panic.

- **Root Cause:** `p99Index = int(float64(len) * 0.99)` followed by `if p99Index >= len { p99Index = len - 1 }` produces `p99Index = -1` when `len == 0`. The outer guard at line 100 prevents the panic today, but the inner logic is silently wrong and a future refactor removing the guard will produce an out-of-bounds panic.
- **Fix:** Add an explicit early return inside the p99 block if `len(sortedValues) == 0`.
- Acceptance: Test with empty histogram returns `p99 == 0` and does not panic.
- Depends on: none.

---

**BUG-019 | Integer Overflow — Retry Backoff Overflows `int64` Beyond 31 Doublings**
`[LOW]` `Token-Impact: 1` `Target: GitHub Copilot`
`pkg/resilience/retry.go:80–83`

> **As a** resilience consumer,
> **I want** retry backoff to be capped at a maximum duration,
> **so that** a high `WithMaxAttempts()` value cannot cause `backoff *= 2` to overflow and produce a negative sleep duration.

- **Root Cause:** `backoff *= 2` starting from 100 ms overflows `int64` after 31 doublings (~62 days). `WithMaxAttempts(50)` reaches this after attempt 32 and subsequent `time.Sleep(negative_duration)` returns immediately, producing a tight retry loop instead of backing off.
- **Fix:** Cap backoff at a configurable `maxBackoff` (default 30 s): `if backoff > maxBackoff { backoff = maxBackoff }`.
- Acceptance: With `maxAttempts=50` and `initialBackoff=100ms`, backoff never exceeds `maxBackoff`. Test verifies sleep durations.
- Depends on: none.

---

**BUG-020 | Global Regex Panic Risk — `regexp.MustCompile` Used Outside `init()`**
`[LOW]` `Token-Impact: 1` `Target: GitHub Copilot`
`pkg/api/middleware_validation.go:8`

> **As a** server operator,
> **I want** regex compilation failures to be caught at build time via a test, not at runtime startup,
> **so that** a typo in the regex pattern causes a test failure rather than a production panic.

- **Root Cause:** `regexp.MustCompile(...)` at package-level panics on an invalid pattern. There is no test that imports this package and exercises the `init()` path to verify the pattern compiles. A future edit that breaks the regex will only be caught when the binary first starts.
- **Fix:** Add a `TestServiceIDRegexCompiles` unit test that imports the package and calls the regex, ensuring CI catches any future breakage.
- Acceptance: New test exists and passes. Any invalid regex in `middleware_validation.go` causes `go test ./pkg/api/...` to fail.
- Depends on: none.

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
