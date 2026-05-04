# Ghost Ops — Backlog (THE Single Source of Truth)

> **One file. Everything is here.** Vision · DoD · Bugs · Tasks · Demo · 30-day
> plan · Vaulted history. Read top-to-bottom: highest priority first, nice-to-haves
> last. Anything not in this file is not work. The only sibling is
> [`STANDARDS.md`](./STANDARDS.md) (engineering rules — HOW, not WHAT).
> **Last reordered:** 2026-05-04.

---

## §1 Vision — Zero-Human Operations

**One-liner:** Code that writes itself, runs itself, and fixes itself.

A running Ghost Ops needs no human in the inner loop:

1. A human declares **intent** in plain English.
2. The system **synthesizes** Go source via an LLM.
3. The system **compiles** to WASM and runs it in a wazero sandbox.
4. The system **observes** runtime metrics (latency, errors).
5. When metrics breach SLOs, the system **re-prompts**, evolves, **shadow-tests**,
   and **hot-swaps** without dropping traffic.
6. The previous version is **discarded**. Source is ephemeral.

A human is needed only for: original intent, EU AI Act Article 14 high-risk
approvals, and prompt/standards reviews.

### The Loop (one diagram)

```
   Intent ──► Evolve ──► Compile ──► Sandbox ──► Observer ──► Re-Prompt ──┐
   (English)  (LLM)      (Go→WASM)   (Wazero)    (p99,err)    (Event Bus) │
      ▲          │            │           │                                │
      │          ▼            ▼           ▼                                │
      │     gosec/govuln  Capability  Capability                           │
      │     (Art 15)      enforcement enforcement                          │
      │                   (network)   (FS jail)                            │
      └────────────────────────────────────────────────────────────────────┘
                Audit Log (Art 13/17, masked secrets, JSONL)
```

### Pipeline Laws (non-negotiable)

1. **Latest stable only.** No backwards-compat shims.
2. **Ephemeral source.** Generated Go is compiled, hashed, then deleted.
3. **Simple-first.** Features serving <1% of cases are deferred.
4. **Adversarial Triad.** Every PR: 95% test, 0 lint, O(n) or better,
   `gosec`+`govulncheck` clean. Enforced in CI.
5. **EU AI Act controls.** Articles 13/14/15/17 wired into runtime, not docs.

### Component Map

| Package | Responsibility |
|---|---|
| `pkg/intent` | blueprint source (file → queue later) |
| `pkg/evolution` | LLM call + Go→WASM compile + scanner gate |
| `pkg/runtime` | wazero host + capability enforcement |
| `pkg/registry` | reconcile, deploy, shadow-mode promote |
| `pkg/observer` | metric & log thresholds, anomaly detection |
| `pkg/event` | pub/sub bus |
| `pkg/store` | JSON file → Etcd later (single SSOT for cluster state) |
| `pkg/api` | REST surface (blueprints + invocations) |
| `pkg/gateway` | L7 router (path/header) |
| `pkg/llm` | provider abstraction + cache |
| `pkg/logging` | structured logs + audit trail |
| `pkg/resilience` | retry, circuit breaker, rate limiter |
| `pkg/sdk/guest` | guest-side ABI stubs |
| `pkg/telemetry` | OpenTelemetry traces + in-memory metrics |

### Roadmap (concise)

| Phase | Timing | Theme | Exit gate |
|---|---|---|---|
| 0 | Done | Hygiene, lint, tests | CI green, structured logs |
| **1** | **Now → 2026-06-02** | **Self-healing loop** | **§5 Demo runs end-to-end** |
| 2 | Q3 2026 | Distribution + security | Etcd leader election, gosec clean |
| 3 | Q4 2026 | Multi-language guests | Go + Rust + Python via same ABI |
| 4 | 2027 | Full ZHO autonomy | EU AI Act audit pass |

---

## §2 Definition of Done — Phase-1

The system can run §5 (Demo) end-to-end with no human touch between steps 2 and 9, and:

- [ ] **DoD-1** Capability enforcement (network egress + FS jail) actually rejects
      unauthorized calls at the wazero host boundary. *(Closes BUG-021, BUG-022.)*
