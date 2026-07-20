# pkgs/agents

Automation side of taskapi: ready-task queue, worker admission, harness cycle choreography, and runner adapters.

Domain overview: [docs/domain/agent-queue.md](../../docs/domain/agent-queue.md), [docs/domain/harness.md](../../docs/domain/harness.md), [docs/domain/runner-adapters.md](../../docs/domain/runner-adapters.md). Code index: [docs/agent-map.md](../../docs/agent-map.md).

## Layout

| Path | Role |
| --- | --- |
| Root (`memory_queue.go`, `pickup_wake.go`, `reconcile.go`, …) | In-process ready-task queue + reconcile + pickup-wake (`NotifyReadyTask`, `ReconcileReadyTasksNotQueued`, `PickupWakeScheduler`) — keep under root until a dedicated `queue/` package peel (B-39) |
| [`worker/`](./worker/) | Queue consumer: admission, ready→running, orphan sweep; calls harness |
| [`harness/`](./harness/) | Execute/verify loop, resume, verify retry, git integrity — [harness README](./harness/README.md) |
| [`runner/`](./runner/) | `Runner` interface, cursor adapter, registry, `runnerfake` |
| [`agentsmoke/`](./agentsmoke/) | Shared real-cursor smoke fixtures for harness/runner integration tests |

Supervisor boot/reload lives in `internal/taskapi/agentworker/` (not under this tree).

## Naming note

- **`pkgs/agents/runner`** — adapter contract + CLI implementations
- **`pkgs/runners`** — HTTP `/runners*` only ([README](../runners/README.md))

## Start here

1. [harness/README.md](./harness/README.md) for cycle orchestration
2. [docs/architecture.md](../../docs/architecture.md) for process wiring
3. Persistence composition: `internal/taskapi/composition` ([ADR-0079](../../docs/adr/ADR-0079-facade-deletion.md))
