# `pkgs/taskchecklist`

Bounded context for task checklist definition rows, verify commands, and per-subject completion. Extracted from `pkgs/tasks` per [ADR-0051](../../docs/adr/ADR-0051-bounded-context-taskchecklist.md); domain and GORM models colocated per [ADR-0056](../../docs/adr/ADR-0056-taskchecklist-domain-model.md).

HTTP routes (`/tasks/{id}/checklist*`) and JSON shapes are unchanged from the pre-extraction API. `pkgs/tasks/handler` registers routes via `pkgs/taskchecklist/handler.Register`.

## Layout

| Package | Path | Responsibility |
| --- | --- | --- |
| Domain | [`domain/`](./domain/) | `TaskChecklistItem`, verify limits, `VerifierKind`, completion rows — stdlib only |
| Contract | [`contract/`](./contract/) | `ChecklistStore` interface + wire DTOs |
| Store | [`store/`](./store/) | GORM persistence facade; `internal/checklist/` holds CRUD; `model/` holds GORM rows + mappers |
| Handler | [`handler/`](./handler/) | `/tasks/{id}/checklist*` REST handlers |

## Wiring

- **`cmd/taskapi`** constructs `internal/taskapi/composition.API` via `composition.NewAPI(db)` ([ADR-0079](../../docs/adr/ADR-0079-facade-deletion.md)).
- Composition holds `*checkliststore.Store` and exposes checklist methods on the composition API.
- After completion mutations that set `criteria_satisfied_at`, composition runs `NotifyUnblockedDependents` for dependent wake.
- `pkgs/tasks/handler/handler_routes.go` calls `taskchecklist/handler.Register` with `NotifyTaskUpdated` for enriched SSE.
- Model registration for AutoMigrate lives in [`pkgs/tasks/postgres/migrate/migrate_models.go`](../tasks/postgres/migrate/migrate_models.go).

## Dependency rules

| Package | May import | Must not import |
| --- | --- | --- |
| `domain` | stdlib | GORM, `pkgs/tasks/*` |
| `contract` | `taskchecklist/domain`, `pkgs/taskcore/domain` (`Actor`) | `pkgs/tasks/handler`, `internal/taskapi/composition` |
| `store` | `taskchecklist/domain`, `taskchecklist/contract`, `taskchecklist/store/model`, GORM, `pkgs/storekernel`, `pkgs/taskcore/domain`, `pkgs/taskcore/store/model` (task/cycle FK), `pkgs/obs/calltrace` | `pkgs/tasks/handler`, `internal/taskapi/composition` |
| `handler` | `taskchecklist/domain`, `taskchecklist/contract`, `pkgs/tasks/handlerhttp`, `pkgs/tasks/apijson`, `pkgs/obs/calltrace`, `pkgs/tasks/logctx`, `pkgs/taskcore/domain` | `internal/taskapi/composition`, `pkgs/tasks/handler` |

Enforced in CI: `scripts/check-go.sh` → `step_taskchecklist_boundary`.

## Tests

```powershell
go test ./pkgs/taskchecklist/... -count=1
```

## See also

- [docs/api.md](../../docs/api.md) — `/tasks/{id}/checklist*`
- [pkgs/taskchecklist/contract/checklist.go](./contract/checklist.go) — `ChecklistStore` interface
- [ADR-0056](../../docs/adr/ADR-0056-taskchecklist-domain-model.md) — domain/model ownership
