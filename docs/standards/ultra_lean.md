# Development Standards

## Languages
- **Host:** Go (Latest Stable)
- **WASM Modules:** Go, Rust, or AssemblyScript (compiled to WASM)

## Style Guide
- **Go:** `gofmt`, `go vet`, `staticcheck`.
- **Commits:** Conventional Commits (e.g., `feat: add intent interface`).

## Package Management
- **Go:** Go Modules (`go.mod`).

## Testing
- Unit tests required for all new logic.
- Table-driven tests preferred for Go.

## Dependency Specs
- Latest stable dependencies only
- No version conflicts allowed
- Use `[VERSION-CHECK-REQ]` tag when unsure about a dependency version

## Hygiene Specs
- 0 style violations
- Strict adherence to this ultra-lean standard

## Test Specs
- **95% unit coverage**
- Integration tests for critical paths
- O(n) check on optimizations
- Security input validation and output sanitization required
