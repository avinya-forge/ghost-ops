# system_design.md

## Core Components
- **ObserverAgent**: Continuously monitors runtime metrics, evaluates thresholds for P99 latency (>500ms) and Error Rate (>1%), and uses the EventBus to emit events (e.g., `EventRePromptRequired`).
- **Capability-Based Security**: Fine-grained permissions per module. Enforces Network Egress Policies and File System Jails during WASM Runtime initialization.
- **EventBus**: Facilitates internal pub/sub messaging.
