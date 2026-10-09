# Ghost Ops — Backlog (Execution SSOT)

> **The execution ladder.** Bugs first (priority-ordered), then features, then
> demo, then 30-day plan, then vaulted history.
> **North star (the why):** [`VISION.md`](./VISION.md).
> **Engineering rules (the how):** [`STANDARDS.md`](./STANDARDS.md).
> Anything not in this file is not work.

---

## Security & AI Act Compliance

- [ ] **TASK-1010**: EU AI Act Art 13 — `X-Ghost-Ops-Synthesized` header + `// ghost-ops:generated` source tag (`OPEN`)
- [ ] **TASK-1011**: EU AI Act Art 15 — wire `gosec`+`govulncheck` into compile pipeline + CI (`OPEN`)
- [ ] **TASK-1014**: Audit-log secret masking + JSONL persistence with rotation (`OPEN`)
- [ ] **TASK-1201**: OWASP LLM Security Audit: Hardcode secret scans, prompt injection filters (`OPEN`)
- [ ] **TASK-1202**: Agent Reach Cross-Repository Dependency Audit: Identify cross-repo dependencies and API schema changes (`OPEN`)
- [ ] **BUG-023**: `X-Ghost-Ops-Synthesized` disclosure header never emitted (EU AI Act Art 13) (`OPEN`)
- [ ] **BUG-024**: Synthesized code never scanned with `gosec`/`govulncheck` (EU AI Act Art 15) (`OPEN`)
- [ ] **BUG-026**: Audit logger emits raw `details` map without secret masking — STANDARDS §2.3 violation (`OPEN`)
- [ ] **BUG-050**: `ai-skills.json` declares mandatory human review but no PR template / CODEOWNERS enforces it (`OPEN`)
- [ ] **TASK-708**: Secret Management integration (Vault/AWS SM) (`OPEN`)
- [ ] **TASK-709**: Security architecture guide (`BLOCKED`)

## Capability-Based Security

- [ ] **EPIC-701**: **De-vault** Capability-Based Security until 702+703 actually enforce (`DEMOTED`)
- [ ] **BUG-021**: Network egress capability never enforced in `rpc()` host fn — `CheckNetworkEgress` is dead code (`OPEN`)
- [ ] **BUG-022**: FS jail check decorative; only `wazero.WithDirMount` is wired — `CheckFSJail` is dead code (`OPEN`)
- [ ] **TASK-702**: Enforce network egress policies via `CheckNetworkEgress` in `rpc()` (`OPEN`)
- [ ] **TASK-703**: Implement FS jails: validate at LoadModule + new `read_file` host fn gated by `CheckFSJail` (`OPEN`)

## Observer Agent & Autonomous Feedback Loop

- [ ] **BUG-027**: Shadow-mode timer absent — no Article 14 ≥5-min gate; promotion is instant or never (`OPEN`)
- [ ] **BUG-028**: ZHO loop is broken — `EventRePromptRequired` has no subscriber; no re-evolve happens (`OPEN`)
- [ ] **BUG-045**: Health check unloads unhealthy services without emitting `EventRePromptRequired` — feedback edge dead on this side too (`OPEN`)
- [ ] **TASK-1012**: EU AI Act Art 14 — shadow-mode timer + comparator + auto-promote (`OPEN`)
- [ ] **TASK-1013**: Re-prompt subscriber — close ZHO feedback edge in registry (`OPEN`)
- [ ] **TASK-400**: Autonomous Feedback Loop Architecture (ADR) (`BLOCKED→D3`)
- [ ] **TASK-400.4**: Re-prompt event payload schema + emit on health-check purge (`OPEN`)
- [ ] **BUG-054**: `EventBus.Publish()` silently drops events when subscriber buffer full — re-prompts vanish under load (`OPEN`)
- [ ] **BUG-055**: `EventBus.Subscribe()` ignores `ctx` — no way to cancel; channel leaks if subscriber goroutine exits (`OPEN`)
- [ ] **BUG-040**: Reconcile loop returns `(true, nil)` on evolve failure — silent success (`OPEN`)
- [ ] **TASK-1019**: Reconcile error visibility (`reconcile_errors_total{stage}`) (`OPEN`)

## LLM Resilience

- [ ] **BUG-025**: OpenAI error path leaks request body (and embedded prompts) into stderr logs (`OPEN`)
- [ ] **BUG-029**: LLM provider has no per-request timeout — stalled provider starves the scheduler (`OPEN`)
- [ ] **BUG-042**: `compiler.go` inherits the host's full env into `go build` — `OPENAI_API_KEY` etc. leak into subprocess (`OPEN`)
- [ ] **BUG-043**: LLM cache "eviction" is map-iteration order (random), not LRU as comment claims (`OPEN`)
- [ ] **TASK-1015**: LLM provider per-request timeout + sanitized error path (`OPEN`)
- [ ] **TASK-1017**: LRU cache for LLM provider (replace random eviction) (`OPEN`)
- [ ] **TASK-1018**: Compiler env hygiene — minimal env to `go build` subprocess (`OPEN`)

## Stakeholder Demo

