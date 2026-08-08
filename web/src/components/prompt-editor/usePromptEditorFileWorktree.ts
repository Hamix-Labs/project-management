import { useEffect, useState } from "react";
import { listGlobalGitWorktrees } from "@/api/gitGlobal";

type WorktreeCandidate = {
  id: string;
  is_main: boolean;
};

type FileWorktreeResolution = {
  worktreeId?: string;
  resolving: boolean;
};

export function resolvePromptEditorFileWorktree(
  worktreeId: string | undefined,
  worktrees: WorktreeCandidate[] | undefined,
) {
  const explicitWorktreeId = worktreeId?.trim();
  if (explicitWorktreeId) {
    return explicitWorktreeId;
  }

  return worktrees?.find((worktree) => worktree.is_main)?.id;
}

export function usePromptEditorFileWorktree({
  worktreeId,
  repositoryId,
}: {
  worktreeId?: string;
  repositoryId?: string;
}): FileWorktreeResolution {
  const explicitWorktreeId = worktreeId?.trim() || undefined;
  const repoId = repositoryId?.trim() || undefined;
  const [resolved, setResolved] = useState<{
    repositoryId?: string;
    worktreeId?: string;
    resolving: boolean;
  }>({ resolving: false });

  useEffect(() => {
    if (explicitWorktreeId || !repoId) {
      setResolved({ repositoryId: repoId, resolving: false });
      return;
    }

    const ac = new AbortController();
    setResolved({ repositoryId: repoId, resolving: true });

    void listGlobalGitWorktrees(repoId, { signal: ac.signal })
      .then((worktrees) => {
        if (ac.signal.aborted) return;
        setResolved({
          repositoryId: repoId,
          worktreeId: resolvePromptEditorFileWorktree(undefined, worktrees),
          resolving: false,
        });
      })
      .catch(() => {
        if (ac.signal.aborted) return;
        setResolved({ repositoryId: repoId, resolving: false });
      });

    return () => {
      ac.abort();
    };
  }, [explicitWorktreeId, repoId]);

  if (explicitWorktreeId) {
    return { worktreeId: explicitWorktreeId, resolving: false };
  }

  if (!repoId) {
    return { resolving: false };
  }

  return {
    worktreeId:
      resolved.repositoryId === repoId ? resolved.worktreeId : undefined,
    resolving: resolved.repositoryId !== repoId || resolved.resolving,
  };
}
