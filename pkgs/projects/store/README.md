# `pkgs/projects/store`

GORM-backed persistence for projects and project context.

| Path | Role |
| --- | --- |
| `store.go` | Public `Store` facade — all `contract.ProjectStore` methods |
| `internal/projects.go` | Project + context item CRUD, snapshots |
| `internal/edges.go` | Context edge CRUD |
| `model/` | GORM models and domain mappers |

`pkgs/tasks/store.Store` embeds `*Store` and delegates project methods. Migrations in `pkgs/tasks/postgres` import `pkgs/projects/store/model` for AutoMigrate.

Shared SQL helpers: `pkgs/tasks/kernel` (temporary until `pkgs/storekernel/` extraction).

Tests: [`store_test.go`](./store_test.go) (unit); integration via [`pkgs/tasks/store/facade_projects_test.go`](../../tasks/store/facade_projects_test.go).
