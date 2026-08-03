import { useCallback, useEffect } from "react";
import type { PromptDocumentAdapter } from "./types";
import {
  saveSessionError,
  type PromptEditorSessionError,
} from "./promptEditorSessionError";
import type { PromptEditorLoadStatus } from "./usePromptEditorDocumentLoad";

const AUTOSAVE_MS = 800;

type Args = {
  adapter: PromptDocumentAdapter | null;
  status: PromptEditorLoadStatus;
  html: string;
  htmlRef: React.MutableRefObject<string>;
  dirtyRef: React.MutableRefObject<boolean>;
  setDirty: (v: boolean) => void;
  setSaving: (v: boolean) => void;
  setSessionError: (err: PromptEditorSessionError | null) => void;
  setLastSavedAt: (at: number) => void;
};

export function usePromptEditorAutosave({
  adapter,
  status,
  html,
  htmlRef,
  dirtyRef,
  setDirty,
  setSaving,
  setSessionError,
  setLastSavedAt,
}: Args) {
  const flushSave = useCallback(async () => {
    if (!adapter) return;
    setSaving(true);
    try {
      await adapter.save(htmlRef.current);
      dirtyRef.current = false;
      setDirty(false);
      setSessionError(null);
      setLastSavedAt(Date.now());
    } catch (err) {
      setSessionError(saveSessionError(err));
      throw err;
    } finally {
      setSaving(false);
    }
  }, [
    adapter,
    dirtyRef,
    htmlRef,
    setDirty,
    setLastSavedAt,
    setSaving,
    setSessionError,
  ]);

  useEffect(() => {
    if (!adapter || status !== "ready" || !dirtyRef.current) return;
    const t = window.setTimeout(() => {
      void flushSave().catch(() => undefined);
    }, AUTOSAVE_MS);
    return () => window.clearTimeout(t);
  }, [html, adapter, status, dirtyRef, flushSave]);

  const retrySave = useCallback(() => {
    void flushSave().catch(() => undefined);
  }, [flushSave]);

  return { flushSave, retrySave };
}
