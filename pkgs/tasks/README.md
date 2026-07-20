# pkgs/tasks

Composition shell for taskapi HTTP: mux registration, middleware, Postgres migrate, SSE realtime, and shared kernels. **Domain CRUD lives in sibling BCs** (`taskcore`, `taskcycles`, `taskchecklist`, …) — not under this tree.

Package comment: [`doc.go`](./doc.go). Persistence wiring: `internal/taskapi/composition` ([ADR-0079](../../docs/adr/ADR-0079-facade-deletion.md)). Index: [docs/agent-map.md](../../docs/agent-map.md).

## Layout

| Path | Role |
| --- | --- |
| [`handler/`](./handler/) | REST/SSE mux; registers BC routes; bootstrap, health, RUM, writepolicy/readpolicy |
| [`handlerhttp/`](./handlerhttp/) | Shared HTTP JSON/path/limit helpers for BC handlers |
| [`middleware/`](./middleware/) | Outer HTTP stack (auth, rate limit, idempotency, …) |
| [`postgres/`](./postgres/) | GORM open + AutoMigrate orchestration (BC models registered here) |
| [`realtime/`](./realtime/) | SSE wire types, Publisher, coalesce |
| [`scheduling/`](./scheduling/) | Worker readiness / pickup predicates (pure Decide) |
| [`apijson/`](./apijson/), [`logctx/`](./logctx/) | JSON errors and request log context |
| [`pkgs/obs/calltrace/`](../obs/calltrace/) | Per-request call stack (`call_path`, helper.io) — not under this tree |
| [`service/`](./service/) | HTTP-agnostic bootstrap/git/retry orchestration used by the shell |
| [`wire/`](./wire/) | `HandlerAPI` composition interface (implemented by `internal/taskapi/composition`) |
| [`devsim/`](./devsim/) | `HAMIX_SSE_TEST` synthetic events |

## Typical wiring (`cmd/taskapi`)

```go
db, err := postgres.Open(dsn, nil)
if err != nil { ... }
if err := postgres.Migrate(ctx, db); err != nil { ... }
s := composition.NewAPI(db)
hub := realtime.NewSSEHub()
http.Handler = handler.NewHandler(s, hub, repoRootOrNil)
```

## Related BCs

| BC | Path |
| --- | --- |
| Task CRUD | [`pkgs/taskcore`](../taskcore/) |
| Cycles | [`pkgs/taskcycles`](../taskcycles/) |
| Checklist | [`pkgs/taskchecklist`](../taskchecklist/) |
| Events | [`pkgs/taskevents`](../taskevents/) |
| Compose | [`pkgs/taskcompose`](../taskcompose/) |
| Projects / settings / git inventory | [`pkgs/projects`](../projects/), [`pkgs/settings`](../settings/), [`pkgs/gitinventory`](../gitinventory/) |
