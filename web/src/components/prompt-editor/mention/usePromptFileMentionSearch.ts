import { useCallback, useEffect, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import type { FileWorktreeResolution } from "../usePromptEditorFileWorktree";
import type { PromptFileMentionItem } from "./PromptEditorMentionMenu";
import {
  runPromptFileMentionSearch,
  type BoundMentionStatus,
} from "./promptFileMentionRequest";
import type { MentionSearchStatus } from "./promptFileMentionStatus";

export type PromptFileMentionSearch = {
  /**
   * Stable across renders. BlockNote's `useLoadSuggestionMenuItems` lists this
   * function in a `useEffect` dependency array, so a fresh identity per render
   * restarts the load forever — and this hook writes status, which re-renders
   * the host on every response.
   */
  getItems: (query: string) => Promise<PromptFileMentionItem[]>;
  status: MentionSearchStatus;
};

export function usePromptFileMentionSearch({
  worktree,
  onSelectPath,
}: {
  worktree: FileWorktreeResolution;
  onSelectPath: (path: string) => void;
}): PromptFileMentionSearch {
  const queryClient = useQueryClient();
  const [bound, setBound] = useState<BoundMentionStatus>({
    worktreeId: undefined,
    value: { kind: "idle" },
  });

  // Everything `getItems` reads lives in a ref, so the callback needs no
  // dependencies and never changes identity.
  const worktreeRef = useRef(worktree);
  worktreeRef.current = worktree;
  const onSelectPathRef = useRef(onSelectPath);
  onSelectPathRef.current = onSelectPath;
  const queryClientRef = useRef(queryClient);
  queryClientRef.current = queryClient;

  // Only the newest request may write status; superseded ones stay silent.
  const generationRef = useRef(0);
  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  const getItems = useCallback(
    async (query: string): Promise<PromptFileMentionItem[]> => {
      const generation = ++generationRef.current;
      const write = (next: BoundMentionStatus) => {
        if (mountedRef.current && generationRef.current === generation) {
          setBound(next);
        }
      };

      const outcome = await runPromptFileMentionSearch({
        query,
        worktree: worktreeRef.current,
        queryClient: queryClientRef.current,
        onSelectPath: (path) => onSelectPathRef.current(path),
        onProgress: write,
      });
      write({ worktreeId: outcome.worktreeId, value: outcome.value });
      return outcome.items;
    },
    // Intentionally empty: identity stability is the contract this hook exists
    // to provide.
    [],
  );

  const status: MentionSearchStatus =
    bound.worktreeId === worktree.worktreeId ? bound.value : { kind: "idle" };

  return { getItems, status };
}
