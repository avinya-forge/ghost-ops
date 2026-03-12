1. **SYNC:** Verify `backlog.md` sync is correct.
2. **DRILL-DOWN:**
   - Select EPIC 701: Implement Capability-Based Security.
   - Atomize into smaller tasks.
     - TASK 701.1: Define Capability Config
     - TASK 701.2: Implement Network Egress Policy Checker
     - TASK 701.3: Implement FS Jail Checker
     - TASK 701.4: Hook Capabilities into WASM Runtime initialization
   - Append to `backlog.md`.
3. **RUN_SH_MODS:** Implement `expanding EPICS...` logic inside `run.sh --backlog` to actually expand un-atomized EPICS into sub-tasks (using bash/awk/grep or Node.js) or just modify the echo to be more useful. Actually, the prompt says "Use https://skills.sh/ patterns to parse backlogs. Update run.sh logic via ./run.sh --skills."
4. **OUTPUT FORMAT:** Follow the specific required output format.
