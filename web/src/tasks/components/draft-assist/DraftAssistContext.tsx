import { createContext, useContext, useMemo } from "react";
import { useDraftAssistController } from "./useDraftAssistController";
import type {
  DraftAssistContextValue,
  DraftAssistProviderProps,
} from "./draftAssistTypes";

export type {
  DraftAssistContextValue,
  DraftAssistProviderProps,
  DraftAssistThreadMessage,
} from "./draftAssistTypes";

const DraftAssistCtx = createContext<DraftAssistContextValue | null>(null);

/** Consumer hook — throws when called outside a {@link DraftAssistProvider}. */
export function useDraftAssistContext(): DraftAssistContextValue {
  const ctx = useContext(DraftAssistCtx);
  if (!ctx) {
    throw new Error(
      "useDraftAssistContext must be rendered inside DraftAssistProvider",
    );
  }
  return ctx;
}

/** Optional consumer hook that returns null when no provider is above. */
export function useOptionalDraftAssistContext(): DraftAssistContextValue | null {
  return useContext(DraftAssistCtx);
}

/**
 * Provides the assist session, SSE stream, status machine, and message
 * list. Session creation is lazy via {@link DraftAssistContextValue.open}.
 */
export function DraftAssistProvider({
  getSnapshot,
  worktreeId,
  onApplyPromptPatch,
  getPromptSnapshot,
  children,
}: DraftAssistProviderProps) {
  const controller = useDraftAssistController({
    getSnapshot,
    worktreeId,
    onApplyPromptPatch,
    getPromptSnapshot,
  });
  const value = useMemo(() => controller, [controller]);

  return (
    <DraftAssistCtx.Provider value={value}>{children}</DraftAssistCtx.Provider>
  );
}
