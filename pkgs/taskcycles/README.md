# pkgs/taskcycles

Bounded context for **execution cycles** — attempts, phases, stream events, verdict reports, and indexed commits. Extracted from `pkgs/tasks` per [ADR-0053](../../docs/adr/ADR-0053-bounded-context-taskcycles.md); domain and GORM models colocated per [ADR-0058](../../docs/adr/ADR-0058-taskcycles-domain-model.md).

HTTP routes (`/tasks/{id}/cycles*`, commits, cycle-failures) and JSON shapes are unchanged from the pre-extraction API. `pkgs/tasks/handler` registers routes via `pkgs/taskcycles/handler.Register`.

## Layout

| Package | Path | Responsibility |
| --- | --- | --- |
| Domain | [`domain/`](./domain/) | Cycle entities, phase/cycle enums, state machine, correlation helpers — stdlib + `taskchecklist/domain` (VerifierKind) only |
| Contract | [`contract/`](./contract/) | `CycleStore` interface + cycle wire input types |
| Store | [`store/`](./store/) | GORM facade; `internal/{cycles,reports,commits}`; `model/` holds GORM rows + mappers |
| Handler | [`handler/`](./handler/) | `/tasks/{id}/cycles*`, commits, `/tasks/cycle-failures` REST routes |

## Wiring

- **`cmd/taskapi`** still constructs `pkgs/tasks/store.Store` as the composition root.
- `tasks/store.Store` embeds `taskcycles/store.Store` and delegates through `facade_cycles.go`, `facade_commits.go`, and `facade_reports.go`.
- `pkgs/tasks/handler/handler_routes.go` calls `taskcycles/handler.Register`.
- `FailureSurfaceMessage` is exported from `taskcycles/store` for `pkgs/tasks/store/internal/stats` (operator-facing failure text on cycle_failed mirrors).
- `pkgs/tasks/store/model/migrate_models.go` registers `taskcycles/store/model` types in FK-safe order.

## Dependency rules

| Package | May import | Must not import |
| --- | --- | --- |
| `domain` | stdlib, `taskchecklist/domain` | GORM, `pkgs/tasks/*` |
| `contract` | `taskcycles/domain`, `pkgs/tasks/domain` (`Actor`) | `pkgs/tasks/handler`, `pkgs/tasks/store/internal` |
| `store` | `taskcycles/domain`, `taskcycles/contract`, `taskcycles/store/model`, GORM, `pkgs/storekernel`, `pkgs/tasks/domain`, `pkgs/tasks/store/model` (task row FK only) | `pkgs/tasks/handler`, `pkgs/tasks/store/internal` |
| `handler` | `taskcycles/domain`, `taskcycles/contract`, `pkgs/tasks/apijson`, `pkgs/tasks/calltrace`, `pkgs/tasks/logctx`, `pkgs/tasks/domain` | `pkgs/tasks/store` facade, `pkgs/tasks/handler` |

Enforced in CI: `scripts/check-go.sh` → `step_taskcycles_boundary`.

## Tests

```powershell
go test ./pkgs/taskcycles/... -count=1
go test ./pkgs/tasks/store/... -run Cycle -count=1
go test ./pkgs/agents/harness/... -run Cycle -count=1
```

## Docs

- [docs/api.md](../../docs/api.md) — cycle, commit, and verdict endpoints
- [docs/data-model.md](../../docs/data-model.md) — cycles and phases
