# Release Notes

## v0.6.0
- EPIC-901: HTTP Gateway path + header routing (Minimal — see backlog for missing pieces)

## v0.5.0
- TASK-320: Sidecar proxy pattern
- TASK-321: mTLS between services
- TASK-331: Resource-aware scheduling (bin-pack)
- TASK-1000..1002: Log streaming + anomaly detection prompt
- TASK-1358: Remove `/cluster/health` API

## v0.4.0
- TASK-1200: Evolution engine priority queue
- Architecture diagrams, code-gen evals, OTLP exporter, trace-log correlation, evolution priority queue, prompt-injection defenses, security chaos testing, audit logging for state changes, cluster health dashboard data.

## v0.3.0
- Service Mesh Lite (Retry / Circuit Breaker / Rate Limit), Redis distributed lock (`SET NX`), OpenTelemetry tracing middleware, in-memory event bus, API validation middleware (`MaxBytes`, `ContentType`, `Recover`), `ghost-dev` hot-reload, pre-commit hooks, end-to-end integration tests, async WASM module instantiation, LRU module cache, health-check active-state verification, `POST /services/{id}/invoke`, CLI subcommands (`init`, `service list/inspect/logs`, `config show`), Wazero CPU+memory limits, secure log capture, `AppError` standardization, Viper-backed config with hot-reload, OpenAPI 3.0 spec, LLM token-usage metrics, dependency graph cycle detection, `gosec` initial pass, RuntimeHost benchmarks (~0.08 ms overhead).

## Pre-v0.3.0
- BUG-001..020: First bug-hunt sprint (17 fixed, 3 confirmed non-bugs) (Shipped 2026-04-22)

## v0.6.1 (Current Sprint)
- BUG-053: `JSONFileStore.save()` is non-atomic — crash mid-write corrupts the entire state file
- BUG-030: CI does not enforce STANDARDS §6 — missing `-race`, coverage gate, `gosec`, `govulncheck`, `gofmt -l`
- BUG-057: `JSONFileStore.save()` does not `fsync` — power-cut after `WriteFile` returns can lose the write
- BUG-032: Duplicate Backlog SSOT — legacy `docs/planning/backlog.md` removed
- BUG-033: Duplicate Roadmap stub `docs/planning/roadmap.md` (1 line) removed
- BUG-034: Duplicate STANDARDS — `docs/rules/standards.md` shadow of SSOT removed
- BUG-035: Stray planning scratch — `docs/planning/{active_tasks.txt, session_summary.md, map.md}` removed
- BUG-036: Stray repo-root scratch — `plan.md`, `move_epic.py`, `test_task500.txt` removed
- BUG-038: TASK-401 ("Multi-Language SDK Support") duplicates TASK-600..609 — collapsed into §8
- BUG-039: TASK-901.1 missing — backlog jumped 901 → 901.2 → 901.3 — backfilled in §10
- BUG-051: Dockerfile runs as root, `WORKDIR /root/`, no `EXPOSE`, no `HEALTHCHECK`
- BUG-058: `run.sh --sync` calls `npx skills add vercel-labs/agent-skills` — irrelevant to a Go project
- BUG-047: `run.sh --sync/--skills` clauses are dead
- BUG-048: `docs/architecture/system_design.md` was a 4-line stub — folded into §1
- BUG-049: `docs/engineering/{README,conventions}.md` were 6-line stubs — deleted
- BUG-052: `docs/release/metrics.md` was a single line — folded into §11
- BUG-062: `Makefile` has no `audit`, `demo`, `gosec`, `govulncheck`, or `coverage` targets
- BUG-064: Multiple Markdown SSOT files (VISION, DEMO, SPRINT-30D, AUDIT-2026-05-04, ROADMAP, CHANGELOG) — folded into this file