- [ ] **DoD-2** AI-synthesized responses carry `X-Ghost-Ops-Synthesized: true`
      header (Art 13). *(Closes BUG-023.)*
- [ ] **DoD-3** Synthesized code passes `gosec` + `govulncheck` before WASM
      compile (Art 15). *(Closes BUG-024.)*
- [ ] **DoD-4** Shadow mode holds new version ≥5 minutes before promotion (Art 14).
      *(Closes BUG-027.)*
- [ ] **DoD-5** `EventRePromptRequired` triggers re-evolve of the offending service.
      *(Closes BUG-028, BUG-045.)*
- [ ] **DoD-6** Audit log persists every synthesis with masked secrets (Art 13/17).
      *(Closes BUG-026, BUG-044.)*
- [ ] **DoD-7** CI runs the full Adversarial Triad and blocks merge on any
      failure. *(Closes BUG-030.)*

---

## §3 P0 — Bugs (must fix before claiming Phase-1)

### Legend
- **ID:** `BUG-NNN` (defect) · `TASK-NNN` (feature) · `EPIC-NNN` (parent).
- **Status:** `OPEN` · `IN-PROGRESS` · `DONE` · `BLOCKED`.
- **Day:** target sprint day (D1..D30) — see §6.

### CRITICAL

| ID | Title | File | Day | Status |
|---|---|---|---|---|
| BUG-021 | Network egress capability never enforced in `rpc()` host fn — `CheckNetworkEgress` is dead code | `pkg/runtime/wazero_host.go:185` | D8 | OPEN |
| BUG-022 | FS jail check decorative; only `wazero.WithDirMount` is wired — `CheckFSJail` is dead code | `pkg/runtime/wazero_host.go:480` | D9 | OPEN |
| BUG-023 | `X-Ghost-Ops-Synthesized` disclosure header never emitted (EU AI Act Art 13) | `pkg/api/server.go:147` | D10 | OPEN |
| BUG-024 | Synthesized code never scanned with `gosec`/`govulncheck` (EU AI Act Art 15) | `pkg/evolution/{ai_engine,compiler}.go` | D11 | OPEN |
| BUG-027 | Shadow-mode timer absent — no Article 14 ≥5-min gate; promotion is instant or never | `pkg/registry/registry.go:272` | D16 | OPEN |
| BUG-028 | ZHO loop is broken — `EventRePromptRequired` has no subscriber; no re-evolve happens | `pkg/observer/metric_observer.go` + `pkg/registry/` | D15 | OPEN |
| BUG-053 | `JSONFileStore.save()` is non-atomic — crash mid-write corrupts the entire state file | `pkg/store/json_store.go:89` | D6 | OPEN |
| BUG-054 | `EventBus.Publish()` silently drops events when subscriber buffer full — re-prompts vanish under load | `pkg/event/bus.go:39-43` | D15 | OPEN |

### HIGH

| ID | Title | File | Day | Status |
|---|---|---|---|---|
| BUG-025 | OpenAI error path leaks request body (and embedded prompts) into stderr logs | `pkg/llm/openai.go` | D13 | OPEN |
| BUG-026 | Audit logger emits raw `details` map without secret masking — STANDARDS §2.3 violation | `pkg/logging/audit.go:8` | D12 | OPEN |
| BUG-029 | LLM provider has no per-request timeout — stalled provider starves the scheduler | `pkg/llm/{openai,ollama}.go` | D13 | OPEN |
| BUG-030 | CI does not enforce STANDARDS §6 — missing `-race`, coverage gate, `gosec`, `govulncheck`, `gofmt -l` | `.github/workflows/ci.yml` | D2 | OPEN |
| BUG-042 | `compiler.go` inherits the host's full env into `go build` — `OPENAI_API_KEY` etc. leak into subprocess | `pkg/evolution/compiler.go:31` | D14 | OPEN |
| BUG-044 | `Audit()` events not persisted — `slog` only; no append-only file, ISO 8.4 evidence missing | `pkg/logging/audit.go` | D12 | OPEN |
| BUG-045 | Health check unloads unhealthy services without emitting `EventRePromptRequired` — feedback edge dead on this side too | `pkg/registry/registry.go:247` | D17 | OPEN |
| BUG-055 | `EventBus.Subscribe()` ignores `ctx` — no way to cancel; channel leaks if subscriber goroutine exits | `pkg/event/bus.go:50` | D15 | OPEN |
| BUG-056 | `JSONFileStore` re-reads + re-parses the entire file on every read API — O(file) per Get/List under lock | `pkg/store/json_store.go:97-115` | D18 | OPEN |
| BUG-057 | `JSONFileStore.save()` does not `fsync` — power-cut after `WriteFile` returns can lose the write | `pkg/store/json_store.go:89` | D6 | OPEN |

