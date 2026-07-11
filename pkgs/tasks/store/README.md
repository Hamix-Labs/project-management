# `pkgs/tasks/store`

GORM-backed persistence for tasks, audit events, checklists, drafts, cycles/phases, the ready-task queue, dev-mirror, and DB health probes. **Architecture and dual-write invariant:** [docs/architecture.md](../../docs/architecture.md), [docs/data-model.md](../../docs/data-model.md). **Behavioral deep-dives:** [docs/domain/persistence.md](../../docs/domain/persistence.md), [docs/domain/task-events.md](../../docs/domain/task-events.md). **How to extend:** [CONTRIBUTING.md](../../CONTRIBUTING.md), [docs/domain/persistence.md](../../docs/domain/persistence.md). API contracts: [docs/api.md](../../docs/api.md).

Package overview and conventions: `go doc -all .` (starts in [doc.go](./doc.go)).

## Architecture

The public package is a **facade**: every `*Store` method in a `facade_*.go` file delegates to a per-domain package under [`internal/`](./internal). Public types (`CreateTaskInput`, `TaskNode`, …) are Go type aliases over the internal definitions, so external callers stay unchanged across reshuffles. Cross-domain transactions are composed by calling the exported `…InTx` helpers from sibling internal packages inside one `*gorm.DB.Transaction(...)`.

Ready-task notifications (`(*Store).notifyReadyTask`) are intentionally only fired by the facade; subpackages return the updated task plus the previous status so the facade can decide whether to notify exactly once. This keeps `internal/notify` out of the per-domain dependency graphs.

## Where code lives

| Concern | Facade file | Tests | Internal package | Notes |
|---|---|---|---|---|
| Wiring | `store.go` | (in `facade_tasks_test.go`) | `internal/notify` | `Store`, `NewStore`, `ReadyTaskNotifier`, `SetReadyTaskNotifier`, `notifyReadyTask`. |
| Projects & project context | `facade_projects.go` | `facade_projects_test.go` | `pkgs/projects/store` | Delegates to `pkgs/projects/store`; integration tests in `facade_projects_test.go`. |
| Tasks — CRUD, lists & trees | `facade_tasks.go` | `facade_tasks_test.go` | `internal/tasks` | `Get`, `Create`, `Update`, `Delete`, `List` / `ListFlat`, `ListRootForest{,After}`, `GetTaskTree`. Readiness delegates to [`pkgs/tasks/scheduling/`](../scheduling/). `CreateTaskInput`, `UpdateTaskInput`, `ParentFieldPatch`, `TaskNode`, `MaxTaskTreeDepth` aliased here. Tests also cover the ready-task notifier wiring and the operation-duration histogram. |
| Stats | `facade_stats.go` | — | `internal/stats` | `GlobalTaskStats`. |
| Checklist | `facade_checklist.go` | `facade_checklist_test.go` | `pkgs/taskchecklist/store` | Delegates to `pkgs/taskchecklist/store`; see [taskchecklist/README.md](../../taskchecklist/README.md). |
| Cycles & phases | `facade_cycles.go` | `facade_cycles_test.go` | `pkgs/taskcycles/store` | Delegates to `pkgs/taskcycles/store`. |
| Cycle commits | `facade_commits.go` | `facade_commits_test.go` | `pkgs/taskcycles/store` | Worker-indexed SHAs — [cycle-commits.md](../../docs/domain/cycle-commits.md). |
| Verify/criteria reports | `facade_reports.go` | `facade_reports_test.go` | `pkgs/taskcycles/store` | Criteria and verify report upserts. |
| Task events | `facade_events.go` | `facade_events_test.go` | `pkgs/taskevents/store` | Audit thread reads; append still dual-written via `storekernel` from CRUD/checklist/cycles. |
| Task drafts + templates | `facade_compose.go` | `facade_compose_test.go` (if present) | `pkgs/taskcompose/store` | Delegates to `pkgs/taskcompose/store`; see [taskcompose/README.md](../../taskcompose/README.md). |
| Agent ready queue | `facade_ready.go` | `facade_ready_test.go`, `scheduling_parity_test.go` | `internal/ready` | `ListReadyTaskQueueCandidates`, `ListReadyTasksUserCreated`, `DefaultReadyTimeout`. SQL dequeuable predicates mirror [`scheduling/`](../scheduling/). |
| Health | `facade_health.go` | `facade_health_test.go` | `internal/health` | `Ping`, `Ready`. |
| Dev simulation | `facade_devmirror.go` | `facade_devmirror_test.go` | `internal/devmirror` | `ApplyDevTaskRowMirror`, `ListDevsimTasks`. |
| Shared kernel | — | `pkgs/storekernel/*_test.go` | `pkgs/storekernel` | `Op*` Prometheus labels, `DeferLatency`, `AppendEvent`, `NextEventSeq`, validators, `MapNotFound`. |

When adding a **new** store method, extend the table above in the same PR and place the implementation in the matching bounded-context store or `internal/<domain>/` package, with a thin `(*Store)` delegation in `facade_<domain>.go` and tests in `facade_<domain>_test.go`.
