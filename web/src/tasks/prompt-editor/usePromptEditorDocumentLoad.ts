import { useEffect, useState } from "react";
import { getGlobalGitWorktree } from "@/api/gitGlobal";
import type { PromptDocumentAdapter } from "./types";
import type { PromptEditorLaunchContext } from "./types";
import { repoBasename } from "./promptEditorPageMeta";

type Args = {
  adapter: PromptDocumentAdapter | null;
  launch: PromptEditorLaunchContext | null;
  dirtyRef: React.MutableRefObject<boolean>;
  setHtml: (html: string) => void;
  setLoaded: (v: boolean) => void;
  setLoadError: (err: string | null) => void;
  setLastSavedAt: (at: number) => void;
  /** Prefer launch worktree; fall back to adapter snapshot. */
  worktreeId?: string;
  onResolvedWorktreeId?: (id: string | undefined) => void;
};

export function usePromptEditorDocumentLoad({
  adapter,
  launch,
  dirtyRef,
  setHtml,
  setLoaded,
  setLoadError,
  setLastSavedAt,
  worktreeId,
  onResolvedWorktreeId,
}: Args) {
  const [repoLabel, setRepoLabel] = useState("No repo");

  useEffect(() => {
    if (!adapter) return;
    const ac = new AbortController();
    void (async () => {
      try {
        const snap = await adapter.load(ac.signal);
        if (ac.signal.aborted) return;
        if (launch?.seedHtml !== undefined && launch.seedHtml !== "") {
          setHtml(launch.seedHtml);
        } else {
          setHtml(snap.html);
        }
        const fromSnap = snap.worktreeId?.trim() || undefined;
        onResolvedWorktreeId?.(fromSnap);
        setLoaded(true);
        if (!dirtyRef.current) setLastSavedAt(Date.now());
      } catch (err) {
        if (ac.signal.aborted) return;
        setLoadError(err instanceof Error ? err.message : String(err));
        setLoaded(true);
      }
    })();
    return () => ac.abort();
  }, [
    adapter,
    dirtyRef,
    launch?.seedHtml,
    onResolvedWorktreeId,
    setHtml,
    setLastSavedAt,
    setLoadError,
    setLoaded,
  ]);

  useEffect(() => {
    const wt = worktreeId;
    if (!wt) {
      setRepoLabel("No repo");
      return;
    }
    const ac = new AbortController();
    void getGlobalGitWorktree(wt, { signal: ac.signal })
      .then((detail) => {
        if (ac.signal.aborted) return;
        const path = detail.repository_path || detail.repository_host_path || "";
        setRepoLabel(path ? repoBasename(path) : "No repo");
      })
      .catch(() => {
        if (!ac.signal.aborted) setRepoLabel("No repo");
      });
    return () => ac.abort();
  }, [worktreeId]);

  return { repoLabel };
}