### MEDIUM

| ID | Title | File | Day | Status |
|---|---|---|---|---|
| BUG-031 | 2.9 MB `ghost-dev` binary committed at repo root | `/ghost-dev` | D1 | OPEN |
| BUG-032 | Duplicate Backlog SSOT — legacy `docs/planning/backlog.md` removed | n/a | D1 | DONE-this-PR |
| BUG-033 | Duplicate Roadmap stub `docs/planning/roadmap.md` (1 line) removed | n/a | D1 | DONE-this-PR |
| BUG-034 | Duplicate STANDARDS — `docs/rules/standards.md` shadow of SSOT removed | n/a | D1 | DONE-this-PR |
| BUG-035 | Stray planning scratch — `docs/planning/{active_tasks.txt, session_summary.md, map.md}` removed | n/a | D1 | DONE-this-PR |
| BUG-036 | Stray repo-root scratch — `plan.md`, `move_epic.py`, `test_task500.txt` removed | n/a | D1 | DONE-this-PR |
| BUG-037 | Version drift — README says v0.5.0, VERSION says 0.5.1, `.system_state` says 0.5.2 | README.md, VERSION | D1 | OPEN |
| BUG-038 | TASK-401 ("Multi-Language SDK Support") duplicates TASK-600..609 — collapsed into §8 | this file | D1 | DONE-this-PR |
| BUG-039 | TASK-901.1 missing — backlog jumped 901 → 901.2 → 901.3 — backfilled in §10 | this file | D1 | DONE-this-PR |
| BUG-040 | Reconcile loop returns `(true, nil)` on evolve failure — silent success | `pkg/registry/registry.go:107` | D18 | OPEN |
| BUG-041 | `nextCommand()` truncates method names silently when caller buffer < name length | `pkg/runtime/wazero_host.go:~268` | D19 | OPEN |
| BUG-043 | LLM cache "eviction" is map-iteration order (random), not LRU as comment claims | `pkg/llm/cache.go:54-59` | D14 | OPEN |
| BUG-051 | Dockerfile runs as root, `WORKDIR /root/`, no `EXPOSE`, no `HEALTHCHECK` | `Dockerfile` | D6 | OPEN |
| BUG-058 | `run.sh --sync` calls `npx skills add vercel-labs/agent-skills` — irrelevant to a Go project | `run.sh` | D1 | OPEN |
| BUG-059 | `examples/test-service/main.go` (173 bytes) is unreferenced and unbuilt | `examples/test-service/` | D1 | DONE-this-PR |

### LOW

| ID | Title | File | Day | Status |
|---|---|---|---|---|
| BUG-047 | `run.sh --sync/--skills` clauses are dead | run.sh | D1 | OPEN |
| BUG-048 | `docs/architecture/system_design.md` was a 4-line stub — folded into §1 | n/a | D1 | DONE-this-PR |
| BUG-049 | `docs/engineering/{README,conventions}.md` were 6-line stubs — deleted | n/a | D1 | DONE-this-PR |
| BUG-050 | `ai-skills.json` declares mandatory human review but no PR template / CODEOWNERS enforces it | `.github/` | D21 | OPEN |
| BUG-052 | `docs/release/metrics.md` was a single line — folded into §11 | n/a | D1 | DONE-this-PR |
| BUG-060 | `examples/blueprints/hello-compiler.json` is the only example blueprint — too thin for stakeholder demo | `examples/blueprints/` | D20 | OPEN |
| BUG-061 | No `examples/blueprints/` schema validation in `intent.NewFileIntentSource` — bad JSON crashes startup | `pkg/intent/file_source.go` | D19 | OPEN |
| BUG-062 | `Makefile` has no `audit`, `demo`, `gosec`, `govulncheck`, or `coverage` targets | `Makefile` | D2 | OPEN |
| BUG-064 | Multiple Markdown SSOT files (VISION, DEMO, SPRINT-30D, AUDIT-2026-05-04, ROADMAP, CHANGELOG) — folded into this file | docs/ | D1 | DONE-this-PR |

