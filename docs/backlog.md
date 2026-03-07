# Backlog
## Phase 1: MVP (The Self-Healing Loop)
### Epic: Codebase Pruning & API Parity
| ID | User Story (As/I want/So that) | Technical Scope (Modules/Files) | Acceptance Criteria | Priority (MoSCoW) | Effort | Estimated LOC | Implementation Logic | Task Index | Next Task |
|---|---|---|---|---|---|---|---|---|---|
| 1358 | As a user, I want to remove the undocumented /cluster/health API so that the system maintains API parity and drops dead code. | pkg/api/ | [CLEANUP] Remove /cluster/health and tests | Must-Have | 1-2hr | 50-200 | [Done] Remove code, sync health table. | 411/411 | None |
## Phase 2: Scale (Distributed & Resilient)
### Epic: Service Mesh Lite
| ID | User Story (As/I want/So that) | Technical Scope (Modules/Files) | Acceptance Criteria | Priority (MoSCoW) | Effort | Estimated LOC | Implementation Logic | Task Index | Next Task |
|---|---|---|---|---|---|---|---|---|---|
| 320 | As a user, I want to implement sidecar proxy pattern so that the system is improved. | src/, tests/ | [ARCH] Deploy sidecar for network interception. | Must-Have | 1-2hr | 50-200 | [Done] Implement logic, add tests, verify. | 1/410 | 321 |
| 321 | As a user, I want to implement mtls between services so that the system is improved. | src/, tests/ | [SEC] Mutual TLS for service-to-service auth. | Must-Have | 1-2hr | 50-200 | [Done] Implement logic, add tests, verify. | 2/410 | 322 |
### Epic: Advanced Scheduling
| ID | User Story (As/I want/So that) | Technical Scope (Modules/Files) | Acceptance Criteria | Priority (MoSCoW) | Effort | Estimated LOC | Implementation Logic | Task Index | Next Task |
|---|---|---|---|---|---|---|---|---|---|
| 331 | As a user, I want to implement resource-aware scheduling so that the system is improved. | src/, tests/ | [OPT] Bin packing algorithm. [TEST] Maximize density. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 3/410 | 332 |
## Phase 3: Future (Autonomous Evolution)
| ID | User Story (As/I want/So that) | Technical Scope (Modules/Files) | Acceptance Criteria | Priority (MoSCoW) | Effort | Estimated LOC | Implementation Logic | Task Index | Next Task |
|---|---|---|---|---|---|---|---|---|---|
| 400 | As a user, I want to autonomous feedback loop architecture so that the system is improved. | src/, tests/ | [STRAT] Design self-optimizing loop. | Should-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 4/410 | 401 |
| 401 | As a user, I want to multi-language guest sdk support so that the system is improved. | src/, tests/ | [STRAT] Design support for Rust/Python. | Should-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 5/410 | 402 |
| 402 | As a user, I want to automated vulnerability scanning so that the system is improved. | src/, tests/ | [STRAT] Design security scanner for generated code. | Should-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 6/410 | 403 |
### Epic: Autonomous Optimization Loop
| ID | User Story (As/I want/So that) | Technical Scope (Modules/Files) | Acceptance Criteria | Priority (MoSCoW) | Effort | Estimated LOC | Implementation Logic | Task Index | Next Task |
|---|---|---|---|---|---|---|---|---|---|
| 500 | As a user, I want to define metric thresholds for optimization so that the system is improved. | src/, tests/ | [STRAT] Establish baseline metrics. | Should-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 7/410 | 501 |
| 501 | As a user, I want to implement observer agent so that the system is improved. | src/, tests/ | [OBS] Continuously monitor runtime state. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 8/410 | 502 |
| 502 | As a user, I want to trigger re-prompt on latency spike so that the system is improved. | src/, tests/ | [REL] Auto-trigger LLM re-prompt if P99 > 500ms. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 9/410 | 503 |
| 503 | As a user, I want to trigger re-prompt on error rate spike so that the system is improved. | src/, tests/ | [REL] Auto-trigger LLM re-prompt if 5xx > 1%. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 10/410 | 504 |
| 504 | As a user, I want to validate synthesized code in shadow mode so that the system is improved. | src/, tests/ | [TEST] Run new code against mirrored traffic. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 11/410 | 505 |
| 505 | As a user, I want to compare shadow and primary metrics so that the system is improved. | src/, tests/ | [OPT] Ensure new code is actually better. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 12/410 | 506 |
| 506 | As a user, I want to implement hot-swap promotion so that the system is improved. | src/, tests/ | [REL] Promote shadow to primary gracefully. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 13/410 | 507 |
| 507 | As a user, I want to implement auto-rollback so that the system is improved. | src/, tests/ | [REL] Revert if new version degrades. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 14/410 | 508 |
| 508 | As a user, I want to document optimization loop so that the system is improved. | src/, tests/ | [DOC] Detail the ZHO feedback cycle. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 15/410 | 509 |
| 509 | As a user, I want to end-to-end optimization test so that the system is improved. | src/, tests/ | [TEST] Full simulation of failure and self-healing. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 16/410 | 510 |
### Epic: Multi-Language Expansion
| ID | User Story (As/I want/So that) | Technical Scope (Modules/Files) | Acceptance Criteria | Priority (MoSCoW) | Effort | Estimated LOC | Implementation Logic | Task Index | Next Task |
|---|---|---|---|---|---|---|---|---|---|
| 600 | As a user, I want to rust guest sdk design so that the system is improved. | src/, tests/ | [STRAT] Map host functions to Rust interfaces. | Should-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 17/410 | 601 |
| 601 | As a user, I want to implement rust guest sdk base so that the system is improved. | src/, tests/ | [FEAT] Basic memory sharing for Rust. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 18/410 | 602 |
| 602 | As a user, I want to implement rust guest sdk logger so that the system is improved. | src/, tests/ | [FEAT] Hook up structured logging. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 19/410 | 603 |
| 603 | As a user, I want to rust compiler evolution engine so that the system is improved. | src/, tests/ | [FEAT] Support `cargo build --target wasm32-wasi`. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 20/410 | 604 |
| 604 | As a user, I want to test rust compiler engine so that the system is improved. | src/, tests/ | [TEST] Validate WASM output from Rust source. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 21/410 | 605 |
| 605 | As a user, I want to python (wasm) guest sdk design so that the system is improved. | src/, tests/ | [STRAT] Evaluate CPython vs MicroPython for WASM. | Should-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 22/410 | 606 |
| 606 | As a user, I want to implement python guest sdk base so that the system is improved. | src/, tests/ | [FEAT] Bootstrapping Python environment in WASM. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 23/410 | 607 |
| 607 | As a user, I want to python evolution engine so that the system is improved. | src/, tests/ | [FEAT] Bundle Python scripts into WASM modules. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 24/410 | 608 |
| 608 | As a user, I want to update examples with rust/python so that the system is improved. | src/, tests/ | [DOC] Add basic examples. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 25/410 | 609 |
| 609 | As a user, I want to cross-language interop testing so that the system is improved. | src/, tests/ | [TEST] Verify Go host can invoke Rust/Python guests uniformly. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 26/410 | 610 |
### Epic: Advanced Security Hardening
| ID | User Story (As/I want/So that) | Technical Scope (Modules/Files) | Acceptance Criteria | Priority (MoSCoW) | Effort | Estimated LOC | Implementation Logic | Task Index | Next Task |
|---|---|---|---|---|---|---|---|---|---|
| 700 | As a user, I want to wasm sandboxing review so that the system is improved. | src/, tests/ | [STRAT] Identify potential host escapes. | Should-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 27/410 | 701 |
| 701 | As a user, I want to implement capability-based security so that the system is improved. | src/, tests/ | [SEC] Fine-grained permissions per module. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 28/410 | 702 |
| 702 | As a user, I want to enforce network egress policies so that the system is improved. | src/, tests/ | [SEC] Block unauthorized outgoing calls from WASM. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 29/410 | 703 |
| 703 | As a user, I want to implement file system jails so that the system is improved. | src/, tests/ | [SEC] Restrict WASM disk access strictly to allowed dirs. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 30/410 | 704 |
| 704 | As a user, I want to automated vulnerability scanning so that the system is improved. | src/, tests/ | [SEC] Scan generated code for common CVEs. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 31/410 | 705 |
| 708 | As a user, I want to secret management integration so that the system is improved. | src/, tests/ | [SEC] Fetch secrets securely (Vault/AWS SM). | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 32/410 | 709 |
| 709 | As a user, I want to security architecture guide so that the system is improved. | src/, tests/ | [DOC] Document trust boundaries. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 33/410 | 710 |
### Epic: Cluster State Management
| ID | User Story (As/I want/So that) | Technical Scope (Modules/Files) | Acceptance Criteria | Priority (MoSCoW) | Effort | Estimated LOC | Implementation Logic | Task Index | Next Task |
|---|---|---|---|---|---|---|---|---|---|
| 800 | As a user, I want to etcd integration strategy so that the system is improved. | src/, tests/ | [STRAT] Evaluate Etcd vs Redis for consensus. | Should-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 34/410 | 801 |
| 801 | As a user, I want to implement etcd client setup so that the system is improved. | src/, tests/ | [FEAT] Basic connection handling. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 35/410 | 802 |
| 802 | As a user, I want to etcd statestore adapter so that the system is improved. | src/, tests/ | [FEAT] CRUD operations using Etcd. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 36/410 | 803 |
| 803 | As a user, I want to distributed leader election so that the system is improved. | src/, tests/ | [REL] Ensure single active reconciler loop. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 37/410 | 804 |
| 804 | As a user, I want to state synchronization protocol so that the system is improved. | src/, tests/ | [REL] Propagate state changes to worker nodes. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 38/410 | 805 |
| 805 | As a user, I want to partition tolerance testing so that the system is improved. | src/, tests/ | [TEST] Simulate network splits. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 39/410 | 806 |
| 807 | As a user, I want to node auto-discovery so that the system is improved. | src/, tests/ | [REL] Dynamic scaling of worker pool. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 40/410 | 808 |
| 808 | As a user, I want to graceful node draining so that the system is improved. | src/, tests/ | [REL] Safely evict services on shutdown. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 41/410 | 809 |
| 809 | As a user, I want to cluster setup guide so that the system is improved. | src/, tests/ | [DOC] Steps to run a multi-node deployment. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 42/410 | 810 |
### Epic: Dynamic Routing
| ID | User Story (As/I want/So that) | Technical Scope (Modules/Files) | Acceptance Criteria | Priority (MoSCoW) | Effort | Estimated LOC | Implementation Logic | Task Index | Next Task |
|---|---|---|---|---|---|---|---|---|---|
| 900 | As a user, I want to layer 7 gateway design so that the system is improved. | src/, tests/ | [STRAT] Define routing rules format. | Should-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 43/410 | 901 |
| 901 | As a user, I want to implement http gateway so that the system is improved. | src/, tests/ | [FEAT] Map external routes to internal services. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 44/410 | 902 |
| 902 | As a user, I want to dynamic route reconfiguration so that the system is improved. | src/, tests/ | [REL] Update routes without dropping connections. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 45/410 | 903 |
| 903 | As a user, I want to blue/green deployment support so that the system is improved. | src/, tests/ | [REL] Route traffic weights (e.g., 90/10). | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 46/410 | 904 |
| 904 | As a user, I want to path-based routing so that the system is improved. | src/, tests/ | [FEAT] e.g., /api/v1/auth -> auth-service. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 47/410 | 905 |
| 905 | As a user, I want to header-based routing so that the system is improved. | src/, tests/ | [FEAT] e.g., X-Beta: true -> beta-service. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 48/410 | 906 |
| 906 | As a user, I want to gateway load testing so that the system is improved. | src/, tests/ | [TEST] Ensure minimal overhead (<2ms). | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 49/410 | 907 |
| 907 | As a user, I want to websocket support in gateway so that the system is improved. | src/, tests/ | [FEAT] Proxy WS connections to WASM. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 50/410 | 908 |
| 908 | As a user, I want to gateway rate limiting so that the system is improved. | src/, tests/ | [SEC] Global limits per IP. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 51/410 | 909 |
| 909 | As a user, I want to routing configuration guide so that the system is improved. | src/, tests/ | [DOC] Document gateway usage. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 52/410 | 910 |
| 1000 | As a user, I want to implement log streaming pipeline for ai analysis so that the system is improved. | src/, tests/ | [FEAT] Implement implement log streaming pipeline for ai analysis. [TEST] 95% coverage. [LINT] 0-err. [OPT] O(n). [SEC] Sanitize. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 53/410 | 1001 |
| 1001 | As a user, I want to design anomaly detection prompt for llm so that the system is improved. | src/, tests/ | [FEAT] Implement design anomaly detection prompt for llm. [TEST] 95% coverage. [LINT] 0-err. [OPT] O(n). [SEC] Sanitize. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 54/410 | 1002 |
| 1002 | As a user, I want to integrate anomaly detection into log observer so that the system is improved. | src/, tests/ | [FEAT] Implement integrate anomaly detection into log observer. [TEST] 95% coverage. [LINT] 0-err. [OPT] O(n). [SEC] Sanitize. | Must-Have | 1-2hr | 50-200 | Implement logic, add tests, verify. | 55/410 | 1003 |

## Finalized Metric Summary
| Metric | Value |
| --- | --- |
| STATUS | REFINED |
| PHASE | 1, 2, 3 |
| TOTAL LOC | ~40000 |
| PR DELTA LOC | 0 |
| TASKS DONE | 4 |
| IMPLEMENTED IDs | 1357, 320, 1358, 321 |
| READY RATIO | 100% |
| SAY/DO % | 100% |
| VELOCITY | 1 |
| TECH DEBT % | 0 |
| COVERAGE % | 95 |
| CLEANLINESS SCORE | 98% (TDR < 5%) |
| BLOCKERS | None |
| ETA | TBD |
| NEXT TASK | 322 |
| ACTION | Review |
