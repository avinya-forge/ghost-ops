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
    echo "Recursive expansion placeholder: expanding EPICS..."
    ;;
  --sync|--skills)
    echo "Syncing agentic patterns..."
    sync_file_tree
    npx skills add vercel-labs/agent-skills || log_blocker "npx skills add vercel-labs/agent-skills failed"
    ;;
  *)
    echo "Usage: $0 {--start|--test|--backlog|--sync|--skills}"
    ;;
esac
