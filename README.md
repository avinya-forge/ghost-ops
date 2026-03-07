# Ghost Ops
> **"Dark Software"**: Code that writes itself, runs itself, and fixes itself.
> **North Star:** Zero-Human Operations (ZHO). The system patches, optimizes, and evolves itself.
> ![Version](https://img.shields.io/badge/version-v0.4.0-blue)

## The Pulse
| Milestone | Version | Phase | Status | Debt | Density |
| :--- | :--- | :--- | :--- | :--- | :--- |
| MVP | v0.4.0 | 1 | ACTIVE | 0% | 60/60 |

## Documentation Map
- [Vision & Architecture](./docs/vision.md)
- [Backlog](./docs/backlog.md)
- [Release Notes](./docs/release-notes.md)
- [Rules: Habits](./docs/rules/habits.md)
- [Rules: Hygiene](./docs/rules/hygiene.md)
- [Rules: Decisions](./docs/rules/decisions.md)
- [Standards: Ultra-Lean](./docs/standards/ultra-lean.md)

## Quick Start (Active Milestone: MVP - The Self-Healing Loop)
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
