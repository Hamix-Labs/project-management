import { useCallback, useEffect, useRef, useState } from "react";
import { listGlobalGitWorktrees } from "@/api/gitGlobal";

type WorktreeCandidate = {
  id: string;
  is_main: boolean;
};

/** Why no worktree is available, when one could not be produced. */
export type FileWorktreeGap =
  /** Neither a task worktree nor a repository to fall back to. */
  | "unbound"
  /** Repository answered, but exposes no main worktree row. */
  | "no-main-worktree"
  /** The repository worktree lookup itself failed. */
  | "lookup-failed";

export type FileWorktreeResolution = {
  worktreeId?: string;
  resolving: boolean;
  gap?: FileWorktreeGap;
  /**
   * Resolves once the worktree is known, so a consumer triggered before the
   * lookup settles can await it instead of reporting "no worktree".
   */
  whenResolved: () => Promise<string | undefined>;
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

type ResolvedState = {
  repositoryId?: string;
  worktreeId?: string;
  resolving: boolean;
  gap?: FileWorktreeGap;
};

export function usePromptEditorFileWorktree({
  worktreeId,
  repositoryId,
}: {
  worktreeId?: string;
  repositoryId?: string;
}): FileWorktreeResolution {
  const explicitWorktreeId = worktreeId?.trim() || undefined;
  const repoId = repositoryId?.trim() || undefined;
  const [resolved, setResolved] = useState<ResolvedState>({ resolving: false });
  // Holds the in-flight lookup so `whenResolved` can await the same request
  // rather than starting a second one.
  const pendingRef = useRef<Promise<string | undefined> | null>(null);
  const explicitRef = useRef(explicitWorktreeId);
  explicitRef.current = explicitWorktreeId;

  useEffect(() => {
    if (explicitWorktreeId || !repoId) {
      pendingRef.current = null;
      setResolved({
        repositoryId: repoId,
        resolving: false,
        gap: explicitWorktreeId ? undefined : "unbound",
      });
      return;
    }

    const ac = new AbortController();
    setResolved({ repositoryId: repoId, resolving: true });

    const lookup = listGlobalGitWorktrees(repoId, { signal: ac.signal })
      .then((worktrees) => {
        const nextWorktreeId = resolvePromptEditorFileWorktree(
          undefined,
          worktrees,
        );
        if (!ac.signal.aborted) {
          setResolved({
            repositoryId: repoId,
            worktreeId: nextWorktreeId,
            resolving: false,
            gap: nextWorktreeId ? undefined : "no-main-worktree",
          });
        }
        return nextWorktreeId;
      })
      .catch(() => {
        if (!ac.signal.aborted) {
          setResolved({
            repositoryId: repoId,
            resolving: false,
            gap: "lookup-failed",
          });
        }
        return undefined;
      });

    pendingRef.current = lookup;

    return () => {
      ac.abort();
    };
  }, [explicitWorktreeId, repoId]);

  const whenResolved = useCallback(async () => {
    const explicit = explicitRef.current;
    if (explicit) return explicit;
    return (await pendingRef.current) ?? undefined;
  }, []);

  if (explicitWorktreeId) {
    return { worktreeId: explicitWorktreeId, resolving: false, whenResolved };
  }

  if (!repoId) {
    return { resolving: false, gap: "unbound", whenResolved };
  }

  // A repository change re-runs the effect on the next commit; until then the
  // stored result still describes the previous repository.
  const settledForRepo = resolved.repositoryId === repoId;

  return {
    worktreeId: settledForRepo ? resolved.worktreeId : undefined,
    resolving: !settledForRepo || resolved.resolving,
    gap: settledForRepo ? resolved.gap : undefined,
    whenResolved,
  };
}
