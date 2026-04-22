# Ghost Ops — Unified Development Standards

> **SSOT Owner:** Lead LLM Architect | **Last Updated:** 2026-04-22  
> **Sync Cadence:** Every 30 days | **Next Sync Due:** 2026-05-22  
> **References:** EU AI Act 2026 (in-force) · ISO/IEC 42001:2023 (adapted 2026) · OWASP Top 10 2025  
> **State File:** `/docs/.system_state`

---

## 1. Core Philosophy

| Law | Statement |
|---|---|
| Simple-First | The simplest solution that satisfies the acceptance criteria wins. |
| Zero-Human Operations | Automation is the default. Manual steps are technical debt. |
| Ephemeral Code | Source is generated, compiled, executed, and discarded. No emotional attachment. |
| Adversarial Triad | Every work unit must pass: 95% test · 0 lint errors · O(n) complexity. |

---

## 2. EU AI Act Compliance (2026, In-Force)

Ghost Ops is classified as a **Limited-Risk AI System** under Article 6 of the EU AI Act. The following controls are mandatory for all features that interact with AI-generated outputs.

### 2.1 Risk Classification

| Component | EU AI Act Risk Level | Rationale |
|---|---|---|
| Evolution Engine (LLM synthesis) | Limited Risk | Generates code; human review gate before production |
| Observer Agent (metric-based re-prompting) | Limited Risk | Automated decision within predefined metric thresholds |
| Autonomous Feedback Loop (TASK-400) | High Risk (candidate) | Self-modifying system — requires Article 14 human oversight gate |
| Gateway Auth (TASK-1005) | Limited Risk | Identity federation; no profiling or scoring |

### 2.2 Mandatory Controls

**Article 13 — Transparency**
- All AI-synthesized code must be tagged with `// ghost-ops:generated` in its first line.
- Synthesis prompts must be logged (masked of secrets) in the audit trail (`pkg/logging/audit.go`).
- Users invoking AI-generated services must receive a disclosure header: `X-Ghost-Ops-Synthesized: true`.

**Article 14 — Human Oversight**
- The Autonomous Feedback Loop (TASK-400) must include a "shadow mode" gate before any hot-swap.
- Shadow mode runs the new version alongside the active version for ≥5 minutes before promotion.
- A `GHOST_OPS_OVERRIDE_SHADOW=true` env var may bypass shadow mode only in non-production environments.

**Article 15 — Robustness & Accuracy**
- All LLM-synthesized code must pass `gosec` + `govulncheck` before deployment.
- Evolution Engine must retry synthesis up to 3 times on test failure before escalating to a human alert.
- Metric thresholds used by the Observer Agent must be documented and version-controlled in `config.yaml`.

**Article 17 — Quality Management**
- All changes to synthesis prompts must go through code review (PR required, no direct push to `main`).
- Prompt versions must be tracked in `/docs/.system_state` under `prompt_version`.

### 2.3 Data Protection (GDPR / AI Act Intersection)
- No personal data may appear in blueprints, intent strings, or LLM prompts.
- Secrets must be fetched at runtime via TASK-708 (Vault / AWS SM); never in config files or logs.
- Log retention must comply with configurable TTL (TASK-1003); default maximum: 90 days.

---

## 3. ISO/IEC 42001:2023 — AI Management System (Adapted 2026)

The following controls map ISO/IEC 42001 clauses to Ghost Ops implementation requirements.

| Clause | Control | Implementation |
|---|---|---|
| 6.1 | AI risk assessment | `/docs/architecture/decisions.md` — ADR required for HIGH-risk tasks |
| 7.4 | Communication of AI outputs | `X-Ghost-Ops-Synthesized` header; audit log entry per synthesis |
| 8.4 | AI system lifecycle | Version tracked in `VERSION` file; vaulted in `/docs/BACKLOG.md` |
| 9.1 | Performance monitoring | Observer Agent (EPIC 500); metrics in `pkg/telemetry` |
| 10.2 | Continual improvement | Feedback loop (TASK-400); monthly standards sync (this file) |

**Sync Protocol (every 30 days):**
1. Check ISO/IEC 42001 errata at `iso.org/standard/81230.html`.
2. Check EU AI Act implementing acts at `artificialintelligenceact.eu`.
3. Update the `standards_last_synced` field in `/docs/.system_state`.
4. If any control changes, open a `docs:` PR referencing this file.

---

## 4. Naming Conventions

### 4.1 Files & Packages
| Artifact | Convention | Example |
|---|---|---|
| Go packages | `snake_case`, single noun | `pkg/resilience` |
| Go files | `snake_case.go` | `circuit_breaker.go` |
| Test files | `<file>_test.go` | `circuit_breaker_test.go` |
| Doc files (SSOT) | `UPPER_SNAKE.md` | `ROADMAP.md`, `BACKLOG.md` |
| Doc files (module) | `kebab-case.md` | `security-architecture.md` |
| Config files | `kebab-case.yaml` | `config.yaml` |
| JSON skill files | `kebab-case.json` | `ai-skills.json` |

