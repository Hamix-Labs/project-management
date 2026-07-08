import type { GitWorktree } from "@/types/git";

/** Branch-bound rows only — matches backend gitWorktreeIsFullyRegistered. */
export function isFullyRegisteredWorktree(worktree: GitWorktree): boolean {
  return Boolean(worktree.branch_id?.trim());
}

/** Rows shown on /worktrees: operator-registered linked worktrees (not repo checkout stubs). */
export function isLinkedWorktreeForDisplay(worktree: GitWorktree): boolean {
  return isFullyRegisteredWorktree(worktree) && !worktree.is_main;
}

/** Rows shown on /worktrees/:repositoryId detail list (includes primary checkout). */
export function isDetailPageWorktree(worktree: GitWorktree): boolean {
  return isFullyRegisteredWorktree(worktree);
}

export function sortDetailPageWorktrees(worktrees: GitWorktree[]): GitWorktree[] {
  return [...worktrees].sort(
    (a, b) =>
      Number(b.is_main) - Number(a.is_main) ||
      a.name.localeCompare(b.name, undefined, { sensitivity: "base" }),
  );
}

/** Default worktree when compose payloads omit worktree_id (main checkout first). */
export function pickDefaultWorktreeId(worktrees: GitWorktree[]): string {
  const registered = worktrees.filter(isFullyRegisteredWorktree);
  if (registered.length === 0) return "";
  return sortDetailPageWorktrees(registered)[0]!.id;
}
