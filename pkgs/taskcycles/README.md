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

- **`cmd/taskapi`** constructs `internal/taskapi/composition.API` via `composition.NewAPI(db)` ([ADR-0079](../../docs/adr/ADR-0079-facade-deletion.md)).
- Composition holds `*cyclesstore.Store` and exposes cycle/commit/report methods on the composition API.
- `pkgs/tasks/handler/handler_routes.go` calls `taskcycles/handler.Register`.
- `FailureSurfaceMessage` is exported from `taskcycles/store` for `pkgs/taskcore/store/internal/stats` (operator-facing failure text on cycle_failed mirrors).
- Model registration for AutoMigrate lives in [`pkgs/tasks/postgres/migrate/migrate_models.go`](../tasks/postgres/migrate/migrate_models.go).

## Dependency rules

| Package | May import | Must not import |
| --- | --- | --- |
| `domain` | stdlib, `taskchecklist/domain` | GORM, `pkgs/tasks/*` |
| `contract` | `taskcycles/domain`, `pkgs/taskcore/domain` (`Actor`) | `pkgs/tasks/handler`, `internal/taskapi/composition` |
| `store` | `taskcycles/domain`, `taskcycles/contract`, `taskcycles/store/model`, GORM, `pkgs/storekernel`, `pkgs/taskcore/domain`, `pkgs/obs/calltrace` | `pkgs/tasks/handler`, `internal/taskapi/composition` |
| `handler` | `taskcycles/domain`, `taskcycles/contract`, `pkgs/tasks/handlerhttp`, `pkgs/tasks/apijson`, `pkgs/obs/calltrace`, `pkgs/tasks/logctx`, `pkgs/taskcore/domain`, `pkgs/taskcore/contract` | `internal/taskapi/composition`, `pkgs/tasks/handler` |

Enforced in CI: `scripts/check-go.sh` → `step_taskcycles_boundary`.

## Tests

```powershell
go test ./pkgs/taskcycles/... -count=1
go test ./pkgs/agents/harness/... -run Cycle -count=1
```

## Docs

- [docs/api.md](../../docs/api.md) — cycle, commit, and verdict endpoints
- [docs/data-model.md](../../docs/data-model.md) — cycles and phases