> **Bug-fix exit gate (Phase-1):** Every CRITICAL + HIGH closed. MEDIUMs may slip
> to Phase-2 with explicit reason recorded in §11.

---

## §4 P1 — Vision-critical features (Phase-1 must-have)

These clear the §2 DoD bullets. Must exist by Day 21.

| ID | Title | Closes | Day | Status |
|---|---|---|---|---|
| TASK-1010 | EU AI Act Art 13 — `X-Ghost-Ops-Synthesized` header + `// ghost-ops:generated` source tag | DoD-2 | D10 | OPEN |
| TASK-1011 | EU AI Act Art 15 — wire `gosec`+`govulncheck` into compile pipeline + CI | DoD-3, DoD-7 | D2, D11 | OPEN |
| TASK-1012 | EU AI Act Art 14 — shadow-mode timer + comparator + auto-promote | DoD-4 | D16 | OPEN |
| TASK-1013 | Re-prompt subscriber — close ZHO feedback edge in registry | DoD-5 | D15 | OPEN |
| TASK-1014 | Audit-log secret masking + JSONL persistence with rotation | DoD-6 | D12 | OPEN |
| TASK-702 | Enforce network egress policies via `CheckNetworkEgress` in `rpc()` | DoD-1 | D8 | OPEN |
| TASK-703 | Implement FS jails: validate at LoadModule + new `read_file` host fn gated by `CheckFSJail` | DoD-1 | D9 | OPEN |
| EPIC-701 | **De-vault** Capability-Based Security until 702+703 actually enforce | DoD-1 | D7 | DEMOTED |
| TASK-400 | Autonomous Feedback Loop Architecture (ADR) | DoD-5 | D3 | BLOCKED→D3 |
| TASK-400.4 | Re-prompt event payload schema + emit on health-check purge | DoD-5 | D17 | OPEN |
| TASK-1015 | LLM provider per-request timeout + sanitized error path | DoD-7 | D13 | OPEN |
| TASK-1019 | Reconcile error visibility (`reconcile_errors_total{stage}`) | DoD-7 | D18 | OPEN |

---

## §5 P2 — Demo & Pitch (so we can ship to stakeholders)

### 5.1 The 30-Second Pitch

> Today, when a service is too slow, an engineer is paged, reads logs, writes
> code, opens a PR, and waits for CI. That round trip is hours.
>
> Ghost Ops cuts the human out. You declare *intent* in English. The system
> writes the Go code, compiles it to WASM, runs it in a sandbox, watches its
> own latency, and — when SLOs slip — re-prompts itself for a faster version,
> tests the replacement in shadow mode, then hot-swaps. No PR. No deploy. No page.
>
> EU AI Act compliant. Security-jailed. Auditable. v0.6.0 today.

### 5.2 Live Demo — 9 Steps, ~5 Minutes

**Setup:**
```bash
git clone https://github.com/avinya-forge/ghost-ops && cd ghost-ops
make demo            # boots ghost-ops + injectors + 3 example blueprints
```

`make demo` (TASK-1100) starts:
- Ghost Ops on `:8080` with mock LLM (offline-safe).
- Latency injector at `:8081` (`/_demo/inject-*`).
- Three blueprints in `examples/blueprints/demo/` queued for evolution.

**Step 1 — Show the intent (10s)**
```bash
cat examples/blueprints/demo/greeter.json
```
```json
{
  "service_id": "greeter",
  "intent": "Return JSON {\"hello\":<name>} when invoked with method=greet, payload=<name>. p99<50ms.",
  "constraints": { "p99_ms": 50, "shadow_mode": false }
}
```
*"This is the only thing a human writes. No code, no Dockerfile."*

