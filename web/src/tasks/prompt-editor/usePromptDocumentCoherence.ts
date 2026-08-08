import { useCallback, useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { invalidatePromptDocumentCoherence } from "@/tasks/mutations";
import { isPromptSourceKind } from "./promptDocumentAdapter";

/**
 * Keeps the lists that render the open document coherent with what the editor
 * wrote. The drafts query is mounted above this route, so navigating back to it
 * never remounts and never refetches on its own — an explicit invalidate is the
 * only thing that refreshes it.
 *
 * Returns an invalidate callback for writes the user must see immediately
 * (a rename), and invalidates on unmount so every exit — Done, Escape, or
 * browser Back — settles the body autosaves too. Unmount invalidates
 * unconditionally: one refetch on a user-initiated navigation is cheaper than
 * tracking write state across every save path and getting it wrong.
 */
export function usePromptDocumentCoherence(
  sourceKind: string,
  sourceId: string,
): () => void {
  const queryClient = useQueryClient();
  const invalidate = useCallback(() => {
    if (!isPromptSourceKind(sourceKind)) return;
    invalidatePromptDocumentCoherence(queryClient, sourceKind, sourceId);
  }, [queryClient, sourceId, sourceKind]);

  useEffect(() => {
    return () => invalidate();
  }, [invalidate]);

  return invalidate;
}
