# Ghost Ops
> **Code that writes itself, runs itself, and fixes itself.**
> **North Star:** Zero-Human Operations — the system patches, optimizes, and evolves itself.
> ![Version](https://img.shields.io/badge/version-v0.5.2-blue)

## Docs (flat — only two files)

- **[`docs/BACKLOG.md`](./docs/BACKLOG.md)** — THE single source of truth.
  Vision, DoD, bugs, tasks, demo scenario, 30-day plan, vaulted history.
  Read top-to-bottom: highest priority first, nice-to-haves last.
- **[`docs/STANDARDS.md`](./docs/STANDARDS.md)** — engineering rules
  (Adversarial Triad, EU AI Act mappings, naming, CI gates).

Anything else is noise.

## Quick Start

### Build
```bash
make build
```

### Mock mode (no API key)
```bash
./ghost-ops -engine mock
```

### AI mode
```bash
export OPENAI_API_KEY=your-key
./ghost-ops -engine ai -blueprints ./blueprints/blueprints.json
```

### Init a new project
```bash
./ghost-ops init
```

### Demo (Phase-1, target D20+)
```bash
make demo
# follow docs/BACKLOG.md §5.2 — 9 scripted steps, ~5 min
```

## CLI

- `ghost-ops service list` — list running services
- `ghost-ops service inspect <id>` — service details
- `ghost-ops service logs <id>` — service output
- `ghost-ops config show` — current config
- `ghost-ops audit tail` — append-only audit trail (Phase-1 D29+)

## Status

| Phase | Theme | Status |
|---|---|---|
| 0 | Hygiene | DONE |
| 1 | Self-healing loop | **IN PROGRESS — target 2026-06-02** |
| 2 | Distribution + security | Planned (Q3) |
| 3 | Multi-language guests (Rust, Python) | Design only this month (Q3 implementation) |
| 4 | Full ZHO autonomy | 2027 |

See [`docs/BACKLOG.md`](./docs/BACKLOG.md) §6 for the day-by-day plan.
