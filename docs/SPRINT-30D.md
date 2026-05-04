# Ghost Ops — 30-Day Sprint Plan (2026-05-04 → 2026-06-02)

> **Cadence:** 1 day = 1 sprint. Each day ships a verifiable artefact (commit + green CI).
> **North Star:** Phase-1 MVP — *honest* self-healing loop with EU AI Act controls.
> **Companion:** `docs/AUDIT-2026-05-04.md` (bug log + alignment matrix).
> **Branch model:** Each sprint is one focused PR (`feat/<task-id>` or `fix/<bug-id>`).

---

## Plan Shape

- **Week 1 (Days 1–7): Hygiene + Unblock** — clean repo, write the four blocking ADRs,
  fix CI gates. *Goal: zero technical debt entering Week 2.*
- **Week 2 (Days 8–14): Wire EPIC-701 for real** — capability enforcement at the host
  boundary, AI Act Article 13 + 15. *Goal: revoke the false "vaulted" claim.*
- **Week 3 (Days 15–21): Close the ZHO Loop** — Article 14 shadow-mode timer,
  re-prompt subscriber, health-check feedback. *Goal: end-to-end metric→re-evolve demo.*
- **Week 4 (Days 22–30): Distribute + Demo** — Etcd adapter behind a flag, gateway
  hardening, e2e load test, release v0.6.0. *Goal: v0.6.0 cut + recorded demo.*

Each daily sprint is sized to <300 LOC + tests, finishable in one focused day.

---

## Week 1 — Hygiene + Unblock

### Day 1 (Mon 2026-05-04) — Repo cleanup
**Closes:** BUG-031, BUG-032, BUG-033, BUG-034, BUG-035, BUG-036, BUG-046, BUG-047, BUG-052
- `git rm --cached ghost-dev`; add to `.gitignore`.
- Delete `plan.md`, `move_epic.py`, `test_task500.txt`.
- Delete `docs/planning/{backlog.md,roadmap.md,active_tasks.txt,session_summary.md,map.md}`.
- Delete `docs/rules/standards.md` (content already in STANDARDS.md).
- Delete `examples/test-service/`.
- Strip `--sync`/`--skills` clauses from `run.sh`.
- README badge → v0.5.2; reconcile with VERSION (BUG-037).
- **Exit gate:** `git ls-files | wc -l` drops; `make build && make test` green.

### Day 2 — CI Adversarial Triad
**Closes:** BUG-030, BUG-024 (CI part)
- Add to `.github/workflows/ci.yml`: `go test -race -coverprofile=cov.out`, coverage gate ≥85% (ratchet later), `gofmt -l` non-zero check, `gosec ./...`, `govulncheck ./...`.
- Cache Go modules in CI.
- **Exit gate:** All gates green on a fresh PR.

### Day 3 — ADR-001: TASK-400 Feedback Loop Architecture
**Closes:** TASK-400 design block
- Write `docs/architecture/decisions.md` ADR: re-prompt subscriber pattern, event payload schema, shadow window, comparator. Token-impact: 4 (Claude Pro).
- **Exit gate:** ADR merged; TASK-400 status → `PENDING`.

### Day 4 — ADR-002: TASK-800 Etcd vs Redis
- ADR comparing latency, consensus model, ops complexity. Pick one. Token-impact: 4.
- **Exit gate:** TASK-800 → `PENDING`.

### Day 5 — ADR-003: TASK-600 Rust ABI + ADR-004: TASK-605 Python WASM
- Two short ADRs (1 page each). Map host functions to Rust FFI. CPython vs MicroPython memory/cold-start tradeoff.
- **Exit gate:** TASK-600, TASK-605 → `PENDING`.

### Day 6 — Dockerfile hardening
**Closes:** BUG-051, TASK-1020
- Add `USER nobody`, `WORKDIR /app`, `EXPOSE 8080`, `HEALTHCHECK CMD ./ghost-ops version`.
- Multi-arch build via Docker buildx in CI.
- **Exit gate:** `docker run` works as non-root.

### Day 7 — Backlog reconciliation pass
**Closes:** BUG-038, BUG-039
- Delete TASK-401 (duplicates 600–609).
- Insert/clarify TASK-901.1.
- Demote EPIC-701 from "vaulted" to ACTIVE in `ROADMAP.md` and `.system_state` (with audit trail entry).
- Add new tasks 1010–1020 from AUDIT-2026-05-04.md to BACKLOG.md.
- **Exit gate:** Backlog passes a `make audit` doc-lint script (writes if needed).

---

## Week 2 — Capability Enforcement + AI Act 13/15

### Day 8 — Wire `CheckNetworkEgress` into `rpc()`
**Closes:** BUG-021, TASK-702
- Add policy-check call in `pkg/runtime/wazero_host.go:rpc()`. Audit-log denials. New error code `ErrCapabilityDenied`.
- Table-driven tests with `allowedDomains` set/empty.
- **Exit gate:** Guest module attempting unauthorized `rpc()` returns 0 + audit entry.

