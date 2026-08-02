import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import type { PromptDocumentAdapter } from "./types";
import type { PromptEditorLaunchContext } from "./types";
import {
  clearPromptEditorLaunch,
  writePromptEditorReturn,
} from "./promptEditorSession";

type Args = {
  adapter: PromptDocumentAdapter | null;
  launch: PromptEditorLaunchContext | null;
  htmlRef: React.MutableRefObject<string>;
  dirtyRef: React.MutableRefObject<boolean>;
  setSaveError: (err: string | null) => void;
  setLastSavedAt: (at: number) => void;
  setSaving: (v: boolean) => void;
};

/** Leave / Escape flush; `beforeunload` when dirty. Avoids `useBlocker` (needs a data router). */
export function usePromptEditorLeave({
  adapter,
  launch,
  htmlRef,
  dirtyRef,
  setSaveError,
  setLastSavedAt,
  setSaving,
}: Args) {
  const navigate = useNavigate();
  const [leavePending, setLeavePending] = useState(false);
  const leavingRef = useRef(false);

  const leaveEditor = useCallback(async () => {
    if (!adapter || leavingRef.current) return;
    leavingRef.current = true;
    setLeavePending(true);
    setSaving(true);
    try {
      await adapter.save(htmlRef.current);
      dirtyRef.current = false;
      setSaveError(null);
      setLastSavedAt(Date.now());
      const returnPath = launch?.returnPath ?? "/";
      writePromptEditorReturn({
        resumeCompose: Boolean(launch?.resumeCompose),
        resumePolish: Boolean(launch?.resumePolish),
        polishTaskId: launch?.polishTaskId,
        returnPath,
        html: htmlRef.current,
      });
      clearPromptEditorLaunch();
      navigate(returnPath);
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : String(err));
      leavingRef.current = false;
    } finally {
      setSaving(false);
      setLeavePending(false);
    }
  }, [
    adapter,
    dirtyRef,
    htmlRef,
    launch,
    navigate,
    setLastSavedAt,
    setSaveError,
    setSaving,
  ]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== "Escape") return;
      const target = e.target as HTMLElement | null;
      if (
        target?.closest?.(
          "[role='listbox'], [role='dialog'], .bn-suggestion-menu, .prompt-editor-mention-menu",
        )
      ) {
        return;
      }
      e.preventDefault();
      void leaveEditor();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [leaveEditor]);

  useEffect(() => {
    const onBeforeUnload = (e: BeforeUnloadEvent) => {
      if (leavingRef.current || !dirtyRef.current) return;
      e.preventDefault();
      e.returnValue = "";
    };
    window.addEventListener("beforeunload", onBeforeUnload);
    return () => window.removeEventListener("beforeunload", onBeforeUnload);
  }, [dirtyRef]);

  return { leaveEditor, leavePending };
}
