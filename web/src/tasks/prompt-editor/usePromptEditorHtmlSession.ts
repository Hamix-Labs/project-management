import { useCallback, useEffect, useRef, useState } from "react";
import { usePromptEditorAutosave } from "./usePromptEditorAutosave";
import {
  usePromptEditorDocumentLoad,
  type PromptEditorLoadStatus,
} from "./usePromptEditorDocumentLoad";
import type { PromptDocumentAdapter, PromptEditorLaunchContext } from "./types";
import {
  HYDRATE_FALLBACK_WARNING,
  type PromptEditorSessionError,
} from "./promptEditorSessionError";

type Args = {
  adapter: PromptDocumentAdapter | null;
  launch: PromptEditorLaunchContext | null;
  applyHydratedName: (name?: string) => void;
};

/** HTML load/autosave session state for the Prompt Editor page. */
export function usePromptEditorHtmlSession({
  adapter,
  launch,
  applyHydratedName,
}: Args) {
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
  // Keep hydrate callback out of onCommit identity — an unstable parent
  // wrapper would retrigger document load forever (status stuck on loading).
  const applyHydratedNameRef = useRef(applyHydratedName);
  applyHydratedNameRef.current = applyHydratedName;

  const worktreeId =
    launch?.worktreeId?.trim() || adapterWorktreeId || undefined;

  const onResolvedWorktreeId = useCallback((id: string | undefined) => {
    setAdapterWorktreeId(id?.trim() || undefined);
  }, []);

  const onCommit = useCallback((snap: { html: string; name?: string }) => {
    setHtml(snap.html);
    applyHydratedNameRef.current(snap.name);
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
  void tick;

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

  const retryLoad = useCallback(() => {
    setSessionError(null);
    setLoadNonce((n) => n + 1);
  }, []);

  const dismissHydrateWarning = useCallback(() => {
    setHydrateWarning(null);
  }, []);

  return {
    status,
    html,
    htmlRef,
    dirtyRef,
    sessionError,
    setSessionError,
    hydrateWarning,
    dismissHydrateWarning,
    saving,
    setSaving,
    dirty,
    lastSavedAt,
    setLastSavedAt,
    repoLabel,
    worktreeId,
    onChange,
    onHydrateFallback,
    retryLoad,
    retrySave,
  };
}
