import type { GitWorktree } from "@/types/git";

/** Branch-bound rows only — matches backend gitWorktreeIsFullyRegistered. */
export function isFullyRegisteredWorktree(worktree: GitWorktree): boolean {
  return Boolean(worktree.branch_id?.trim());
}

/**
 * Hamix-managed / task worktrees only — never the primary repo checkout
 * (`is_main`), which exists for sync identity and is not an agent target.
 */
export function isLinkedWorktreeForDisplay(worktree: GitWorktree): boolean {
  return isFullyRegisteredWorktree(worktree) && !worktree.is_main;
}

/** Rows shown on /worktrees/:repositoryId — same as linked display (no primary). */
export function isDetailPageWorktree(worktree: GitWorktree): boolean {
  return isLinkedWorktreeForDisplay(worktree);
}

export function sortDetailPageWorktrees(worktrees: GitWorktree[]): GitWorktree[] {
  return [...worktrees].sort((a, b) =>
    a.name.localeCompare(b.name, undefined, { sensitivity: "base" }),
  );
}

/** Prefer a managed (non-main) worktree when a default id is still needed. */
export function pickDefaultWorktreeId(worktrees: GitWorktree[]): string {
  const managed = worktrees.filter(isLinkedWorktreeForDisplay);
  if (managed.length === 0) return "";
  return sortDetailPageWorktrees(managed)[0]!.id;
}