**Step 2 — Watch synthesis + deploy (30s)**
```bash
ghost-ops service list --watch
```
```
greeter  PENDING
greeter  SYNTHESIZED  v1  hash=ab12cd…  (gosec PASS, govulncheck PASS)
greeter  COMPILED     v1  wasm=8.4 KB
greeter  ACTIVE       v1
```
*"Article 15 — gosec and govulncheck ran before compile. Any HIGH finding would have blocked deployment."*

**Step 3 — Invoke + show Article 13 disclosure (15s)**
```bash
curl -i -X POST 'http://localhost:8080/services/greeter/invoke?method=greet' -d 'Stakeholder'
```
```
HTTP/1.1 200 OK
X-Ghost-Ops-Synthesized: true        ← EU AI Act Article 13
X-Ghost-Ops-Service-Version: 1
X-Ghost-Ops-Wasm-Hash: ab12cd…
{"hello":"Stakeholder"}
```
*"Every response from an AI-synthesized service ships with this header."*

**Step 4 — Show source tag (10s)**
```bash
ghost-ops service inspect greeter --show-source | head -1
```
```
// ghost-ops:generated  blueprint=greeter  v=1  hash=ab12cd…
```

**Step 5 — Prove the sandbox is real (30s)**
```bash
curl -X POST 'http://localhost:8080/_demo/probe-egress?service=greeter&target=evil-api'
# {"result":"denied","reason":"evil-api not in NetworkEgress allowlist"}

curl -X POST 'http://localhost:8080/_demo/probe-fs?service=greeter&path=/etc/passwd'
# {"result":"denied","reason":"/etc/passwd outside FSJails"}
```
*"The blueprint doesn't grant network or filesystem access. The host enforces it at the wazero boundary."*

**Step 6 — Inject a latency regression (15s)**
```bash
curl -X POST 'http://localhost:8081/_demo/inject-latency?service=greeter&p99_ms=600'
```
Within ~5s:
```
greeter  ACTIVE        v1  p99=620ms  ⚠ THRESHOLD BREACHED
greeter  RE-PROMPTING  v1  reason=p99_breach
```
*"The Observer just fired EventRePromptRequired. No human touched anything."*

**Step 7 — Watch the system evolve v2 (30s)**
```
greeter  SYNTHESIZED  v2  prompt_hint="current p99=620ms, halve it"
greeter  COMPILED     v2  wasm=7.9 KB  (scans PASS)
greeter  SHADOW       v2  ← runs alongside v1 with traffic copy
```

**Step 8 — Auto-promote after shadow window (15s)**
Demo uses `GHOST_OPS_SHADOW_DURATION=15s` so the gate clears quickly.
```
greeter  SHADOW     v2  shadow_age=15s  shadow_p99=42ms  active_p99=620ms
greeter  PROMOTING  v2  ← shadow beat active, swap
greeter  ACTIVE     v2  ← v1 unloaded
```
```bash
curl -s 'http://localhost:8080/services/greeter/invoke?method=greet' -d 'Stakeholder'
# {"hello":"Stakeholder"}    ← still works, atomic swap
```
*"Zero requests dropped. v1 unloaded. The system fixed itself."*

**Step 9 — Show the audit trail (15s)**
```bash
ghost-ops audit tail --service greeter
```
```
synthesize    v=1  prompt_hash=8a1f…  secrets=***
scan_pass     v=1  gosec=0  govulncheck=0
deploy        v=1  state=ACTIVE
metric_breach service=greeter p99_ms=620
reprompt      v_old=1  hint="p99=620ms, halve it"
synthesize    v=2  prompt_hash=9b2e…  secrets=***
scan_pass     v=2  gosec=0  govulncheck=0
deploy        v=2  state=SHADOW
promote       v_old=1  v_new=2  shadow_p99=42ms  active_p99=620ms
unload        v=1
```
*"Every step in append-only audit log. Secrets masked. ISO/IEC 42001 8.4 compliant."*

### 5.3 Demo Tasks (build the demo so it actually runs)

