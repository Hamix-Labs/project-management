# `pkgs/projects/store`

GORM-backed persistence for projects.

| Path | Role |
| --- | --- |
| `store.go` | Public `Store` facade — all `contract.ProjectStore` methods |
| `internal/projects.go` | Project CRUD, defaults, repo cascade delete |
| `model/` | GORM models and domain mappers |

`internal/taskapi/composition.API` holds `*Store` and exposes project methods to HTTP. Migrations in `pkgs/tasks/postgres` import `pkgs/projects/store/model` for AutoMigrate.

Shared SQL helpers: [`pkgs/storekernel`](../../storekernel/) ([ADR-0050](../../../docs/adr/ADR-0050-storekernel-extraction.md)).

Tests: [`store_test.go`](./store_test.go).
