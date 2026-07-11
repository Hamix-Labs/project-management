# `pkgs/repo/handler`

HTTP surface for workspace file search and commit diff used by the SPA (`/repo/*`). **Contracts:** [docs/api.md](../../docs/api.md), [docs/domain/workspace-repo.md](../../docs/domain/workspace-repo.md).

## Routes

| Method | Path | Handler |
| --- | --- | --- |
| GET | `/repo/search` | Path substring search under worktree root |
| GET | `/repo/file` | File preview for `@`-mention UI |
| GET | `/repo/validate-range` | Line-range validation for mentions |
| GET | `/repo/diff` | Unified commit patch (`git show`) |

All routes require `?worktree_id=`; workspace resolution uses `pkgs/repo.RepoProvider` (see `pkgs/repo/provider.go`).

## Registration

`repohandler.Register(m, repohandler.Deps{Provider: …})` is called from `pkgs/tasks/handler/handler_routes.go`. HTTP contract tests remain in `pkgs/tasks/handler/handler_http_repo_test.go` (full taskapi mux).

## Files

| File | Role |
| --- | --- |
| `handler.go` | `Register`, `Deps`, `Handler` |
| `handlers.go` | Route handlers and wire types |
| `http_helpers.go` | JSON responses, invalid-input detail stripping |
| `httplog.go` | Debug request logging |
