# Architecture Decisions

## 2026-03-07 — Log streaming pipeline and Anomaly Detection

- Implemented `LogObserver` in `pkg/observer` for log streaming pipeline.
- Implemented O(n) sanitation of logs to remove potentially sensitive information (PII).
- Implemented batched log flushing per `serviceID`.
- Integrated LLM for anomaly detection with JSON formatting prompt.

## 2026-03-07 — Resource-aware scheduler for evolution tasks

- Implemented `ResourceScheduler` in `pkg/evolution` to derive queue priority from blueprint constraints.
- Introduced normalized CPU/memory scoring (`cpu_milli`, `memory_mb`) to favor tasks that maximize fit density.
- Added overflow penalty for requests that exceed node capacity to reduce starvation risk for smaller workloads.
- Preserved compatibility by reusing the existing `PriorityQueue` as the scheduler backend.
