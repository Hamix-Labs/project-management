import { useEffect, useState } from "react";
import { getGlobalGitWorktree } from "@/api/gitGlobal";
import type { PromptDocumentAdapter } from "./types";
import type { PromptEditorLaunchContext } from "./types";
import { repoBasename } from "./promptEditorPageMeta";
import {
  loadSessionError,
  type PromptEditorSessionError,
} from "./promptEditorSessionError";

export type PromptEditorLoadStatus = "loading" | "ready" | "error";

type CommitSnapshot = {
  html: string;
  worktreeId?: string;
};

type Args = {
  adapter: PromptDocumentAdapter | null;
  launch: PromptEditorLaunchContext | null;
  /** Bump to re-run load (Retry). */
  loadNonce: number;
  dirtyRef: React.MutableRefObject<boolean>;
  onCommit: (snap: CommitSnapshot) => void;
  onLoadError: (err: PromptEditorSessionError) => void;
  onStatus: (status: PromptEditorLoadStatus) => void;
  worktreeId?: string;
  onResolvedWorktreeId?: (id: string | undefined) => void;
};

/**
 * Loads the prompt snapshot with a cancelled-flag generation so in-flight
 * results never leave the session stuck in loading after abort/remount.
 */
export function usePromptEditorDocumentLoad({
  adapter,
  launch,
  loadNonce,
  dirtyRef,
  onCommit,
  onLoadError,
  onStatus,
  worktreeId,
  onResolvedWorktreeId,
}: Args) {
  const [repoLabel, setRepoLabel] = useState("No repo");

  useEffect(() => {
    if (!adapter) {
      onStatus("error");
      onLoadError(
        loadSessionError(new Error("Unknown or missing prompt document.")),
      );
      return;
    }

    let cancelled = false;
    onStatus("loading");

    void (async () => {
      try {
        const snap = await adapter.load();
        if (cancelled) return;
        const html =
          launch?.seedHtml !== undefined && launch.seedHtml !== ""
            ? launch.seedHtml
            : snap.html;
        const fromSnap = snap.worktreeId?.trim() || undefined;
        onResolvedWorktreeId?.(fromSnap);
        onCommit({ html, worktreeId: fromSnap });
        if (!dirtyRef.current) {
          // parent sets lastSavedAt on commit
        }
        onStatus("ready");
      } catch (err) {
        if (cancelled) return;
        onLoadError(loadSessionError(err));
        onStatus("error");
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [
    adapter,
    dirtyRef,
    launch?.seedHtml,
    loadNonce,
    onCommit,
    onLoadError,
    onResolvedWorktreeId,
    onStatus,
  ]);

  useEffect(() => {
    const wt = worktreeId;
    if (!wt) {
      setRepoLabel("No repo");
      return;
    }
    let cancelled = false;
    void getGlobalGitWorktree(wt)
      .then((detail) => {
        if (cancelled) return;
        const path = detail.repository_path || detail.repository_host_path || "";
        setRepoLabel(path ? repoBasename(path) : "No repo");
      })
      .catch(() => {
        if (!cancelled) setRepoLabel("No repo");
      });
    return () => {
      cancelled = true;
    };
  }, [worktreeId]);

  return { repoLabel };
}
