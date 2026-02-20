# Hello World (WASM)

This is a basic example of a Go program compiled to WASM for use with Ghost Ops.

## Build

To build the WASM module, run:

```bash
make build
```

This will produce `hello.wasm` in the current directory.

## Prerequisites

- Go 1.21 or later
- `tinygo` (optional, for smaller binaries) or standard Go compiler (supported by Makefile)

## Notes

The `Makefile` uses `GOOS=wasip1` and `GOARCH=wasm` to target WASI.
