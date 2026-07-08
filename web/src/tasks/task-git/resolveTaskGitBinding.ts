import {
  listGlobalGitBranches,
  listGlobalGitWorktrees,
} from "@/api/gitGlobal";
import type { GitContextFields } from "@/tasks/components/task-detail/commits/commitDisplay";
import type { GitRepository } from "@/types/git";

function orderedRepositoryIds(
  repositories: GitRepository[],
  repositoryIdHint?: string,
): string[] {
  const all = repositories.map((repository) => repository.id);
  const hint = repositoryIdHint?.trim();
  if (!hint) {
    return all;
  }
  if (!all.includes(hint)) {
    return [hint, ...all];
  }
  return [hint, ...all.filter((id) => id !== hint)];
}

/**
 * Resolves a task's `worktree_id` into repo/worktree/branch paths for display.
 * Scans registered repositories until the worktree row is found.
 */
export async function resolveTaskGitBinding(
  worktreeId: string,
  repositories: GitRepository[],
  options?: { repositoryIdHint?: string; signal?: AbortSignal },
): Promise<GitContextFields | null> {
  const wtId = worktreeId.trim();
  if (wtId === "" || repositories.length === 0) {
    return null;
  }

  const repoById = new Map(repositories.map((repository) => [repository.id, repository]));

  for (const repositoryId of orderedRepositoryIds(
    repositories,
    options?.repositoryIdHint,
  )) {
    const worktrees = await listGlobalGitWorktrees(repositoryId, {
      signal: options?.signal,
    });
    const worktree = worktrees.find((row) => row.id === wtId);
    if (!worktree) {
      continue;
    }

    const repository = repoById.get(repositoryId);
    const branches = await listGlobalGitBranches(repositoryId, {
      signal: options?.signal,
    });
    const branch = worktree.branch_id
      ? branches.find((row) => row.id === worktree.branch_id)
      : undefined;

    return {
      repo: repository?.path ?? "",
      worktree: worktree.path,
      branch: branch?.name ?? "",
    };
  }

  return null;
}
