# `pkgs/tasks/realtime`

SSE wire contracts, coalesce policy, `Publisher` port, and the in-process **SSEHub** (ring buffer, subscribe/replay, publish fanout). HTTP `GET /events` framing stays in [`pkgs/tasks/handler`](../handler/).

| File | Role |
|------|------|
| `wire.go` | `ChangeType`, `Event`, `RunProgressPayload` |
| `coalesce.go` | Pure `CoalesceKey` for hub dedup window |
| `publisher.go` | `Publisher` interface (`Publish(Event)`) |
| `sse_hub.go` | `SSEHub`, `Subscribe` / `SubscribeSince`, ring buffer, eviction |
| `notify.go` / `publish_task.go` | Shared publish helpers |

**Importers:** `handler` (stream/notify aliases + HTTP), `cmd/taskapi` / `internal/taskapi` (construct hub), `internal/taskapi/agentworker` (`Publisher` only for produce paths).

See [docs/domain/sse-hub.md](../../docs/domain/sse-hub.md), [ADR-0020](../../docs/adr/ADR-0020-realtime-sse-layout.md), and [ADR-0080](../../docs/adr/ADR-0080-sse-hub-realtime-ownership.md).