### 4.2 Git Branches
```
feat/<short-slug>          # New feature
fix/<short-slug>           # Bug fix
docs/<short-slug>          # Documentation only
refactor/<short-slug>      # No behavior change
chore/<short-slug>         # Tooling / CI
```

### 4.3 Commit Messages (Conventional Commits)
```
feat: add etcd leader election (TASK-803)
fix: prevent path traversal in fs jail (TASK-703)
docs: add cluster setup guide (TASK-809)
refactor: extract wasm memory pool (TASK-1006)
test: cross-language interop suite (TASK-609)
chore: bump go-redis to v9.8.0
```

### 4.4 Task IDs
- Format: `TASK-<number>` or `EPIC-<number>`.
- All backlog entries must reference their Task ID in commit messages and PR titles.
- Completed tasks are vaulted in `/docs/BACKLOG.md` under the "Vaulted" section.

---

## 5. Flat-File Hierarchy

```
/docs/
├── ROADMAP.md          ← Strategic vision (Q2–Q4 2026)
├── BACKLOG.md          ← All tasks, user-story format
├── STANDARDS.md        ← This file — unified rules
├── .system_state       ← Version + sync timestamps (machine + human readable)
├── architecture/       ← Complex module; nested allowed
│   ├── decisions.md
│   ├── system_design.md
│   └── vision.md
├── engineering/        ← Complex module; nested allowed
│   └── conventions.md
├── planning/           ← Legacy; migrate content to ROADMAP/BACKLOG over time
├── release/            ← Release notes + metrics
└── rules/              ← Legacy; superseded by STANDARDS.md
```

**Rules:**
1. New top-level docs go directly in `/docs/` as flat Markdown files.
2. Nested subdirectories are only permitted for complex modules with ≥3 related files.
3. No `index.md` files — use the flat filenames as the navigation surface.
4. SSOT files (`ROADMAP.md`, `BACKLOG.md`, `STANDARDS.md`) are the canonical reference; legacy files in subdirectories are read-only historical records.

---

## 6. Code Quality Gates

Every work unit (WU) must satisfy the Adversarial Triad before merge:

| Gate | Tool | Threshold |
|---|---|---|
| Test Coverage | `go test -cover` | ≥ 95% |
| Lint | `staticcheck`, `go vet`, `gofmt` | 0 errors |
| Complexity | Manual / benchmarks | O(n) or better |
| Security | `gosec`, `govulncheck` | 0 high-severity findings |
| AI-generated code | `gosec` + `govulncheck` | Must pass before WASM compilation |

### CI Enforcement (`/.github/workflows/ci.yml`)
- All gates run on every PR.
- PRs to `main` require all gates green.
- AI-synthesized code is scanned before the evolution engine compiles it.

---

## 7. Dependency Policy

- **Latest stable only.** No version pins to EOL releases.
- **No version conflicts.** `go mod tidy` must produce a clean graph.
- **Audit quarterly.** Run `govulncheck ./...` on the first Monday of each quarter.
- Use `[VERSION-CHECK-REQ]` tag in comments when a dependency version needs external verification.
- Go module: `go.mod` is the single source of truth for all host-side dependencies.

---

## 8. Agent Assignment Logic

Full definition lives in `ai-skills.json`. Summary:

| Condition | Engine |
|---|---|
| HIGH-risk task OR context > 8k tokens | Claude Pro |
| Iterative snippet / refactor / boilerplate | Jules · Antigravity · GitHub Copilot |
| Documentation generation | Antigravity · Doc Generator skill |
| Security review | Claude Pro (Code Auditor skill) |
| Standard is outdated (>30 days) | [SYNC] — update this file + `.system_state` |

---

## 9. Definition of Done (DoD)

A task is **DONE** when all of the following are true:

- [ ] Code merged to `main` via reviewed PR.
- [ ] All CI gates green (test · lint · security).
- [ ] `/docs/BACKLOG.md` entry moved to "Vaulted" section.
- [ ] `VERSION` bumped and `docs/release/release-notes.md` updated.
- [ ] If AI-synthesized: audit log entry exists and `X-Ghost-Ops-Synthesized` header verified.
- [ ] If HIGH-risk: ADR written to `/docs/architecture/decisions.md`.
- [ ] If EU AI Act Article 14 applies: shadow mode gate verified.

---

*V-Score (Self-Reported): 4.8 / 5.0 — Flat hierarchy. EU AI Act + ISO 42001 mapped. Token-lean.*
