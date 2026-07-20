# pkgs/gitwork

Git worktree and branch operations for Hamix worktree management ([Issue #39](https://github.com/AlexsanderHamir/Hamix/issues/39)).

## Which package?

| Need | Package |
| --- | --- |
| Raw `git -C <dir> …` subprocess + stderr capture | [`pkgs/gitcore`](../gitcore/) |
| Bounded fixed-subcommand helpers (e.g. commit patch) over gitcore | [`pkgs/gitexec`](../gitexec/) |
| Worktree/branch/status ops for registered checkouts | **`pkgs/gitwork`** (this package) |
| Persisted repos/worktrees/branches + `/git/*` HTTP | [`pkgs/gitinventory`](../gitinventory/) |
| Workspace path/@-mention search + `/repo/*` HTTP | [`pkgs/repo`](../repo/) |

## Scope

- `OpenRepository` — validate a path is a git root
- `OpenRegisteredCheckout` — open a registered repo from cache, candidate path, or bounded sibling discovery ([git-checkout-resolution.md](../../docs/domain/git-checkout-resolution.md))
- Worktrees — list, add, remove (`git worktree`)
- Branches — list, create, delete, checkout
- `CheckoutStatus` — dirty/clean, HEAD commit time, upstream ahead/behind for one worktree path

## Out of scope

- HTTP handlers — live in [`pkgs/gitinventory/handler`](../gitinventory/handler/) (`/git/*`) and [`pkgs/repo/handler`](../repo/handler/) (`/repo/*`)
- Task/worker binding — `pkgs/agents/worker` + harness
- Merge, rebase, fetch, pull
- Worktrunk CLI integration

## Usage

```go
svc := gitwork.New()
repo, err := svc.OpenRepository(ctx, "/path/to/main")
wt, err := svc.AddWorktree(ctx, repo, "/path/to/wt", gitwork.AddWorktreeOptions{
    Branch: "feature", CreateBranch: true,
})
```

All paths returned are absolute with forward slashes.

## Tests

```powershell
go test ./pkgs/gitwork/... -count=1 -race
```

Requires `git` on PATH (tests skip when absent).
