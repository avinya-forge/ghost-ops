# Ghost Ops Architecture

## System Overview

```mermaid
graph TD
    CLI[Ghost Ops CLI] --> API[API Server]
    API --> REG[Service Registry]
    REG --> STORE[(State Store)]
    REG --> ENGINE[Evolution Engine]
    REG --> RUNTIME[Runtime Host (WASM)]
    RUNTIME --> WASM[WASM Guest Modules]
    WASM --> HOST[Host Functions]
    HOST --> STORE
    HOST --> RUNTIME
    OBS[Telemetry] -.-> RUNTIME
    OBS -.-> REG
    OBS -.-> API
```

## Evolution Flow

```mermaid
sequenceDiagram
    participant User
    participant Source as Intent Source
    participant Registry
    participant Engine as Evolution Engine
    participant Store
    participant Runtime

    User->>Source: Update Blueprint (Intent)
    loop Reconcile Loop
        Registry->>Source: GetNextBlueprint()
        Source-->>Registry: Blueprint
        Registry->>Store: ListServices()
        Registry->>Engine: Evolve(Blueprint)
        Engine-->>Registry: WASM Binary
        Registry->>Runtime: LoadModule(WASM)
        Registry->>Runtime: SetActiveVersion()
        Registry->>Store: UpdateService()
    end
```
