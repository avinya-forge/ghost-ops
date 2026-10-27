# Ghost Ops — Backlog (Execution SSOT)

> **The execution ladder.** Bugs first (priority-ordered), then features, then
> demo, then 30-day plan, then vaulted history.
> **North star (the why):** [`VISION.md`](./VISION.md).
> **Engineering rules (the how):** [`STANDARDS.md`](./STANDARDS.md).
> 
## 🌌 6–8 Month Strategic Roadmap & Horizon Plan

| Horizon | Core Strategic Milestone | Lead Focus | Key Deliverables |
|:---:|:---|:---|:---|
| **Month 1–2** | **EU AI Act & Capability Security Enforcement** | Security & Sandboxing | Art 13 disclosure headers, `gosec`/`govulncheck` compile gates, network egress / FS jail enforcement |
| **Month 3–4** | **Observer Agent & Autonomous Shadow Loop** | Self-Healing Resilience | Art 14 5-min shadow testing, re-prompt event bus wiring, automated canary traffic promotion |
| **Month 5–6** | **Distributed State & Etcd Cluster Coordination** | Scalability & Fault Tolerance | Etcd state store adapter, distributed leader election, atomic gateway route swapping |
| **Month 7–8** | **Multi-Language Guest SDKs & Enterprise Gateway** | Rust/Python Guest & Edge Proxy | Rust/MicroPython WASM guests, IP token bucket rate limiting, automated cluster draining |

---

## EPIC: Security & AI Act Compliance (P0)

- [x] **BUG-023**: `X-Ghost-Ops-Synthesized` disclosure header never emitted (EU AI Act Art 13) (`DONE`)
- [x] **BUG-024**: Synthesized code never scanned with `gosec`/`govulncheck` (EU AI Act Art 15) (`DONE`)
- [x] **BUG-026**: Audit logger emits raw `details` map without secret masking — STANDARDS §2.3 violation (`DONE`)
- [x] **BUG-050**: `ai-skills.json` declares mandatory human review but no PR template / CODEOWNERS enforces it (`DONE`)
- [x] **TASK-1010.1**: Implement `X-Ghost-Ops-Synthesized` header middleware for EU AI Act Art 13 (`DONE`)
- [x] **TASK-1010.2**: Inject `// ghost-ops:generated` source tag into synthesized code (`DONE`)
- [x] **TASK-1011.1**: Wire `gosec` scanner into evolution engine compile pipeline (`DONE`)
- [x] **TASK-1011.2**: Wire `govulncheck` scanner into evolution engine compile pipeline (`DONE`)
- [x] **TASK-1011.3**: Update CI to enforce `gosec` and `govulncheck` strict passing (`DONE`)
- [ ] **TASK-1014.1**: Implement secret masking algorithm in audit logger (`OPEN`)
- [ ] **TASK-1014.2**: Implement JSONL persistence for audit logs (`OPEN`)
- [ ] **TASK-1014.3**: Implement log rotation policy for JSONL audit logs (`OPEN`)
- [ ] **TASK-1201.1**: Perform OWASP LLM Top 10 security audit: Hardcode secret scans (`OPEN`)
- [ ] **TASK-1201.2**: Perform OWASP LLM Top 10 security audit: Prompt injection filters (`OPEN`)
- [ ] **TASK-1202.1**: Audit Ghost Ops cross-repository dependencies and API schemas (`OPEN`)
- [ ] **TASK-1202.2**: Document identified cross-repo boundaries and update deployment manifests (`OPEN`)
- [ ] **TASK-708.1**: Design Secret Management integration architecture (Vault/AWS SM) (`OPEN`)
- [ ] **TASK-708.2**: Implement Secret Management integration client for Ghost Ops (`OPEN`)

## EPIC: Capability-Based Security (P0)

- [ ] **BUG-021**: Network egress capability never enforced in `rpc()` host fn — `CheckNetworkEgress` is dead code (`OPEN`)
- [ ] **BUG-022**: FS jail check decorative; only `wazero.WithDirMount` is wired — `CheckFSJail` is dead code (`OPEN`)
- [ ] **TASK-702.1**: Implement network egress policy validation logic (`OPEN`)
- [ ] **TASK-702.2**: Enforce network egress policies via `CheckNetworkEgress` in `rpc()` host function (`OPEN`)
- [ ] **TASK-703.1**: Implement robust FS jail validation logic for LoadModule (`OPEN`)
- [ ] **TASK-703.2**: Implement new `read_file` host function gated strictly by `CheckFSJail` (`OPEN`)

## EPIC: Observer Agent & Autonomous Feedback Loop (P1)

