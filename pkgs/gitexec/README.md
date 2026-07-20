# pkgs/gitexec

Thin bounded wrappers over [`pkgs/gitcore`](../gitcore/) for fixed subcommands (for example commit patch via `git show`). Does **not** expose arbitrary git argument passthrough.

Prefer this package from HTTP/non-harness callers that need a single known git operation. Worktree/branch orchestration stays in [`pkgs/gitwork`](../gitwork/).

See the stack map in [`pkgs/gitwork/README.md`](../gitwork/README.md#which-package).

```powershell
go test ./pkgs/gitexec/... -count=1
```