- [ ] **TASK-1100**: `make demo` target — single command boots Ghost Ops + injectors + 3 example services (`OPEN`)
- [ ] **TASK-1101**: `examples/blueprints/demo/{greeter,summarizer,slow-service}.json` (`OPEN`)
- [ ] **TASK-1102**: `/_demo/inject-latency` admin endpoint (debug-build only) (`OPEN`)
- [ ] **TASK-1103**: `/_demo/inject-error-rate` admin endpoint (debug-build only) (`OPEN`)
- [ ] **TASK-1107**: `/_demo/probe-egress` and `/_demo/probe-fs` admin endpoints (validate sandbox) (`OPEN`)
- [ ] **TASK-1104**: `examples/demo/walkthrough.sh` — scripted demo running all 9 steps (`OPEN`)
- [ ] **TASK-1105**: Pre-recorded 5-min demo video attached to v0.6.0 release (`OPEN`)
- [ ] **TASK-1108**: `ghost-ops audit tail` CLI subcommand (`OPEN`)
- [ ] **TASK-1109**: `ghost-ops service inspect --show-source` flag (with provenance line) (`OPEN`)
- [ ] **TASK-1110**: `ghost-ops service list --watch` streaming CLI (`OPEN`)
- [ ] **BUG-060**: `examples/blueprints/hello-compiler.json` is the only example blueprint — too thin for stakeholder demo (`OPEN`)
- [ ] **BUG-061**: No `examples/blueprints/` schema validation in `intent.NewFileIntentSource` — bad JSON crashes startup (`OPEN`)

## State Management

- [ ] **TASK-800**: Etcd integration strategy ADR (`BLOCKED→D4`)
- [ ] **TASK-801**: Etcd client setup behind `STORE_BACKEND=etcd` flag (`OPEN`)
- [ ] **TASK-802**: Etcd statestore adapter (CRUD against existing protocol) (`OPEN`)
- [ ] **TASK-803**: Distributed leader election (only leader runs reconcile) (`OPEN`)
- [ ] **TASK-804**: State synchronization protocol (`OPEN`)
- [ ] **TASK-805**: Partition tolerance testing (`OPEN`)
- [ ] **TASK-807**: Node auto-discovery (`OPEN`)
- [ ] **TASK-808**: Graceful node draining (`OPEN`)
- [ ] **BUG-056**: `JSONFileStore` re-reads + re-parses the entire file on every read API — O(file) per Get/List under lock (`OPEN`)
- [ ] **BUG-044**: `Audit()` events not persisted — `slog` only; no append-only file, ISO 8.4 evidence missing (`OPEN`)
- [ ] **TASK-1003**: Configurable log retention TTL (`OPEN`)
- [ ] **TASK-1004**: Log filtering UI (`BLOCKED`)
- [ ] **TASK-1006**: Wazero memory pool optimization (`OPEN`)

## HTTP Gateway

- [ ] **TASK-908**: Gateway rate limiting (per-IP token bucket) (`OPEN`)
- [ ] **TASK-902**: Gateway dynamic route reconfiguration (atomic pointer swap) (`OPEN`)
- [ ] **TASK-906**: Gateway load test — establish <2 ms p99 overhead baseline (`OPEN`)
- [ ] **TASK-901.1**: (BACKFILL) document/define routing-rules format (`OPEN`)
- [ ] **TASK-903**: Blue/green deployment (traffic weights 90/10) (`OPEN`)
- [ ] **TASK-907**: WebSocket support in gateway (`OPEN`)
- [ ] **TASK-1005**: OAuth2 external auth providers (HIGH-RISK) (`BLOCKED`)
- [ ] **TASK-909**: Routing configuration guide (`BLOCKED`)

## Multi-Language SDKs

- [ ] **TASK-600**: Rust Guest SDK design ADR (`BLOCKED→D5`)
- [ ] **TASK-605**: Python (WASM) Guest SDK design ADR (CPython vs MicroPython) (`BLOCKED→D5`)
- [ ] **TASK-601**: Implement Rust Guest SDK base (`OPEN`)
- [ ] **TASK-602**: Rust Guest SDK logger (`OPEN`)
- [ ] **TASK-603**: Rust compiler evolution engine (`OPEN`)
- [ ] **TASK-604**: Test Rust compiler engine (`OPEN`)
- [ ] **TASK-606**: Python Guest SDK base (`OPEN`)
- [ ] **TASK-607**: Python evolution engine (`OPEN`)
- [ ] **TASK-608**: Update examples with Rust/Python (`OPEN`)
- [ ] **TASK-609**: Cross-language interop testing (`OPEN`)
- [ ] **TASK-401**: (DELETE) duplicates TASK-600..609 (`DUPLICATE`)

## General / Tech Debt

- [ ] **BUG-031**: 2.9 MB `ghost-dev` binary committed at repo root (`OPEN`)
- [ ] **BUG-037**: Version drift — README says v0.5.0, VERSION says 0.5.1, `.system_state` says 0.5.2 (`OPEN`)
- [ ] **BUG-041**: `nextCommand()` truncates method names silently when caller buffer < name length (`OPEN`)
- [ ] **TASK-1020**: Hardened Dockerfile (non-root, EXPOSE, HEALTHCHECK, multi-arch) (`OPEN`)
- [ ] **TASK-209**: Cluster setup guide (`OPEN`)
