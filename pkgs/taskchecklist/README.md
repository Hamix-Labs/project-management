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

- **`cmd/taskapi`** still constructs `pkgs/tasks/store.Store` as the composition root.
- `tasks/store.Store` embeds `taskchecklist/store.Store` and delegates through [`facade_checklist.go`](../tasks/store/facade_checklist.go).
- `notifyUnblockedDependents` runs in the tasks facade after completion mutations that set `criteria_satisfied_at`.
- `pkgs/tasks/handler/handler_routes.go` calls `taskchecklist/handler.Register` with `NotifyTaskUpdated` for enriched SSE.
- `pkgs/tasks/store/model/migrate_models.go` registers `taskchecklist/store/model` types in FK-safe order.

## Dependency rules

| Package | May import | Must not import |
| --- | --- | --- |
| `domain` | stdlib | GORM, `pkgs/tasks/*` |
| `contract` | `taskchecklist/domain`, `pkgs/tasks/domain` (`Actor`) | `pkgs/tasks/handler`, `pkgs/tasks/store/internal` |
| `store` | `taskchecklist/domain`, `taskchecklist/contract`, `taskchecklist/store/model`, GORM, `pkgs/storekernel`, `pkgs/tasks/domain`, `pkgs/tasks/store/model` (task/cycle rows only) | `pkgs/tasks/handler`, `pkgs/tasks/store/internal` |
| `handler` | `taskchecklist/domain`, `taskchecklist/contract`, `pkgs/tasks/apijson`, `pkgs/tasks/calltrace`, `pkgs/tasks/logctx`, `pkgs/tasks/domain` | `pkgs/tasks/store` facade, `pkgs/tasks/handler` |

Enforced in CI: `scripts/check-go.sh` → `step_taskchecklist_boundary`.

## Tests

```powershell
go test ./pkgs/taskchecklist/... -count=1
go test ./pkgs/tasks/store/... -run Checklist -count=1
go test ./pkgs/tasks/handler/... -run Checklist -count=1
```

Cross-route HTTP contract tests for checklist remain in [`pkgs/tasks/handler/handler_http_checklist_contract_test.go`](../pkgs/tasks/handler/handler_http_checklist_contract_test.go).

## See also

- [docs/api.md](../../docs/api.md) — `/tasks/{id}/checklist*`
- [pkgs/taskchecklist/contract/checklist.go](./contract/checklist.go) — `ChecklistStore` interface
- [ADR-0056](../../docs/adr/ADR-0056-taskchecklist-domain-model.md) — domain/model ownership
