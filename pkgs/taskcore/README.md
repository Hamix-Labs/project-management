# pkgs/taskcore

Task CRUD bounded context: domain types, persistence, and `/tasks*` core HTTP routes.

## Layout

| Path | Role |
| --- | --- |
| `domain/` | `Task`, `TaskDependency`, `TaskGate`, `Status`, `Priority`, `Actor`, retry payloads |
| `contract/` | `TaskCRUDStore`, stats types, health wire interfaces |
| `store/` | GORM CRUD, ready queue, stats, dev-mirror, health; `store/model/` for task tables |
| `handler/` | `Register(m, Deps)` — POST/GET/PATCH/DELETE `/tasks`, stats, gate, dependencies, retry |

## Wiring

- **Store:** `tasks/store.NewStore` constructs `taskcorestore.NewStore(db)` and delegates facades.
- **HTTP:** `tasks/handler.registerRoutes` calls `taskcorehandler.Register(m, deps)` with settings/git-compose callbacks from the tasks handler shell.
- **Contracts:** `tasks/contract` aliases `TaskCRUDStore` and related types from `taskcore/contract`.

## Not in taskcore

- SSE hub, bootstrap, writepolicy, readpolicy (`pkgs/tasks/handler`)
- Checklist, cycles, events, compose, settings, git routes (sibling BC handlers)
- Worker scheduling predicates (`pkgs/tasks/scheduling`)

## Verification

```powershell
go test ./pkgs/taskcore/... -count=1
```

Boundary: `scripts/check-go.sh` → `step_taskcore_boundary`.
