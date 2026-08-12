import { listGlobalGitWorktrees } from "@/api/gitGlobal";

/**
 * Resolves which worktree `/repo/*` and `@` mentions should use.
 * Prefer an explicit task/template worktree; otherwise the repository's main checkout.
 */
export async function resolveMentionWorktreeId(input: {
  worktreeId?: string | null;
  repositoryId?: string | null;
  signal?: AbortSignal;
}): Promise<string | null> {
  const fromWorktree = input.worktreeId?.trim() ?? "";
  if (fromWorktree !== "") return fromWorktree;

  const repoId = input.repositoryId?.trim() ?? "";
  if (repoId === "") return null;

  try {
    const trees = await listGlobalGitWorktrees(repoId, { signal: input.signal });
    const main = trees.find((wt) => wt.is_main);
    return main?.id ?? null;
  } catch {
    return null;
  }
}
