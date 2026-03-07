# Development Standards and Contributing to Ghost Ops

Ghost Ops is a high-velocity, automated operations platform. We prioritize simplicity, reliability, and automation.

## Core Philosophy

1. **Simple-First**: Prioritize the simplest solution that works. Avoid over-engineering.
2. **Zero-Human Operations**: Aim for full automation.
3. **Ephemeral Code**: Treat code as disposable. Don't get attached.
4. **Adversarial Triad**:
   - 95% Test Coverage
   - 0 Lint Errors
   - O(n) Optimization

## Development Workflow

1. **Pick a Task**: Check `docs/backlog.md` for open tasks.
2. **Create a Branch**: Use a descriptive name (e.g., `feat/add-config-struct`).
3. **Develop**:
   - Write tests first (TDD).
   - Keep changes small and focused.
   - Run `make lint` and `make test` frequently.
   - **Recommended**: Run `make hooks` to install git pre-commit hooks that automatically run lint and tests before commit.
4. **Submit a PR**:
   - Ensure checks pass.
   - Describe your changes clearly.
   - Reference the Task ID from the backlog.

## Languages
- **Host:** Go (Latest Stable)
- **WASM Modules:** Go, Rust, or AssemblyScript (compiled to WASM)

## Style Guide
- **Go:** `gofmt`, `go vet`, `staticcheck`. Follow standard Go conventions.
- **Commits:** Conventional Commits (e.g., `feat: add intent interface`).

## Package Management
- **Go:** Go Modules (`go.mod`).

## Testing
- Unit tests required for all new logic.
- Table-driven tests preferred for Go.
- Integration tests for major flows.
- **95% unit coverage**.
- O(n) check on optimizations.
- Security input validation and output sanitization required.

## Dependency Specs
- Latest stable dependencies only.
- No version conflicts allowed.
- Use `[VERSION-CHECK-REQ]` tag when unsure about a dependency version.

## Hygiene Specs
- 0 style violations. We use `staticcheck` and `go vet`. Zero tolerance for lint errors.
- Strict adherence to this ultra-lean standard.
- Update `docs/` as needed. Keep comments concise.

## Release Process

1. Update `VERSION`.
2. Update `docs/release-notes.md`.
3. Create a git tag.

## Reporting Issues

If you find a bug, please create an issue with:
- Steps to reproduce.
- Expected vs actual behavior.
- Logs/Screenshots.

Thank you for contributing!