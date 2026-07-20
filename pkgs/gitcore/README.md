# pkgs/gitcore

Lowest git layer: run `git -C <dir> <args...>` as a subprocess and return stdout/stderr.

No domain types, HTTP, or observability — callers own sentinel mapping and tracing. Used by [`pkgs/gitexec`](../gitexec/) and [`pkgs/gitwork`](../gitwork/).

See the stack map in [`pkgs/gitwork/README.md`](../gitwork/README.md#which-package).

```powershell
go test ./pkgs/gitcore/... -count=1
```
