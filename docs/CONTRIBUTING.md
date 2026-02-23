# Contributing to Ghost Ops

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

## Coding Standards

- **Go**: Follow standard Go conventions. Use `gofmt`.
- **Linting**: We use `staticcheck` and `go vet`. Zero tolerance for lint errors.
- **Testing**:
   - Unit tests for all logic.
   - Integration tests for major flows.
   - Use table-driven tests where appropriate.
- **Documentation**: Update `docs/` as needed. Keep comments concise.

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
