import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import type { PromptDocumentAdapter } from "./types";
import type { PromptEditorLaunchContext } from "./types";
import {
  clearPromptEditorLaunch,
  writePromptEditorReturn,
} from "./promptEditorSession";
import {
  leaveSaveSessionError,
  type PromptEditorSessionError,
} from "./promptEditorSessionError";

type Args = {
  adapter: PromptDocumentAdapter | null;
  launch: PromptEditorLaunchContext | null;
  htmlRef: React.MutableRefObject<string>;
  titleRef: React.MutableRefObject<string>;
  dirtyRef: React.MutableRefObject<boolean>;
  setSessionError: (err: PromptEditorSessionError | null) => void;
  setLastSavedAt: (at: number) => void;
  setSaving: (v: boolean) => void;
};

/** Leave / Escape flush; `beforeunload` when dirty. Avoids `useBlocker` (needs a data router). */
export function usePromptEditorLeave({
  adapter,
  launch,
  htmlRef,
  titleRef,
  dirtyRef,
  setSessionError,
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
      setSessionError(null);
      setLastSavedAt(Date.now());
      const returnPath = launch?.returnPath ?? "/";
      writePromptEditorReturn({
        resumeCompose: Boolean(launch?.resumeCompose),
        resumePolish: Boolean(launch?.resumePolish),
        polishTaskId: launch?.polishTaskId,
        returnPath,
        html: htmlRef.current,
        title: titleRef.current,
      });
      clearPromptEditorLaunch();
      navigate(returnPath);
    } catch (err) {
      setSessionError(leaveSaveSessionError(err));
      leavingRef.current = false;
    } finally {
      setSaving(false);
      setLeavePending(false);
    }
  }, [
    adapter,
    dirtyRef,
    htmlRef,
    titleRef,
    launch,
    navigate,
    setLastSavedAt,
    setSessionError,
    setSaving,
  ]);

  /** Leave without saving — used from the load-error panel. */
  const leaveWithoutSave = useCallback(() => {
    const returnPath = launch?.returnPath ?? "/";
    clearPromptEditorLaunch();
    navigate(returnPath);
  }, [launch?.returnPath, navigate]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== "Escape") return;
      const target = e.target as HTMLElement | null;
      if (
        target?.closest?.(
          "[role='listbox'], [role='dialog'], .bn-suggestion-menu, .prompt-editor-mention-menu, .prompt-editor-doc-header__title",
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

  return { leaveEditor, leaveWithoutSave, leavePending };
}
