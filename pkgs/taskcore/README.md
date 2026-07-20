# pkgs/taskcore

Task CRUD bounded context: domain types, persistence, and `/tasks*` core HTTP routes.

## Layout

| Path | Role |
| --- | --- |
| `domain/` | `Task`, `TaskDependency`, `TaskGate`, `Status`, `Priority`, `Actor`, retry payloads |
| `contract/` | Focused seams (`TaskReader`/`TaskWriter`/`TaskDepsStore`/`TaskOpsStore`) composed as `TaskCRUDStore`; `TaskGetter`; stats/health |
| `store/` | GORM CRUD, ready queue, stats, dev-mirror, health; `store/model/` for task tables |
| `handler/` | `Register(m, Deps)` — POST/GET/PATCH/DELETE `/tasks`, stats, gate, dependencies, retry |

## Wiring

- **Composition:** `internal/taskapi/composition.NewAPI(db)` constructs `taskcorestore.NewStore(db)` and implements `pkgs/tasks/wire.HandlerAPI` / worker store surfaces ([ADR-0079](../../docs/adr/ADR-0079-facade-deletion.md)).
- **HTTP:** `pkgs/tasks/handler/handler_routes.go` calls `taskcorehandler.Register(m, deps)` with settings/git-compose callbacks from the tasks handler shell.
- **Migrations:** BC models register in [`pkgs/tasks/postgres/migrate/migrate_models.go`](../tasks/postgres/migrate/migrate_models.go).

## Not in taskcore

- SSE hub, bootstrap, writepolicy, readpolicy (`pkgs/tasks/handler`)
- Checklist, cycles, events, compose, settings, git routes (sibling BC handlers)
- Worker scheduling predicates (`pkgs/tasks/scheduling`) — pure Decide; Apply uses this store’s ready queue

## Verification

```powershell
go test ./pkgs/taskcore/... -count=1
```

Boundary: `scripts/check-go.sh` → `step_taskcore_boundary`.
