import { useEffect, useState } from "react";
import { resolveMentionWorktreeId } from "@/lib/resolveMentionWorktreeId";

/**
 * Resolves the worktree id used for `@` mentions: explicit worktree, else repo main.
 * Returns "" while resolving or when neither binding is available.
 */
export function useResolvedMentionWorktreeId(
  worktreeId?: string,
  repositoryId?: string,
): string {
  const explicit = worktreeId?.trim() ?? "";
  const repoId = repositoryId?.trim() ?? "";
  const [resolved, setResolved] = useState(() => explicit);

  useEffect(() => {
    if (explicit !== "") {
      setResolved(explicit);
      return;
    }
    if (repoId === "") {
      setResolved("");
      return;
    }

    const ac = new AbortController();
    setResolved("");
    void resolveMentionWorktreeId({
      repositoryId: repoId,
      signal: ac.signal,
    }).then((id) => {
      if (ac.signal.aborted) return;
      setResolved(id ?? "");
    });
    return () => {
      ac.abort();
    };
  }, [explicit, repoId]);

  return resolved;
}
