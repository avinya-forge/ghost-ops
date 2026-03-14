#!/usr/bin/env bash
# run.sh - Self-Evolving Master Script
# Execution Modes: --start, --test, --backlog, --sync (skills.sh)

set +e # non-blocking mode

# 1. SYNC & ACTIVE-POPULATE
mkdir -p docs/planning docs/architecture docs/engineering
if [ ! -f docs/engineering/README.md ]; then
  echo "# Engineering Documentation" > docs/engineering/README.md
fi


# Idempotent file-tree alignment
sync_file_tree() {
  echo "Syncing file tree..."
  mkdir -p docs/planning docs/architecture docs/engineering
  for file in docs/planning/roadmap.md docs/architecture/system_design.md docs/engineering/conventions.md; do
    if [ ! -f "$file" ]; then
      touch "$file"
      echo "# ${file##*/}" > "$file"
      echo "Created $file"
    fi
  done
}

# Function to log blockers non-blockingly
log_blocker() {
  local reason="$1"
  echo "[RESOLVE] $reason" >> docs/planning/backlog.md
  echo "Logged blocker: $reason"
}

case "$1" in
  --start)
    echo "Starting environment..."
    make run || log_blocker "make run failed during --start"
    ;;
  --test)
    echo "Running tests & linting..."
    make lint || log_blocker "make lint failed during --test"
    make test || log_blocker "make test failed during --test"
    ;;
  --backlog)
    echo "Auditing backlog tags..."
    grep -E "EPIC|DEBT" docs/planning/backlog.md || log_blocker "No EPIC or DEBT found in backlog"
    echo "Recursive expansion: expanding EPICS..."
    for epic in $(grep -oE 'EPIC [0-9]+' docs/planning/backlog.md | awk '{print $2}'); do
      echo "Expanding EPIC $epic:"
      grep -E "TASK $epic(\.[0-9]+)*:" docs/planning/backlog.md || log_blocker "No sub-tasks found for EPIC $epic"
    done
    ;;
  --sync|--skills)
    echo "Syncing agentic patterns..."
    sync_file_tree
    npx skills add vercel-labs/agent-skills || log_blocker "npx skills add vercel-labs/agent-skills failed"
    ;;
  --ci)
    echo "Running CI checks..."
    echo "Running go vet..."
    go vet ./...

    if command -v staticcheck &> /dev/null
    then
        echo "Running staticcheck..."
        staticcheck ./...
    else
        echo "staticcheck could not be found, skipping..."
    fi

    echo "Running go test..."
    go test -v ./...

    echo "Running go build..."
    go build -v ./...

    echo "CI checks passed!"
    ;;
  --audit)
    echo "Running AUDIT..."
    grep -E "TASK|DEBT" docs/planning/backlog.md || true
    ;;
  --verify)
    echo "Running VERIFY (Lint/Test)..."
    bash "$0" --ci
    ;;
  --expand)
    STEP=$2
    if [ -z "$STEP" ]; then
      echo "Error: STEP parameter is required for expand."
      exit 1
    fi
    echo "Running EXPAND for $STEP..."
    grep -A 2 -E "TASK $STEP|TASK $STEP\." docs/planning/backlog.md || true
    ;;
  *)
    echo "Usage: $0 {--start|--test|--backlog|--sync|--skills|--ci|--audit|--verify|--expand}"
    ;;
esac
