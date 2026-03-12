#!/usr/bin/env bash
# run.sh - Self-Evolving Master Script
# Execution Modes: --start, --test, --backlog, --sync (skills.sh)

set +e # non-blocking mode

# 1. SYNC & ACTIVE-POPULATE
mkdir -p docs/planning docs/architecture docs/engineering
if [ ! -f docs/engineering/README.md ]; then
  echo "# Engineering Documentation" > docs/engineering/README.md
fi

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
    grep -E "TASK|DEBT" docs/planning/backlog.md || log_blocker "No TASK or DEBT found in backlog"
    ;;
  --sync|--skills)
    echo "Syncing agentic patterns..."
    npx skills add vercel-labs/agent-skills || log_blocker "npx skills add vercel-labs/agent-skills failed"
    ;;
  *)
    echo "Usage: $0 {--start|--test|--backlog|--sync|--skills}"
    ;;
esac
