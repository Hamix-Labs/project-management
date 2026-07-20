# `pkgs/projects/store`

GORM-backed persistence for projects and project context.

| Path | Role |
| --- | --- |
| `store.go` | Public `Store` facade — all `contract.ProjectStore` methods |
| `internal/projects.go` | Project + context item CRUD, snapshots |
| `internal/edges.go` | Context edge CRUD |
| `model/` | GORM models and domain mappers |

`internal/taskapi/composition.API` holds `*Store` and exposes project methods to HTTP and harness. Migrations in `pkgs/tasks/postgres` import `pkgs/projects/store/model` for AutoMigrate.

Shared SQL helpers: [`pkgs/storekernel`](../../storekernel/) ([ADR-0050](../../../docs/adr/ADR-0050-storekernel-extraction.md)).

Tests: [`store_test.go`](./store_test.go).