- [ ] **BUG-027**: Shadow-mode timer absent — no Article 14 ≥5-min gate; promotion is instant or never (`OPEN`)
- [ ] **BUG-028**: ZHO loop is broken — `EventRePromptRequired` has no subscriber; no re-evolve happens (`OPEN`)
- [ ] **BUG-045**: Health check unloads unhealthy services without emitting `EventRePromptRequired` — feedback edge dead on this side too (`OPEN`)
- [x] **BUG-054**: `EventBus.Publish()` silently drops events when subscriber buffer full — re-prompts vanish under load (`DONE`)
- [x] **BUG-055**: `EventBus.Subscribe()` ignores `ctx` — no way to cancel; channel leaks if subscriber goroutine exits (`DONE`)
- [ ] **BUG-040**: Reconcile loop returns `(true, nil)` on evolve failure — silent success (`OPEN`)
- [ ] **TASK-1012.1**: Implement shadow-mode timer (EU AI Act Art 14 compliance) (`OPEN`)
- [ ] **TASK-1012.2**: Implement shadow-mode traffic comparator and telemetry observer (`OPEN`)
- [ ] **TASK-1012.3**: Implement shadow-mode auto-promote logic upon timer completion (`OPEN`)
- [ ] **TASK-1013.1**: Implement Re-prompt subscriber to close ZHO feedback edge in registry (`OPEN`)
- [ ] **TASK-400.1**: Write Architectural Decision Record (ADR) for Autonomous Feedback Loop (`OPEN`)
- [ ] **TASK-400.4.1**: Define Re-prompt event payload strict schema (`OPEN`)
- [ ] **TASK-400.4.2**: Emit Re-prompt event payload upon health-check failure/purge (`OPEN`)
- [ ] **TASK-1019.1**: Instrument reconcile error visibility metrics (`reconcile_errors_total{stage}`) (`OPEN`)
- [ ] **TASK-1019.2**: Add OpenTelemetry trace correlation for reconcile errors (`OPEN`)

## EPIC: LLM Resilience (P1)

- [ ] **BUG-025**: OpenAI error path leaks request body (and embedded prompts) into stderr logs (`OPEN`)
- [ ] **BUG-029**: LLM provider has no per-request timeout — stalled provider starves the scheduler (`OPEN`)
- [ ] **BUG-042**: `compiler.go` inherits the host's full env into `go build` — `OPENAI_API_KEY` etc. leak into subprocess (`OPEN`)
- [ ] **BUG-043**: LLM cache "eviction" is map-iteration order (random), not LRU as comment claims (`OPEN`)
- [ ] **TASK-1015.1**: Implement strict per-request timeouts in LLM provider clients (`OPEN`)
- [ ] **TASK-1015.2**: Implement sanitized error paths for LLM provider clients (`OPEN`)
- [ ] **TASK-1017.1**: Design LRU cache data structure for LLM provider (`OPEN`)
- [ ] **TASK-1017.2**: Replace random eviction cache with strict LRU cache (`OPEN`)
- [ ] **TASK-1018.1**: Isolate and sanitize environment variables passed to `go build` subprocess (`OPEN`)

## EPIC: Stakeholder Demo (P2)

- [ ] **BUG-060**: `examples/blueprints/hello-compiler.json` is the only example blueprint — too thin for stakeholder demo (`OPEN`)
- [ ] **BUG-061**: No `examples/blueprints/` schema validation in `intent.NewFileIntentSource` — bad JSON crashes startup (`OPEN`)
- [ ] **TASK-1100.1**: Implement `make demo` Makefile target (`OPEN`)
- [ ] **TASK-1100.2**: Ensure `make demo` script boots Ghost Ops API, Gateway, and 3 example services (`OPEN`)
- [ ] **TASK-1101.1**: Author `examples/blueprints/demo/greeter.json` (`OPEN`)
- [ ] **TASK-1101.2**: Author `examples/blueprints/demo/summarizer.json` (`OPEN`)
- [ ] **TASK-1101.3**: Author `examples/blueprints/demo/slow-service.json` (`OPEN`)
- [ ] **TASK-1102**: Implement `/_demo/inject-latency` admin debug endpoint (`OPEN`)
- [ ] **TASK-1103**: Implement `/_demo/inject-error-rate` admin debug endpoint (`OPEN`)
- [ ] **TASK-1107**: Implement `/_demo/probe-egress` and `/_demo/probe-fs` admin debug endpoints (`OPEN`)
- [ ] **TASK-1104**: Write `examples/demo/walkthrough.sh` — scripted demo running all 9 demo steps (`OPEN`)
- [ ] **TASK-1105**: Record and publish pre-recorded 5-min demo video for v0.6.0 release (`OPEN`)
- [ ] **TASK-1108**: Implement `ghost-ops audit tail` CLI subcommand (`OPEN`)
- [ ] **TASK-1109**: Implement `ghost-ops service inspect --show-source` CLI flag (`OPEN`)
- [ ] **TASK-1110**: Implement `ghost-ops service list --watch` streaming CLI table (`OPEN`)

## EPIC: State Management (P2)

