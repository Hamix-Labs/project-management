import { useCallback, useEffect, useRef, useState } from "react";
import {
  deriveEditedLabel,
  deriveSaveStatus,
  deriveWordCountLabel,
  pickLoadError,
  pickSaveError,
} from "./promptEditorPageViewModel";
import { usePromptEditorAutosave } from "./usePromptEditorAutosave";
import {
  usePromptEditorDocumentLoad,
  type PromptEditorLoadStatus,
} from "./usePromptEditorDocumentLoad";
import { usePromptEditorLeave } from "./usePromptEditorLeave";
import { usePromptEditorRouteAdapter } from "./usePromptEditorRouteAdapter";
import {
  HYDRATE_FALLBACK_WARNING,
  type PromptEditorSessionError,
} from "./promptEditorSessionError";

export function usePromptEditorPageController() {
  const { sourceKind, sourceId, kindOk, launch, adapter } =
    usePromptEditorRouteAdapter();

  const [status, setStatus] = useState<PromptEditorLoadStatus>("loading");
  const [html, setHtml] = useState("");
  const [sessionError, setSessionError] =
    useState<PromptEditorSessionError | null>(null);
  const [hydrateWarning, setHydrateWarning] =
    useState<PromptEditorSessionError | null>(null);
  const [saving, setSaving] = useState(false);
  const [dirty, setDirty] = useState(false);
  const [lastSavedAt, setLastSavedAt] = useState<number | null>(null);
  const [tick, setTick] = useState(0);
  const [loadNonce, setLoadNonce] = useState(0);
  const [adapterWorktreeId, setAdapterWorktreeId] = useState<
    string | undefined
  >(undefined);
  const htmlRef = useRef(html);
  htmlRef.current = html;
  const dirtyRef = useRef(false);

  const worktreeId =
    launch?.worktreeId?.trim() || adapterWorktreeId || undefined;
  const title = launch?.title?.trim() || "Untitled task";

  const onResolvedWorktreeId = useCallback((id: string | undefined) => {
    setAdapterWorktreeId(id?.trim() || undefined);
  }, []);

  const onCommit = useCallback((snap: { html: string }) => {
    if (!dirtyRef.current) setHtml(snap.html);
    setSessionError(null);
    setHydrateWarning(null);
    if (!dirtyRef.current) setLastSavedAt(Date.now());
  }, []);

  const onLoadError = useCallback((err: PromptEditorSessionError) => {
    setSessionError(err);
  }, []);

  const onStatus = useCallback((next: PromptEditorLoadStatus) => {
    setStatus(next);
  }, []);

  const { repoLabel } = usePromptEditorDocumentLoad({
    adapter,
    launch,
    loadNonce,
    dirtyRef,
    onCommit,
    onLoadError,
    onStatus,
    worktreeId,
    onResolvedWorktreeId,
  });

  useEffect(() => {
    const id = window.setInterval(() => setTick((t) => t + 1), 30_000);
    return () => window.clearInterval(id);
  }, []);

  const { retrySave } = usePromptEditorAutosave({
    adapter,
    status,
    html,
    htmlRef,
    dirtyRef,
    setDirty,
    setSaving,
    setSessionError,
    setLastSavedAt,
  });

  const onChange = useCallback((next: string) => {
    dirtyRef.current = true;
    setDirty(true);
    setHtml(next);
  }, []);

  const onHydrateFallback = useCallback(() => {
    setHydrateWarning(HYDRATE_FALLBACK_WARNING);
  }, []);

  const { leaveEditor, leaveWithoutSave, leavePending } = usePromptEditorLeave({
    adapter,
    launch,
    htmlRef,
    dirtyRef,
    setSessionError,
    setLastSavedAt,
    setSaving,
  });

  const saveError = pickSaveError(sessionError);
  const loadError = pickLoadError(status, sessionError);
  const ready = status === "ready";
  void tick;

  return {
    kindOk,
    sourceId,
    sourceKind,
    launch,
    html,
    status,
    ready,
    loadError,
    saveError,
    hydrateWarning,
    dismissHydrateWarning: () => setHydrateWarning(null),
    saving,
    leavePending,
    saveStatus: deriveSaveStatus({
      saveError,
      saving,
      leavePending,
      dirty,
    }),
    title,
    editedLabel: deriveEditedLabel(status, ready, lastSavedAt),
    wordCountLabel: deriveWordCountLabel(ready, html),
    repoLabel: ready ? repoLabel : "—",
    worktreeId,
    onChange,
    onHydrateFallback,
    leaveEditor,
    leaveWithoutSave,
    retryLoad: () => {
      setSessionError(null);
      setLoadNonce((n) => n + 1);
    },
    retrySave,
  };
}
