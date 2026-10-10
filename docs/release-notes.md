# Release Notes

## v0.6.1 (Current Sprint)

### Security & AI Act Compliance
- [x] BUG-030: CI does not enforce STANDARDS §6 — missing `-race`, coverage gate, `gosec`, `govulncheck`, `gofmt -l`
- [x] BUG-051: Dockerfile runs as root, `WORKDIR /root/`, no `EXPOSE`, no `HEALTHCHECK`

### State Management
- [x] BUG-053: `JSONFileStore.save()` is non-atomic — crash mid-write corrupts the entire state file
- [x] BUG-057: `JSONFileStore.save()` does not `fsync` — power-cut after `WriteFile` returns can lose the write

### General / Tech Debt
- [x] BUG-032: Duplicate Backlog SSOT — legacy `docs/planning/backlog.md` removed
- [x] BUG-033: Duplicate Roadmap stub `docs/planning/roadmap.md` (1 line) removed
- [x] BUG-034: Duplicate STANDARDS — `docs/rules/standards.md` shadow of SSOT removed
- [x] BUG-035: Stray planning scratch — `docs/planning/{active_tasks.txt, session_summary.md, map.md}` removed
- [x] BUG-036: Stray repo-root scratch — `plan.md`, `move_epic.py`, `test_task500.txt` removed
- [x] BUG-038: TASK-401 ("Multi-Language SDK Support") duplicates TASK-600..609 — collapsed into §8
- [x] BUG-039: TASK-901.1 missing — backlog jumped 901 → 901.2 → 901.3 — backfilled in §10
- [x] BUG-058: `run.sh --sync` calls `npx skills add vercel-labs/agent-skills` — irrelevant to a Go project
- [x] BUG-047: `run.sh --sync/--skills` clauses are dead
- [x] BUG-048: `docs/architecture/system_design.md` was a 4-line stub — folded into §1
- [x] BUG-049: `docs/engineering/{README,conventions}.md` were 6-line stubs — deleted
- [x] BUG-052: `docs/release/metrics.md` was a single line — folded into §11
- [x] BUG-062: `Makefile` has no `audit`, `demo`, `gosec`, `govulncheck`, or `coverage` targets
- [x] BUG-064: Multiple Markdown SSOT files (VISION, DEMO, SPRINT-30D, AUDIT-2026-05-04, ROADMAP, CHANGELOG) — folded into this file

---

## v0.6.0

### HTTP Gateway
- [x] EPIC-901: HTTP Gateway path + header routing (Minimal)

---

## v0.5.0

### Microservices
- [x] TASK-320: Sidecar proxy pattern
- [x] TASK-321: mTLS between services
- [x] TASK-331: Resource-aware scheduling (bin-pack)
- [x] TASK-1000..1002: Log streaming + anomaly detection prompt
- [x] TASK-1358: Remove `/cluster/health` API

---

## v0.4.0

### Infrastructure
- [x] TASK-1200: Evolution engine priority queue
- [x] Architecture diagrams, code-gen evals, OTLP exporter, trace-log correlation, evolution priority queue, prompt-injection defenses, security chaos testing, audit logging for state changes, cluster health dashboard data.

---

## v0.3.0

### Core Engine
- [x] Service Mesh Lite (Retry / Circuit Breaker / Rate Limit)
- [x] Redis distributed lock (`SET NX`)
- [x] OpenTelemetry tracing middleware
- [x] In-memory event bus
- [x] API validation middleware (`MaxBytes`, `ContentType`, `Recover`)
- [x] `ghost-dev` hot-reload
- [x] Pre-commit hooks
- [x] End-to-end integration tests
- [x] Async WASM module instantiation
- [x] LRU module cache
- [x] Health-check active-state verification
- [x] `POST /services/{id}/invoke`
- [x] CLI subcommands (`init`, `service list/inspect/logs`, `config show`)
- [x] Wazero CPU+memory limits
- [x] Secure log capture
- [x] `AppError` standardization
- [x] Viper-backed config with hot-reload
- [x] OpenAPI 3.0 spec
- [x] LLM token-usage metrics
- [x] Dependency graph cycle detection
- [x] `gosec` initial pass
- [x] RuntimeHost benchmarks (~0.08 ms overhead).

---

## Pre-v0.3.0

- [x] BUG-001..020: First bug-hunt sprint (17 fixed, 3 confirmed non-bugs) (Shipped 2026-04-22)
