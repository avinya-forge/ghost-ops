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
