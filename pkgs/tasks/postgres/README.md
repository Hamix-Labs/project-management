# `pkgs/tasks/postgres`

GORM open and schema migrate orchestration for taskapi. BC GORM models live in each `pkgs/*/store/model`; this package owns dial/pool config, public `Migrate`, and schema revision recording ([ADR-0079](../../../docs/adr/ADR-0079-facade-deletion.md)).

## Entry points

| File | Role |
| --- | --- |
| `postgres.go` | `Open`, thin `Migrate` wrapper, timeouts |
| `migrate/` (`package migrate`) | Ordered one-shot steps + AutoMigrate model list |
| `schema_revision.go` | Integer `SchemaRevision` + `schema_meta` bump |
| `gorm_config.go` / `gorm_slog.go` | Logger and pool helpers |
| `startup_log.go` | Migrate/start slog lines |

## Migrate subpackage (`migrate/`)

`postgres.Migrate` calls `migrate.Run`, then `RecordSchemaRevision`. External callers (`cmd`, `tasktestdb`) keep using `postgres.Migrate` only.

| Theme | Files (production) |
| --- | --- |
| **Orchestration** | `run.go`, `columns.go`, `legacy_steps.go` |
| **Git tree / repo identity** | `migrate_contract_git_tree.go`, `migrate_git_common_dir.go`, `migrate_repo_root_to_git_repository.go`, `migrate_drop_repo_root.go`, `migrate_fixed_worktree_branch.go` |
| **Compose / payload + worktree** | `migrate_compose_payload_worktree.go` |
| **Seeds** | `migrate_seed_worktree_branch_tree.go`, `migrate_repo_default_projects.go` |
| **Removals / cleanup** | `migrate_remove_subtasks.go`, `migrate_remove_task_type.go`, `migrate_remove_draft_evaluations.go` |
| **Model registry** | `migrate_models.go` |

Keep new one-shot upgrade steps as `migrate_<verb>_<subject>.go` under `migrate/`. Bump `SchemaRevision` in the same PR as any model or migrate-step change.

## See also

- [docs/configuration.md](../../../docs/configuration.md) - env and migrate overview
- [docs/domain/persistence.md](../../../docs/domain/persistence.md)
