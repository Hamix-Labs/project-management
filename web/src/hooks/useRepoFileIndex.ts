import { useEffect, useState } from "react";
import {
  clearRepoFileIndex,
  getRepoFileIndexSnapshot,
  subscribeRepoFileIndex,
  warmRepoFileIndex,
  type RepoFileIndexSnapshot,
} from "@/lib/repoFileIndex";

/**
 * Keeps a gitignore-aware file path index warm for `@` mentions.
 * Starts fetching pages as soon as worktreeId is set.
 */
export function useRepoFileIndex(worktreeId: string): RepoFileIndexSnapshot {
  const id = worktreeId.trim();
  const [snap, setSnap] = useState(() => getRepoFileIndexSnapshot(id));

  useEffect(() => {
    if (id === "") {
      setSnap(getRepoFileIndexSnapshot(""));
      return;
    }
    setSnap(getRepoFileIndexSnapshot(id));
    const unsub = subscribeRepoFileIndex(id, () => {
      setSnap(getRepoFileIndexSnapshot(id));
    });
    warmRepoFileIndex(id);
    return () => {
      unsub();
    };
  }, [id]);

  useEffect(() => {
    return () => {
      // Keep ready indexes across remounts within the same session; only clear
      // when the hook unmounts with a worktree change handled by the next effect.
    };
  }, []);

  return snap;
}

/** Clear indexes when leaving create/edit surfaces that owned the warm. */
export function disposeRepoFileIndex(worktreeId?: string): void {
  clearRepoFileIndex(worktreeId);
}
