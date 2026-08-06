import { useCallback, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import type { PromptDocumentAdapter } from "./types";
import { patchCachedTaskTitle } from "./patchCachedTaskTitle";
import {
  renameSessionError,
  type PromptEditorSessionError,
} from "./promptEditorSessionError";

type Args = {
  launchTitle?: string;
  adapter: PromptDocumentAdapter | null;
  sourceKind: string;
  sourceId: string;
  setSessionError: (err: PromptEditorSessionError | null) => void;
};

/**
 * Owns the Prompt IDE document title: hydrate from load, commit via adapter.saveName.
 * Title edits are independent of HTML dirty/autosave.
 */
export function usePromptEditorTitle({
  launchTitle,
  adapter,
  sourceKind,
  sourceId,
  setSessionError,
}: Args) {
  const queryClient = useQueryClient();
  const [title, setTitle] = useState(
    () => launchTitle?.trim() || "Untitled task",
  );
  const titleRef = useRef(title);
  titleRef.current = title;

  const applyHydratedName = useCallback((name?: string) => {
    const next = name?.trim();
    if (next) setTitle(next);
  }, []);

  const onTitleCommit = useCallback(
    async (next: string) => {
      const trimmed = next.trim();
      if (!adapter || !trimmed) return;
      if (trimmed === titleRef.current) return;
      const previous = titleRef.current;
      setTitle(trimmed);
      try {
        await adapter.saveName(trimmed);
        setSessionError(null);
        if (sourceKind === "task" && sourceId) {
          patchCachedTaskTitle(queryClient, sourceId, trimmed);
        }
      } catch (err) {
        setTitle(previous);
        setSessionError(renameSessionError(err));
      }
    },
    [adapter, queryClient, setSessionError, sourceId, sourceKind],
  );

  return { title, titleRef, onTitleCommit, applyHydratedName };
}