### Day 9 — Wire `CheckFSJail` into a new `read_file` host function
**Closes:** BUG-022, TASK-703
- Add `read_file(pathPtr, pathLen, outPtr, outLen)` host function gated by `CheckFSJail`.
- Validate `FSJails` paths exist at `LoadModule`.
- Update guest SDK stub.
- **Exit gate:** Guest reading `/etc/passwd` returns 0; reading allowed jail succeeds.

### Day 10 — Article 13: `X-Ghost-Ops-Synthesized` header
**Closes:** BUG-023, TASK-1010
- In `handleServiceInvoke`, set header iff service was synthesized (engine ∈ {ai}).
- Add `// ghost-ops:generated` prepend in `compiler.go` and a test that flagged code is detectable.
- **Exit gate:** `curl -i /services/{id}/invoke` shows header.

### Day 11 — Article 15: `gosec` + `govulncheck` in synthesis pipeline
**Closes:** BUG-024 (synthesis part), TASK-1011
- In `CompileGo`, after writing source, run `gosec -severity high -fmt json` and `govulncheck`. Fail compile on HIGH.
- Configurable bypass via env `GHOST_OPS_SKIP_SCAN=true` (non-prod only, audit-logged).
- **Exit gate:** Test seeds known-vulnerable Go snippet; compile rejects.

### Day 12 — Audit log secret masking + JSONL store
**Closes:** BUG-026, BUG-044, TASK-1014
- Add `pkg/logging/sanitize.go` with regex-based redaction.
- Add `pkg/logging/audit_store.go` JSONL append + size rotation (default 100 MB).
- Wire into existing `Audit()` callers.
- **Exit gate:** Test asserts `api_key=sk-...` becomes `api_key=***`.

### Day 13 — LLM timeout + sanitized errors
**Closes:** BUG-025, BUG-029, TASK-1015
- `cfg.LLM.Timeout` (default 60s); wrap `Generate` calls.
- OpenAI error path parses `error.message`, never logs body.
- **Exit gate:** `httptest` server stalls 90s → request fails fast at 60s.

### Day 14 — LRU cache for LLM provider + compiler env hygiene
**Closes:** BUG-042, BUG-043, TASK-1017, TASK-1018
- Replace `pkg/llm/cache.go` map-iteration eviction with `container/list` LRU.
- `compiler.go` passes minimal env to `go build`.
- Benchmark cache hit ratio vs random eviction.
- **Exit gate:** Bench shows hit ratio improvement on repeated prompts.

---

## Week 3 — Close the ZHO Loop

### Day 15 — Re-prompt subscriber (registry side)
**Closes:** BUG-028 (half), TASK-1013
- In `pkg/registry/registry.go:StartEventLoop`, subscribe to `EventRePromptRequired`.
- Look up service blueprint, append re-prompt context (current p99, error rate) to intent, requeue.
- **Exit gate:** Unit test: emit event → see new blueprint in intent source.

### Day 16 — Article 14 shadow-mode timer
**Closes:** BUG-027, TASK-1012
- Add `ShadowEnteredAt time.Time` to `ServiceRecord`.
- Background promoter goroutine: if shadow age ≥ `cfg.ShadowMinDuration` (default 5m) **and** shadow error rate ≤ active error rate, promote.
- **Exit gate:** Integration test: shadow promoted at 5m+ε, not before; shadow with bad metrics rolled back.

### Day 17 — Health check → re-prompt feedback
**Closes:** BUG-045 (other half of BUG-028)
- After `purge_unhealthy_service`, emit `EventRePromptRequired{reason: "unhealthy"}`.
- **Exit gate:** Integration test: kill module → registry purges → new blueprint enqueued → re-evolve runs.

### Day 18 — Reconcile error visibility
**Closes:** BUG-040, TASK-1019
- Distinguish `(true, nil)` (deployed) vs `(true, err)` (evolved-but-failed).
- New Prometheus counter `reconcile_errors_total{stage}`.
- **Exit gate:** Forced compile-failure shows up in metrics.

### Day 19 — Buffer truncation hardening
**Closes:** BUG-041
- `nextCommand`/`rpc` return `0xFFFFFFFF` when caller buffer < required size.
- Update guest SDK to retry with doubled buffer on sentinel.
- **Exit gate:** Round-trip test with small buffer succeeds via retry.

### Day 20 — End-to-end ZHO demo test
- New `pkg/api/integration_e2e_test.go::TestZHOFullLoop`:
  1. Submit blueprint → service deploys.
  2. Inject high p99 latency via metric collector seam.
  3. Observer fires re-prompt event.
  4. Registry re-evolves; new version goes to shadow.
  5. After shadow window, promotes.
  6. Old version unloaded.
- **Exit gate:** Test passes with `-race` and `-count=10`.