- [ ] **BUG-056**: `JSONFileStore` re-reads + re-parses the entire file on every read API — O(file) per Get/List under lock (`OPEN`)
- [ ] **BUG-044**: `Audit()` events not persisted — `slog` only; no append-only file, ISO 8.4 evidence missing (`OPEN`)
- [ ] **TASK-800.1**: Write Etcd integration strategy ADR (`OPEN`)
- [ ] **TASK-801.1**: Implement Etcd client connection wrapper (`OPEN`)
- [ ] **TASK-801.2**: Guard Etcd client setup behind `STORE_BACKEND=etcd` CLI flag (`OPEN`)
- [ ] **TASK-802.1**: Implement Etcd statestore adapter CRUD logic for ServiceRecords (`OPEN`)
- [ ] **TASK-802.2**: Implement Etcd statestore adapter CRUD logic for dynamic Configs (`OPEN`)
- [ ] **TASK-803.1**: Implement distributed leader election using Etcd leases (`OPEN`)
- [ ] **TASK-803.2**: Restrict reconcile execution to current elected leader node (`OPEN`)
- [ ] **TASK-804**: Design and implement state synchronization protocol (`OPEN`)
- [ ] **TASK-805**: Create partition tolerance test suite for distributed state (`OPEN`)
- [ ] **TASK-807**: Implement node auto-discovery mechanisms (`OPEN`)
- [ ] **TASK-808**: Implement graceful node draining during shutdown (`OPEN`)
- [ ] **TASK-1003**: Implement configurable log retention TTL for Audit Logs (`OPEN`)
- [ ] **TASK-1006.1**: Analyze wazero memory pool allocation bottlenecks (`OPEN`)
- [ ] **TASK-1006.2**: Optimize Wazero memory pool object recycling (`OPEN`)

## EPIC: HTTP Gateway (P3)

- [ ] **TASK-908.1**: Implement Gateway rate limiting token bucket logic (`OPEN`)
- [ ] **TASK-908.2**: Integrate per-IP token bucket limits into Gateway router (`OPEN`)
- [ ] **TASK-902**: Implement Gateway dynamic route reconfiguration via atomic pointer swap (`OPEN`)
- [ ] **TASK-906**: Write and execute Gateway load test to establish <2 ms p99 overhead baseline (`OPEN`)
- [ ] **TASK-901.1**: Document and define JSON schema for routing-rules format (`OPEN`)
- [ ] **TASK-903.1**: Implement Blue/green deployment traffic splitting logic (`OPEN`)
- [ ] **TASK-903.2**: Configure default Blue/green traffic weights to 90/10 (`OPEN`)
- [ ] **TASK-907**: Implement WebSocket upgrade support in gateway (`OPEN`)

## EPIC: Multi-Language SDKs (P3)

- [ ] **TASK-600.1**: Write Rust Guest SDK design ADR (`OPEN`)
- [ ] **TASK-605.1**: Write Python (WASM) Guest SDK design ADR (CPython vs MicroPython) (`OPEN`)
- [ ] **TASK-601.1**: Implement Rust Guest SDK host-ABI bindings (`OPEN`)
- [ ] **TASK-601.2**: Implement Rust Guest SDK struct abstractions (`OPEN`)
- [ ] **TASK-602**: Implement Rust Guest SDK structured logger (`OPEN`)
- [ ] **TASK-603.1**: Implement Rust compiler logic in evolution engine (`OPEN`)
- [ ] **TASK-603.2**: Handle cargo manifest generation for dynamic Rust compilation (`OPEN`)
- [ ] **TASK-604**: Write test suite for Rust compiler evolution engine (`OPEN`)
- [ ] **TASK-606.1**: Implement Python Guest SDK base (MicroPython wrapper) (`OPEN`)
- [ ] **TASK-607**: Implement Python execution logic in evolution engine (`OPEN`)
- [ ] **TASK-608.1**: Update demo blueprints to include Rust examples (`OPEN`)
- [ ] **TASK-608.2**: Update demo blueprints to include Python examples (`OPEN`)
- [ ] **TASK-609**: Implement cross-language SDK interop integration test suite (`OPEN`)

## EPIC: General / Tech Debt (P3)

- [ ] **BUG-031**: 2.9 MB `ghost-dev` binary committed at repo root (`OPEN`)
- [ ] **BUG-037**: Version drift — README says v0.5.0, VERSION says 0.5.1, `.system_state` says 0.5.2 (`OPEN`)
- [ ] **BUG-041**: `nextCommand()` truncates method names silently when caller buffer < name length (`OPEN`)
- [ ] **TASK-1020.1**: Update Dockerfile to run as non-root user (`OPEN`)
- [ ] **TASK-1020.2**: Update Dockerfile with `EXPOSE`, `HEALTHCHECK`, and multi-arch build tags (`OPEN`)
- [ ] **TASK-209.1**: Write comprehensive Ghost Ops Cluster Setup Guide (`OPEN`)
- [ ] **TASK-209.2**: Include Etcd and Gateway deployment manifests in Setup Guide (`OPEN`)
