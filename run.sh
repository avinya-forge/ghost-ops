#!/usr/bin/env bash
# run.sh - Ghost Ops dev convenience wrapper
# Modes: --start  (boot mock-mode binary)
#        --test   (lint + test)
#        --ci     (vet + staticcheck + test + build, mirrors .github/workflows/ci.yml)
#        --audit  (delegates to `make audit`)

set -euo pipefail

case "${1:-}" in
  --start)
    echo "Starting Ghost Ops (mock engine)..."
    make run
    ;;
  --test)
    echo "Running tests & linting..."
    make lint
    make test
    ;;
  --ci)
    echo "Running CI checks..."
    go vet ./...
    if command -v staticcheck >/dev/null 2>&1; then
      echo "Running staticcheck..."
      staticcheck ./...
    else
      echo "staticcheck not found — install: go install honnef.co/go/tools/cmd/staticcheck@latest"
    fi
    echo "Running go test..."
    go test -race -count=1 ./...
    echo "Running go build..."
    go build -v ./...
    echo "CI checks passed."
    ;;
  --audit)
    echo "Running audit..."
    make audit
    ;;
  *)
    echo "Usage: $0 {--start|--test|--ci|--audit}"
    exit 1
    ;;
esac
