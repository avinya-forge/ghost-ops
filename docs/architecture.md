# Ghost Ops Architecture

## System Components

The Ghost Ops system is composed of several core components that work together to provide a self-healing, evolutionary WebAssembly (WASM) runtime environment.

```mermaid
graph TD
    Client([Client]) -->|HTTP Request| API[API Server]

    subgraph "Core System"
        API -->|Submit Intent| Registry[Service Registry]
        API -->|Invoke Service| Registry

        Registry <-->|Trigger Synthesis| Evolution[Evolution Engine]
        Evolution <-->|Generate Code| LLM[LLM Provider]

        Registry -->|Manage Services| Runtime[WASM Runtime Host]

        Registry <-->|Persist State| StateStore[(State Store)]

        Runtime -->|Emit Events| EventBus[Event Bus]
        Registry -->|Subscribe| EventBus

        Runtime -.->|Metrics/Traces| Telemetry[Telemetry/Observability]
        API -.->|Metrics/Traces| Telemetry
        Registry -.->|Metrics/Traces| Telemetry
    end

    subgraph "WASM Guest"
        Runtime -->|Instantiate Module| WASMModule[WASM Module]
        WASMModule -->|Guest SDK calls| HostFunctions[Host Functions]
        HostFunctions -->|Access KV/Logs| Runtime
    end
```

### Core Components

*   **API Server (`pkg/api`)**: Exposes REST endpoints for submitting intents, invoking services, and retrieving logs/metrics.
*   **Service Registry (`pkg/registry`)**: The brain of the system. Manages the lifecycle of services, coordinates evolution, state persistence, and runtime deployment.
*   **Evolution Engine (`pkg/evolution`)**: Takes a blueprint (intent) and synthesizes Go code using an LLM provider, then compiles it into a WASM module.
*   **WASM Runtime Host (`pkg/runtime`)**: Powered by Wazero. Responsible for securely loading, executing, and isolating compiled WASM modules.
*   **State Store (`pkg/store`)**: Persists service records, dependencies, and cluster state (JSON file, Redis, or In-Memory).
*   **Telemetry/Observability (`pkg/telemetry`, `pkg/logging`)**: Integrates with OpenTelemetry for distributed tracing and Prometheus-style metrics, providing full visibility into the system.
*   **Event Bus (`pkg/event`)**: Facilitates internal pub/sub messaging (e.g., service deployed, service unhealthy) to decouple components.

---

## Execution Flow (The ZHO Loop)

This diagram illustrates the lifecycle of a single Work Unit (WU) from a user's initial intent to the final deployed WASM service.

```mermaid
sequenceDiagram
    participant User
    participant API as API Server
    participant Registry as Service Registry
    participant Engine as Evolution Engine
    participant Store as State Store
    participant Runtime as WASM Runtime

    User->>API: POST /reconcile (Blueprint/Intent)
    API->>Registry: Reconcile(Blueprint)

    %% Cycle Detection & Validation
    Registry->>Registry: Validate() & Sanitize()

    %% Evolution Phase
    Registry->>Engine: Evolve(Blueprint)
    Engine->>Engine: Synthesize Go Code (via LLM)
    Engine->>Engine: Compile Go -> WASM
    Engine-->>Registry: Return compiled WASM bytes

    %% Deployment Phase
    Registry->>Runtime: LoadModule(ServiceID, Version, WASM)
    Runtime-->>Registry: Module Loaded successfully

    Registry->>Runtime: SetActiveVersion(ServiceID, Version)
    Runtime-->>Registry: Version Active

    %% Persistence
    Registry->>Store: UpdateService(ServiceRecord)
    Store-->>Registry: State Persisted

    Registry-->>API: Reconciliation completed
    API-->>User: 200 OK

    %% Execution
    Note over User,Runtime: Post-Deployment Execution
    User->>API: POST /services/{id}/invoke
    API->>Registry: Invoke(ServiceID, Payload)
    Registry->>Runtime: Invoke(ServiceID, Payload)
    Runtime->>Runtime: Execute WASM _start
    Runtime-->>Registry: Execution Result
    Registry-->>API: Result Payload
    API-->>User: 200 OK
```

## Resilience and Self-Healing

The system employs Service Mesh Lite patterns (`pkg/resilience`) including Retries, Circuit Breakers, and Rate Limiters to ensure stability during failures or high load. The Registry runs a continuous background health check that unloads unhealthy services, triggering the Zero-Human Operations (ZHO) re-evolution feedback loop.
