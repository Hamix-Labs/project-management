import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import type { PromptEditorSaveStatusKind } from "@/components/prompt-editor/chrome";
import {
  createPromptDocumentAdapter,
  isPromptSourceKind,
} from "./promptDocumentAdapter";
import { readPromptEditorLaunch } from "./promptEditorSession";
import {
  formatRelativeEdited,
  wordCountFromHtml,
} from "./promptEditorPageMeta";
import { usePromptEditorDocumentLoad } from "./usePromptEditorDocumentLoad";
import { usePromptEditorLeave } from "./usePromptEditorLeave";
import type { PromptEditorLaunchContext } from "./types";

const AUTOSAVE_MS = 800;

export function usePromptEditorPageController() {
  const { sourceKind = "", sourceId = "" } = useParams<{
    sourceKind: string;
    sourceId: string;
  }>();
  const launchRef = useRef<PromptEditorLaunchContext | null>(null);
  if (launchRef.current === null) {
    launchRef.current = readPromptEditorLaunch();
  }
  const launch = launchRef.current;

  const kindOk = isPromptSourceKind(sourceKind);
  const adapter = useMemo(() => {
    if (!kindOk || !sourceId) return null;
    return createPromptDocumentAdapter(
      { kind: sourceKind, id: sourceId },
      launch ?? undefined,
    );
  }, [kindOk, sourceKind, sourceId, launch]);

  const [html, setHtml] = useState(launch?.seedHtml ?? "");
  const [loaded, setLoaded] = useState(Boolean(launch?.seedHtml));
  const [loadError, setLoadError] = useState<string | null>(null);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [dirty, setDirty] = useState(false);
  const [lastSavedAt, setLastSavedAt] = useState<number | null>(
    launch?.seedHtml ? Date.now() : null,
  );
  const [tick, setTick] = useState(0);
  const [adapterWorktreeId, setAdapterWorktreeId] = useState<
    string | undefined
  >(undefined);
  const htmlRef = useRef(html);
  htmlRef.current = html;
  const dirtyRef = useRef(false);

  const launchWorktreeId = launch?.worktreeId?.trim() || undefined;
  const worktreeId = launchWorktreeId || adapterWorktreeId;
  const title = launch?.title?.trim() || "Untitled task";

  const onResolvedWorktreeId = useCallback((id: string | undefined) => {
    setAdapterWorktreeId(id?.trim() || undefined);
  }, []);

  const { repoLabel } = usePromptEditorDocumentLoad({
    adapter,
    launch,
    dirtyRef,
    setHtml,
    setLoaded,
    setLoadError,
    setLastSavedAt,
    worktreeId,
    onResolvedWorktreeId,
  });

  useEffect(() => {
    const id = window.setInterval(() => setTick((t) => t + 1), 30_000);
    return () => window.clearInterval(id);
  }, []);

  const flushSave = useCallback(async () => {
    if (!adapter) return;
    setSaving(true);
    try {
      await adapter.save(htmlRef.current);
      dirtyRef.current = false;
      setDirty(false);
      setSaveError(null);
      setLastSavedAt(Date.now());
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : String(err));
      throw err;
    } finally {
      setSaving(false);
    }
  }, [adapter]);

  useEffect(() => {
    if (!adapter || !loaded || !dirtyRef.current) return;
    const t = window.setTimeout(() => {
      void flushSave().catch(() => undefined);
    }, AUTOSAVE_MS);
    return () => window.clearTimeout(t);
  }, [html, adapter, loaded, flushSave]);

  const onChange = useCallback((next: string) => {
    dirtyRef.current = true;
    setDirty(true);
    setHtml(next);
  }, []);

  const { leaveEditor, leavePending } = usePromptEditorLeave({
    adapter,
    launch,
    htmlRef,
    dirtyRef,
    setSaveError,
    setLastSavedAt,
    setSaving,
  });

  const saveStatus: PromptEditorSaveStatusKind = saveError
    ? "error"
    : saving || leavePending
      ? "saving"
      : dirty
        ? "unsaved"
        : "saved";

  const words = wordCountFromHtml(html);
  void tick;

  return {
    kindOk,
    sourceId,
    sourceKind,
    launch,
    html,
    loaded,
    loadError,
    saveError,
    saving,
    leavePending,
    saveStatus,
    title,
    editedLabel: formatRelativeEdited(lastSavedAt),
    wordCountLabel: words === 0 ? "0 words" : `~${words} words`,
    repoLabel,
    worktreeId,
    onChange,
    leaveEditor,
    retrySave: () => {
      void flushSave().catch(() => undefined);
    },
  };
}