| ID | Title | Day | Status |
|---|---|---|---|
| TASK-1100 | `make demo` target — single command boots Ghost Ops + injectors + 3 example services | D20 | OPEN |
| TASK-1101 | `examples/blueprints/demo/{greeter,summarizer,slow-service}.json` | D20 | OPEN |
| TASK-1102 | `/_demo/inject-latency` admin endpoint (debug-build only) | D20 | OPEN |
| TASK-1103 | `/_demo/inject-error-rate` admin endpoint (debug-build only) | D20 | OPEN |
| TASK-1107 | `/_demo/probe-egress` and `/_demo/probe-fs` admin endpoints (validate sandbox) | D20 | OPEN |
| TASK-1104 | `examples/demo/walkthrough.sh` — scripted demo running all 9 steps | D29 | OPEN |
| TASK-1105 | Pre-recorded 5-min demo video attached to v0.6.0 release | D30 | OPEN |
| TASK-1108 | `ghost-ops audit tail` CLI subcommand | D29 | OPEN |
| TASK-1109 | `ghost-ops service inspect --show-source` flag (with provenance line) | D10 | OPEN |
| TASK-1110 | `ghost-ops service list --watch` streaming CLI | D29 | OPEN |

### 5.4 Demo File Inventory

```
examples/
├── blueprints/
│   ├── hello-compiler.json          (existing — kept for compiler engine demo)
│   └── demo/
│       ├── greeter.json             ← star of the show
│       ├── summarizer.json          ← multi-service hop
│       └── slow-service.json        ← intentionally breaches SLO for chaos demo
└── demo/
    └── walkthrough.sh               ← all 9 steps, scripted, idempotent
```

### 5.5 Demo Risk Register

| Risk | Mitigation |
|---|---|
| LLM provider latency / outage | Demo uses `-llm mock` by default; deterministic ms-scale synthesis |
| Long shadow window kills pacing | `GHOST_OPS_SHADOW_DURATION=15s` env override (debug-build only) |
| Audit dump unreadable on screen | `audit tail --service greeter` filters to one service |
| Live coding fear | Always run via `walkthrough.sh`, never type live |
| Network/firewall blocks | Whole demo runs on `localhost`, no internet needed |

---

## §6 30-Day Day-by-Day (one PR per day)

| Day | Theme | Closes |
|---|---|---|
| D1 | Repo cleanup, doc flatten | BUG-031..039, 048..049, 052, 058..059, 064 |
| D2 | CI Adversarial Triad live | BUG-030, BUG-062, TASK-1011 (CI half) |
| D3 | ADR — TASK-400 feedback loop architecture | TASK-400 |
| D4 | ADR — TASK-800 Etcd vs Redis | TASK-800 |
| D5 | ADRs — Rust ABI (TASK-600), Python WASM (TASK-605) | TASK-600, TASK-605 |
| D6 | Hardened Dockerfile + atomic store writes + fsync | BUG-051, BUG-053, BUG-057, TASK-1020 |
| D7 | Backlog reconciliation, EPIC-701 demote, TASK-901.1 backfill | BUG-038, BUG-039, EPIC-701 status |
| D8 | Wire `CheckNetworkEgress` into `rpc()` | BUG-021, TASK-702 |
| D9 | Wire FS jail + new `read_file` host fn | BUG-022, TASK-703 |
| D10 | Article 13 header + source tag + `inspect --show-source` | BUG-023, TASK-1010, TASK-1109 |
| D11 | Article 15 — gosec/govulncheck in synthesis | BUG-024, TASK-1011 |
| D12 | Audit-log secret masking + JSONL store | BUG-026, BUG-044, TASK-1014 |
| D13 | LLM timeout + sanitized errors | BUG-025, BUG-029, TASK-1015 |
| D14 | LRU cache + compiler env hygiene | BUG-042, BUG-043, TASK-1017, TASK-1018 |
| D15 | Re-prompt subscriber + event-bus dropped-event metric + ctx-aware Subscribe | BUG-028, BUG-054, BUG-055, TASK-1013 |
| D16 | Article 14 shadow-mode timer | BUG-027, TASK-1012 |
| D17 | Health-check → re-prompt edge | BUG-045, TASK-400.4 |
| D18 | Reconcile error visibility + store read cache | BUG-040, BUG-056, TASK-1019 |
| D19 | Buffer truncation hardening + blueprint schema validation | BUG-041, BUG-061 |
| D20 | Demo scaffold (`make demo`, blueprints, injector + probe endpoints) | TASK-1100..1103, TASK-1107 |
| D21 | PR template + CODEOWNERS + ADR linter | BUG-050 |
| D22 | Etcd client setup behind flag | TASK-801 |
| D23 | Etcd statestore adapter | TASK-802 |
| D24 | Distributed leader election | TASK-803 |
| D25 | Gateway rate limiting | TASK-908 |
| D26 | Gateway dynamic route reconfig | TASK-902 |
| D27 | Gateway load test (<2 ms p99) | TASK-906 |
| D28 | Race + leak hunt (`go test -race -count=20`, goleak) | hardening |
| D29 | Demo walkthrough script + CLI polish + doc freeze | TASK-1104, TASK-1108, TASK-1110 |
| D30 | v0.6.0 tag + recorded demo + pitch one-pager | TASK-1105, §1 polish |

