# ADR-0097: Stack layer branches on a shared worktree

**Date:** 2026-08-02  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

[ADR-0039](./ADR-0039-fixed-worktree-branch.md) fixed one immutable branch per worktree and forbade checkout at pickup. Same-worktree enqueue (issue #95) needs each task to open its own GitHub stacked PR layer. GitHub stacks require one head branch per PR, each targeting the layer below ([stacked PRs](https://docs.github.com/en/pull-requests/get-started/about-stacked-prs)).

## Decision

1. **Shared worktree, per-task layer branch** — Tasks that share `worktree_id` keep one checkout path and `WorktreeGate` serialization. Each task owns layer branch `hamix/task-<8 hex>` (`TaskBranchName`).
2. **Active branch is mutable** — `git_worktrees.branch_id` is the **currently checked-out** layer. `git_worktrees.name` remains the allocate-time root branch name and is the stable key for computed `worktree_root_task_id`.
3. **Always a local stack** — Root allocate runs `gh stack init --base <default> <rootBranch>` in the managed worktree. Non-root first run runs `gh stack add <layerBranch>` from the current tip (may be an in-progress parent). Single-task worktrees are stacks of one.
4. **Worker checkout** — Under the worktree gate, before HEAD verify, the worker ensures the task’s layer exists, checks it out, and rebinds `branch_id`. Drift against the active layer still fails the run.
5. **Publish** — Only the worktree root task may open PRs; it submits the whole stack (`gh stack submit`). See follow-on open-pr changes (ADR-0082 amendment).

## Consequences

### Positive

- Enqueue UX unchanged (same `worktree_id`).
- Maps directly onto GitHub stacked PRs / `gh stack`.
- Root identity stays stable when the active layer moves.

### Negative / trade-offs

- Softens ADR-0039 immutability; reconcile must tolerate temporary HEAD ≠ historical allocate branch.
- Host requires the `gh stack` extension.
- Parent evolution after child fork needs rebase/sync before submit.

## Alternatives considered

| Alternative | Why rejected |
| --- | --- |
| One branch for the whole family | Cannot form a GitHub stack |
| New worktree per enqueued task | Breaks enqueue “same worktree” product rule |
| Checkout without updating `branch_id` | Breaks ResolveTaskGitContext / HEAD verify |

## See also

- [ADR-0039](./ADR-0039-fixed-worktree-branch.md)
- [ADR-0081](./ADR-0081-hamix-managed-worktrees.md)
- [docs/domain/worktrees-and-branches.md](../domain/worktrees-and-branches.md)
