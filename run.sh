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

    # 1. SYNC logic: Reconcile docs/planning/backlog.md against actual codebase
    echo "Reconciling backlog against codebase..."
    # The prompt explicitly specifies the following constraints and behaviors for SYNC logic:
    # - "If code exists but task is `[ ]`, mark `[x]`."
    # - "If task is `[x]` but code is missing, mark `[DEBT]`."

    awk '
    BEGIN { FS="\\|"; OFS="|" }
    function check_files(files_str) {
        # files_str can be like "file1.go, file2.go"
        split(files_str, files, ",")
        for (i in files) {
            # Trim leading/trailing spaces
            file = files[i]
            gsub(/^[ \t]+|[ \t]+$/, "", file)
            # If at least one file exists, we consider code exists
            if (system("test -f \"" file "\"") == 0) {
                return 1
            }
        }
        return 0
    }

    # Match tasks that are incomplete: `> - [ ] TASK:` or `- [ ] TASK:`
    /^[ \t]*>? *-[ \t]+\[[ \t]\] TASK:/ {
        # Find Loc: [files...]
        loc_start = index($0, "Loc: [")
        if (loc_start > 0) {
            files_part = substr($0, loc_start + 6)
            loc_end = index(files_part, "]")
            if (loc_end > 0) {
                files_str = substr(files_part, 1, loc_end - 1)
                if (check_files(files_str)) {
                    sub(/\[[ \t]\]/, "[x]")
                }
            }
        }
    }

    # Match tasks that are complete: `> - [x] TASK:` or `- [x] TASK:`
    /^[ \t]*>? *-[ \t]+\[x\] TASK:/ {
        loc_start = index($0, "Loc: [")
        if (loc_start > 0) {
            files_part = substr($0, loc_start + 6)
            loc_end = index(files_part, "]")
            if (loc_end > 0) {
                files_str = substr(files_part, 1, loc_end - 1)
                if (!check_files(files_str)) {
                    sub(/\[x\]/, "[DEBT]")
                }
            }
        }
    }
    { print $0 }
    ' docs/planning/backlog.md > docs/planning/backlog.md.tmp && mv docs/planning/backlog.md.tmp docs/planning/backlog.md


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
