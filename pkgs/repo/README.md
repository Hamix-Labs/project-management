# pkgs/repo

Workspace path resolution, `@`-mention parsing/validation, file search/preview, and related helpers for prompts and the SPA.

| Path | Role |
| --- | --- |
| Root package | Path safety under configured roots, browse/list, mentions, line counts |
| [`handler/`](./handler/) | `/repo/*` HTTP (`search`, `file`, `validate-range`, `diff`) — [handler README](./handler/README.md) |

Not git inventory: registered repos/worktrees live in [`pkgs/gitinventory`](../gitinventory/). Raw git subprocess helpers live in [`pkgs/gitcore`](../gitcore/) / [`pkgs/gitexec`](../gitexec/) / [`pkgs/gitwork`](../gitwork/).

Stack map: [`pkgs/gitwork/README.md`](../gitwork/README.md#which-package). Domain: [docs/domain/workspace-repo.md](../../docs/domain/workspace-repo.md), [ADR-0049](../../docs/adr/ADR-0049-repo-http-handler.md).

```powershell
go test ./pkgs/repo/... -count=1
```
