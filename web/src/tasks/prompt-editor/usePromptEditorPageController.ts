import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  createPromptDocumentAdapter,
  isPromptSourceKind,
} from "./promptDocumentAdapter";
import {
  clearPromptEditorLaunch,
  readPromptEditorLaunch,
  writePromptEditorReturn,
} from "./promptEditorSession";
import type { PromptEditorLaunchContext } from "./types";

const AUTOSAVE_MS = 800;

export function usePromptEditorPageController() {
  const { sourceKind = "", sourceId = "" } = useParams<{
    sourceKind: string;
    sourceId: string;
  }>();
  const navigate = useNavigate();
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
  const [donePending, setDonePending] = useState(false);
  const htmlRef = useRef(html);
  htmlRef.current = html;
  const dirtyRef = useRef(false);

  useEffect(() => {
    if (!adapter) return;
    const ac = new AbortController();
    void (async () => {
      try {
        const snap = await adapter.load(ac.signal);
        if (ac.signal.aborted) return;
        if (launch?.seedHtml !== undefined && launch.seedHtml !== "") {
          setHtml(launch.seedHtml);
        } else {
          setHtml(snap.html);
        }
        setLoaded(true);
      } catch (err) {
        if (ac.signal.aborted) return;
        setLoadError(err instanceof Error ? err.message : String(err));
        setLoaded(true);
      }
    })();
    return () => ac.abort();
  }, [adapter, launch?.seedHtml]);

  useEffect(() => {
    if (!adapter || !loaded || !dirtyRef.current) return;
    const t = window.setTimeout(() => {
      setSaving(true);
      void adapter
        .save(htmlRef.current)
        .then(() => setSaveError(null))
        .catch((err: unknown) =>
          setSaveError(err instanceof Error ? err.message : String(err)),
        )
        .finally(() => setSaving(false));
    }, AUTOSAVE_MS);
    return () => window.clearTimeout(t);
  }, [html, adapter, loaded]);

  const onChange = useCallback((next: string) => {
    dirtyRef.current = true;
    setHtml(next);
  }, []);

  const onDone = useCallback(async () => {
    if (!adapter) return;
    setDonePending(true);
    try {
      await adapter.save(htmlRef.current);
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
    } finally {
      setDonePending(false);
    }
  }, [adapter, launch, navigate]);

  return {
    kindOk,
    sourceId,
    launch,
    html,
    loaded,
    loadError,
    saveError,
    saving,
    donePending,
    onChange,
    onDone,
  };
}
