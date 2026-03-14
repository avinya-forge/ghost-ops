# Ghost Ops
> **"Dark Software"**: Code that writes itself, runs itself, and fixes itself.
> **North Star:** Zero-Human Operations (ZHO). The system patches, optimizes, and evolves itself.
> ![Version](https://img.shields.io/badge/version-v0.5.0-blue)

## The Hook
- **Value Prop:** Zero-Human Operations. Code that writes itself, runs itself, and fixes itself.
- **Context:** Autonomous Architecture & Documentation Engine.
- **Quick Start:**
  - Mock Mode: `./ghost-ops -engine mock`
  - Local/Docker: `./run.sh --start`

## Pulse-Table
| Milestone | Ver | Phase | Status | Debt% |
| :--- | :--- | :--- | :--- | :--- |
| AUTONOMY | v0.5.0 | 3 | ACTIVE | 0% |
| CAPABILITY | v0.6.0 | 1 | PLANNING | 0% |

## Visual-Index
- [Vision & Architecture](./docs/architecture/vision.md)
- [Backlog](./docs/planning/backlog.md)
- [Map](./docs/planning/map.md)
- [Release Notes](./docs/release/release-notes.md)
- [Metrics](./docs/release/metrics.md)
- [Rules: Habits](./docs/rules/habits.md)
- [Rules: Hygiene](./docs/rules/hygiene.md)
- [Rules: Decisions](./docs/architecture/decisions.md)
- [Standards: Ultra-Lean](./docs/rules/standards.md)
- [Engineering](./docs/engineering/README.md)

## Quick Start Details
Closing the loop between Intent, Code, and Runtime.
- **Goal:** Intent -> Code -> WASM -> Runtime -> Feedback.
- **Key Deliverables:** AI Evolution Engine, Shadow Mode, persistent state, basic CLI.

### Build
```bash
make build
```

### Run (Mock Mode)
Run with the mock engine (no API key needed, returns dummy WASM):
```bash
./ghost-ops -engine mock
```

### Initialization
Initialize a new project with default configuration and blueprints:
```bash
./ghost-ops init
```

### Run (AI Mode)
```bash
export OPENAI_API_KEY=your-key
./ghost-ops -engine ai -blueprints ./blueprints/blueprints.json
```

### CLI Commands
- `ghost-ops service list`: List all running services.
- `ghost-ops service inspect <id>`: Inspect service details.
- `ghost-ops service logs <id>`: View logs for a service.
- `ghost-ops config show`: Show current configuration.