### Daily Discipline (every day)

1. Pull main; branch `<type>/<task-or-bug-id>`.
2. Open draft PR with the day's exit gate as the checklist.
3. Implement <300 LOC + tests.
4. `make lint && make test && make audit` (D2 onwards).
5. Push; CI must be green.
6. Mark this BACKLOG row `DONE` with date.
7. Merge before EOD.

---

## §7 P3 — Phase-2 (Distribution + Hardening)

| ID | Title | Day | Status |
|---|---|---|---|
| TASK-800 | Etcd integration strategy ADR | D4 | BLOCKED→D4 |
| TASK-801 | Etcd client setup behind `STORE_BACKEND=etcd` flag | D22 | OPEN |
| TASK-802 | Etcd statestore adapter (CRUD against existing protocol) | D23 | OPEN |
| TASK-803 | Distributed leader election (only leader runs reconcile) | D24 | OPEN |
| TASK-908 | Gateway rate limiting (per-IP token bucket) | D25 | OPEN |
| TASK-902 | Gateway dynamic route reconfiguration (atomic pointer swap) | D26 | OPEN |
| TASK-906 | Gateway load test — establish <2 ms p99 overhead baseline | D27 | OPEN |
| TASK-1018 | Compiler env hygiene — minimal env to `go build` subprocess | D14 | OPEN |
| TASK-1017 | LRU cache for LLM provider (replace random eviction) | D14 | OPEN |
| TASK-1020 | Hardened Dockerfile (non-root, EXPOSE, HEALTHCHECK, multi-arch) | D6 | OPEN |
| TASK-708 | Secret Management integration (Vault/AWS SM) | post-D30 | OPEN |
| TASK-804 | State synchronization protocol | post-D30 | OPEN |
| TASK-805 | Partition tolerance testing | post-D30 | OPEN |
| TASK-807 | Node auto-discovery | post-D30 | OPEN |
| TASK-808 | Graceful node draining | post-D30 | OPEN |
| TASK-704 | Automated vulnerability scanning | n/a | DUPLICATE-of-TASK-1011 |

---

## §8 P4 — Phase-3 (Multi-Language Guests — design only this month)

| ID | Title | Day | Status |
|---|---|---|---|
| TASK-600 | Rust Guest SDK design ADR | D5 | BLOCKED→D5 |
| TASK-605 | Python (WASM) Guest SDK design ADR (CPython vs MicroPython) | D5 | BLOCKED→D5 |
| TASK-401 | (DELETE) duplicates TASK-600..609 | — | DUPLICATE |
| TASK-601 | Implement Rust Guest SDK base | Q3 | OPEN |
| TASK-602 | Rust Guest SDK logger | Q3 | OPEN |
| TASK-603 | Rust compiler evolution engine | Q3 | OPEN |
| TASK-604 | Test Rust compiler engine | Q3 | OPEN |
| TASK-606 | Python Guest SDK base | Q3 | OPEN |
| TASK-607 | Python evolution engine | Q3 | OPEN |
| TASK-608 | Update examples with Rust/Python | Q3 | OPEN |
| TASK-609 | Cross-language interop testing | Q3 | OPEN |

