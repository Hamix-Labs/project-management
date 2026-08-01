import { getGlobalGitWorktree } from "@/api/gitGlobal";
import type { GitContextFields } from "@/tasks/components/task-detail/commits/commitDisplay";

/**
 * Resolves a task's `worktree_id` into repo/worktree/branch paths for display.
 * Uses O(1) GET /git/worktrees/{id} (no full-list scan).
 */
export async function resolveTaskGitBinding(
  worktreeId: string,
  options?: { signal?: AbortSignal },
): Promise<GitContextFields | null> {
  const wtId = worktreeId.trim();
  if (wtId === "") {
    return null;
  }

  try {
    const detail = await getGlobalGitWorktree(wtId, { signal: options?.signal });
    const openPath = detail.host_path.trim() || detail.path;
    return {
      repo: detail.repository_path,
      worktree: detail.path,
      openPath,
      branch: detail.branch_name,
    };
  } catch (err) {
    if (options?.signal?.aborted) {
      throw err;
    }
    return null;
  }
}
