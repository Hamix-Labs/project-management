# `pkgs/tasks/scheduling`

Pure **Decide** layer for worker readiness and post-commit notify policy. Apply stays in `internal/taskapi/composition` (notify/wake) and `pkgs/agents/worker` (admission). ADR: [docs/adr/ADR-0023-task-scheduling-domain.md](../../../docs/adr/ADR-0023-task-scheduling-domain.md).

## Read order

1. [`types.go`](./types.go) — `FailedPredicate`, `ReadinessResult`, `NotifyDecision`
2. [`predicates.go`](./predicates.go) — `EvaluateWorkerReadiness` (four predicates, fixed order)
3. [`pickup.go`](./pickup.go) — `ShouldNotifyReadyNow` (enqueue pickup gate only)
4. [`dependency.go`](./dependency.go) — `EdgeSatisfied`
5. [`decide_notify.go`](./decide_notify.go) — `DecideNotifyAfterReadyTransition`

## When a task is stuck `ready`

1. Check row: gate, dependencies, `pickup_not_before`, `on_hold` / status
2. Search logs for `failed_predicate=` on the task id (worker admission defer)
3. If not in memory queue → reconcile backstop (2m) or pickup wake
4. SQL mirror lives in [`pkgs/taskcore/store/internal/ready/predicates.go`](../../taskcore/store/internal/ready/predicates.go)

## Invariants (I1–I7)

See ADR-0023. **I7:** immediate notify after `ready` transition does not require full worker readiness; worker defer is the safety net.

## Boundaries

- Imports: `pkgs/taskcore/domain` and `pkgs/taskcore/contract` only (production `*.go`)
- Must not import: store packages, `handler`, `agents`, `gorm`, `internal/taskapi/composition`
