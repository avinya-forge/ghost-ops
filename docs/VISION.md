# Ghost Ops — Vision

> The north star. Lives here, not in BACKLOG.md. The backlog is the ladder;
> this is the rooftop. **Engineering rules:** [`STANDARDS.md`](./STANDARDS.md).
> **Execution:** [`BACKLOG.md`](./BACKLOG.md).

---

## 1. North Star — Zero-Human Operations (ZHO)

**One-liner:** Code that writes itself, runs itself, and fixes itself.

A running Ghost Ops cluster needs **no human in the inner loop**:

1. A human (or upstream system) declares **intent** in plain English.
2. The system **synthesizes** Go source code via an LLM.
3. The system **compiles** to WASM and runs it in a wazero sandbox.
4. The system **observes** runtime metrics (latency, errors).
5. When metrics breach SLOs, the system **re-prompts**, evolves a new version,
   tests it in **shadow mode**, and **hot-swaps** without dropping traffic.
6. The previous version is **discarded**. Source is ephemeral.

A human is needed only for: original intent, EU AI Act Article 14 high-risk
approvals, and prompt/standards reviews.

---

## 2. The Loop (one diagram)

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

---

## 3. Pipeline Laws (non-negotiable)

1. **Latest stable only.** No backwards-compat shims.
2. **Ephemeral source.** Generated Go is compiled, hashed, then deleted.
3. **Simple-first.** Features serving <1% of cases are deferred.
4. **Adversarial Triad.** Every PR: 95% test, 0 lint, O(n) or better,
   `gosec` + `govulncheck` clean. Enforced in CI.
5. **EU AI Act controls.** Articles 13/14/15/17 wired into runtime, not docs.

---

## 4. Component Map

| Package | Responsibility |
|---|---|
| `pkg/intent` | blueprint source (file → queue later) |
| `pkg/evolution` | LLM call + Go→WASM compile + scanner gate |
| `pkg/runtime` | wazero host + capability enforcement |
| `pkg/registry` | reconcile, deploy, shadow-mode promote |
| `pkg/observer` | metric & log thresholds, anomaly detection |
| `pkg/event` | pub/sub bus |
| `pkg/store` | JSON file → Etcd later (cluster-state SSOT) |
| `pkg/api` | REST surface (blueprints + invocations) |
| `pkg/gateway` | L7 router (path/header) |
| `pkg/llm` | provider abstraction + cache |
| `pkg/logging` | structured logs + audit trail |
| `pkg/resilience` | retry, circuit breaker, rate limiter |
| `pkg/sdk/guest` | guest-side ABI stubs |
| `pkg/telemetry` | OpenTelemetry traces + in-memory metrics |

---

## 5. Definition of Done — Phase-1

The system can run the demo (BACKLOG §5) end-to-end with no human touch
between steps 2 and 9, and:

- [ ] **DoD-1** Capability enforcement (network egress + FS jail) actually
      rejects unauthorized calls at the wazero host boundary.
- [ ] **DoD-2** AI-synthesized responses carry `X-Ghost-Ops-Synthesized: true`
      header (Art 13).
- [ ] **DoD-3** Synthesized code passes `gosec` + `govulncheck` before WASM
      compile (Art 15).
- [ ] **DoD-4** Shadow mode holds new version ≥5 min before promotion (Art 14).
- [ ] **DoD-5** `EventRePromptRequired` triggers re-evolve of the offending service.
- [ ] **DoD-6** Audit log persists every synthesis with masked secrets (Art 13/17).
- [ ] **DoD-7** CI runs the full Adversarial Triad and blocks merge on any
      failure.

---

## 6. Roadmap (concise)

| Phase | Timing | Theme | Exit gate |
|---|---|---|---|
| 0 | Done | Hygiene, lint, tests | CI green, structured logs |
| **1** | **Now → 2026-06-02** | **Self-healing loop** | **Demo runs end-to-end** |
| 2 | Q3 2026 | Distribution + security | Etcd leader election, gosec clean |
| 3 | Q4 2026 | Multi-language guests | Go + Rust + Python via same ABI |
| 4 | 2027 | Full ZHO autonomy | EU AI Act audit pass |

---

## 7. The Stakeholder Pitch (30 seconds)

> Today, when a service is too slow, an engineer is paged, reads logs, writes
> code, opens a PR, and waits for CI. That round trip is hours.
>
> Ghost Ops cuts the human out. You declare *intent* in English. The system
> writes the Go code, compiles it to WASM, runs it in a sandbox, watches its
> own latency, and — when SLOs slip — re-prompts itself for a faster version,
> tests the replacement in shadow mode, then hot-swaps. No PR. No deploy. No page.
>
> EU AI Act compliant. Security-jailed. Auditable. v0.6.0 today.

The runnable proof of this pitch is in [`BACKLOG.md`](./BACKLOG.md) §5
(9-step live demo).
