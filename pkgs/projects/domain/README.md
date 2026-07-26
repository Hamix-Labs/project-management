# `pkgs/projects/domain`

Core project types and validation for the projects bounded context. No imports from `pkgs/tasks` or persistence layers.

| File | Contents |
| --- | --- |
| `project.go` | `Project` |
| `enums.go` | `ProjectStatus` |
| `default_project.go` | Legacy global default project helpers |
| `errors.go` | `ErrNotFound`, `ErrInvalidInput`, `ErrConflict` |

Tasks reference projects by `ProjectID *string` only — no import cycle.
