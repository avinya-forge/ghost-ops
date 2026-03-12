#!/bin/bash
set -e

COMMAND=$1

case "$COMMAND" in
  audit)
    echo "Running AUDIT..."
    grep -E "TASK|DEBT" docs/planning/backlog.md || true
    ;;
  verify)
    echo "Running VERIFY (Lint/Test)..."
    bash scripts/ci.sh
    ;;
  expand)
    STEP=$2
    if [ -z "$STEP" ]; then
      echo "Error: STEP parameter is required for expand."
      exit 1
    fi
    echo "Running EXPAND for $STEP..."
    grep -A 2 -E "TASK $STEP|TASK $STEP\." docs/planning/backlog.md || true
    ;;
  *)
    echo "Usage: $0 {audit|verify|expand}"
    # omitted exit 1 for bash session compatibility
    ;;
esac