---

## §9 P5 — Good-to-haves (post-Phase-1)

| ID | Title | Day | Status |
|---|---|---|---|
| TASK-901.1 | (BACKFILL) document/define routing-rules format | D7 | OPEN |
| TASK-903 | Blue/green deployment (traffic weights 90/10) | Q3 | OPEN |
| TASK-907 | WebSocket support in gateway | Q3 | OPEN |
| TASK-1003 | Configurable log retention TTL | Q3 | OPEN |
| TASK-1004 | Log filtering UI | Q4 | BLOCKED |
| TASK-1005 | OAuth2 external auth providers (HIGH-RISK) | Q4 | BLOCKED |
| TASK-1006 | Wazero memory pool optimization | Q4 | OPEN |
| TASK-209 | Cluster setup guide | Q3 | OPEN |
| TASK-709 | Security architecture guide | Q3 | BLOCKED |
| TASK-909 | Routing configuration guide | Q3 | BLOCKED |

---

## §10 New Bugs / Change Requests

When you find a new bug, add it as the next `BUG-NNN` directly under the right
severity heading in §3. **Never** open a parallel doc.

When you propose a new task, add it under the right priority section (§4..§9)
with a Day target. If it slips past D30, move it to §9.

---

## §11 Vaulted (shipped — kept brief)

| ID | Title | Shipped | Notes |
|---|---|---|---|
| EPIC-500 | Observer Agent & metric thresholds | v0.5.0 | Partial — re-prompt edge dead, see BUG-028 |
| EPIC-901 | HTTP Gateway path + header routing | v0.6.0 | Minimal — see §9 for missing pieces |
| TASK-320 | Sidecar proxy pattern | v0.5.0 | |
| TASK-321 | mTLS between services | v0.5.0 | |
| TASK-331 | Resource-aware scheduling (bin-pack) | v0.5.0 | |
| TASK-1000..1002 | Log streaming + anomaly detection prompt | v0.5.0 | |
| TASK-1200 | Evolution engine priority queue | v0.4.0 | |
| TASK-1358 | Remove `/cluster/health` API | v0.5.0 | |
| BUG-001..020 | First bug-hunt sprint (17 fixed, 3 confirmed non-bugs) | 2026-04-22 | |

> **EPIC-701** (Capability-Based Security) was previously listed as vaulted at
> v0.6.0 but **demoted** here — see BUG-021/022. It will re-vault once D8+D9 land.

### Detailed historical release notes (was `docs/release/release-notes.md`)

**v0.5.0** — Observer Agent + metric thresholds (EPIC-500), Sidecar Proxy +
mTLS (TASK-320/321), Resource-aware scheduler (TASK-331), Log streaming for AI
analysis (TASK-1000..1002), Removed undocumented `/cluster/health` (TASK-1358).

**v0.4.0** — Architecture diagrams, code-gen evals, OTLP exporter, trace-log
correlation, evolution priority queue, prompt-injection defenses, security
chaos testing, audit logging for state changes, cluster health dashboard data.

**v0.3.0** — Service Mesh Lite (Retry / Circuit Breaker / Rate Limit), Redis
distributed lock (`SET NX`), OpenTelemetry tracing middleware, in-memory event
bus, API validation middleware (`MaxBytes`, `ContentType`, `Recover`),
`ghost-dev` hot-reload, pre-commit hooks, end-to-end integration tests, async
WASM module instantiation, LRU module cache, health-check active-state
verification, `POST /services/{id}/invoke`, CLI subcommands (`init`, `service
list/inspect/logs`, `config show`), Wazero CPU+memory limits, secure log
capture, `AppError` standardization, Viper-backed config with hot-reload,
OpenAPI 3.0 spec, LLM token-usage metrics, dependency graph cycle detection,
`gosec` initial pass, RuntimeHost benchmarks (~0.08 ms overhead).

(Earlier version detail is preserved in git history; not material to current
priorities.)

---

*This file is the only Markdown SSOT for execution. The companion is
`STANDARDS.md` (engineering rules). README links here. Anything else is noise.*