### Day 21 — PR template + CODEOWNERS + ADR linter
**Closes:** BUG-050
- `.github/PULL_REQUEST_TEMPLATE.md` with reviewer attestation, AI-Act checkboxes.
- `CODEOWNERS` requires reviewer for `pkg/evolution/`, `pkg/runtime/`, `STANDARDS.md`.
- Simple `make adr-lint` checking ADR file exists for any HIGH-RISK task in BACKLOG.
- **Exit gate:** Lint catches a deliberately under-documented task.

---

## Week 4 — Distribute + Demo + Cut v0.6.0

### Day 22 — Etcd client setup
**Closes:** TASK-801
- New `pkg/store/etcd_store.go` skeleton; minimal Get/Put/Watch.
- Behind `STORE_BACKEND=etcd` env flag.
- **Exit gate:** Unit tests pass against `embed.Etcd` test fixture.

### Day 23 — Etcd statestore adapter (CRUD)
**Closes:** TASK-802
- Implement `protocol.StateStore` over Etcd. Reuse JSON marshalling.
- **Exit gate:** Existing `store_test.go` table runs against Etcd backend.

### Day 24 — Distributed leader election
**Closes:** TASK-803
- `etcd-clientv3/concurrency` Election + Mutex; only leader runs reconcile loop.
- Followers serve read API.
- **Exit gate:** Three-node integration test: kill leader → new leader within 5s.

### Day 25 — Gateway: rate limiting
**Closes:** TASK-908
- Reuse `pkg/resilience/rate_limiter.go` per-IP bucket.
- 429 with `Retry-After`.
- **Exit gate:** Test: 100 reqs/s on 10/s bucket → 90 rejections.

### Day 26 — Gateway: dynamic route reconfiguration
**Closes:** TASK-902
- Atomic pointer swap of routes table; no connection drops.
- **Exit gate:** Test: change route mid-stream of 1k requests → 0 dropped, 0 errors.

### Day 27 — Gateway load test (<2ms p99 overhead)
**Closes:** TASK-906
- `*_benchmark_test.go` simulating 10k qps; record overhead distribution.
- **Exit gate:** p99 < 2ms baseline established and committed.

### Day 28 — Hardening pass: race + leak hunt
- `go test -race ./...`, `go test -timeout=5m -count=20 ./pkg/runtime/...`.
- Goroutine leak detector (e.g. `go.uber.org/goleak`) at process boundaries.
- **Exit gate:** Zero new findings.

### Day 29 — Documentation freeze + release notes
- Update `docs/release/release-notes.md` for v0.6.0.
- Refresh `docs/architecture/system_design.md` (BUG-048): real component diagram + contracts.
- Bump VERSION → 0.6.0; update README badge; update `.system_state`.
- **Exit gate:** `make build && make test && docker build .` clean.

### Day 30 — v0.6.0 cut + recorded demo
- Tag `v0.6.0`. Push.
- Record 5-min demo: blueprint submission → AI-act header → metric injection →
  re-prompt → shadow → promote → invoke new version.
- **Exit gate:** GitHub Release with binary, demo video link, and a one-page changelog.

---

## Daily Discipline (every day)

1. Pull main, branch `<type>/<task-or-bug-id>`.
2. Open draft PR with checklist linked to today's exit gate.
3. Implement (<300 LOC + tests).
4. Run locally: `make lint && make test && gosec ./... && govulncheck ./...`.
5. Push; CI must be green.
6. Update `docs/.system_state` audit_log entry: `actor`, `action`, `task`, `notes`.
7. Mark BACKLOG entry `[x]` with date.
8. Merge before EOD.

---

## What This Plan Deliberately Does NOT Try

- **Multi-language SDKs (TASK-600–609 implementation)** — design ADRs only (Days 5).
  Real Rust/Python implementation is Q3 work; cramming it into 30 days breaks the
  Phase-1 quality bar.
- **OAuth2 gateway auth (TASK-1005)** — Q4 work, Article-14 high-risk-candidate.
- **Log filtering UI (TASK-1004)** — needs design; out of MVP scope.
- **Blue/green traffic split (TASK-903)** — depends on Day 26 (dynamic routes); deferred.

These remain in BACKLOG.md, scheduled for Q3/Q4 per ROADMAP.md.

---

## Risk Register

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| ADR-002 (Etcd) takes >1 day | M | H | Time-box to 1 day; if blocked, ship Redis-only and defer Etcd to Q3 |
| `gosec` flags existing code | H | M | Day 2 includes triage budget; suppress with explicit `// #nosec G104 — reason` |
| Shadow-mode timer flake-prone | M | M | Use injectable clock interface; fake clock in tests |
| Goroutine leaks discovered Day 28 | M | H | Daily `go test -race` is mandatory from Day 1 |
| LLM provider stalls block CI | L | H | Mock provider for CI; real OpenAI smoke test only on tag |

---

*Sprint plan author: Claude (Opus 4.7), 2026-05-04. Linked audit: `docs/AUDIT-2026-05-04.md`.*
