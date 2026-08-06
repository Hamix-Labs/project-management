import { useRef } from "react";
import {
  buildPromptEditorChromeLabels,
  pickLoadError,
  pickSaveError,
} from "./promptEditorPageViewModel";
import { usePromptEditorHtmlSession } from "./usePromptEditorHtmlSession";
import { usePromptEditorLeave } from "./usePromptEditorLeave";
import { usePromptEditorRouteAdapter } from "./usePromptEditorRouteAdapter";
import { usePromptEditorTitle } from "./usePromptEditorTitle";

export function usePromptEditorPageController() {
  const { sourceKind, sourceId, kindOk, launch, adapter } =
    usePromptEditorRouteAdapter();

  const applyHydratedNameRef = useRef<(name?: string) => void>(() => {});

  const htmlSession = usePromptEditorHtmlSession({
    adapter,
    launch,
    applyHydratedName: (name) => applyHydratedNameRef.current(name),
  });

  const titleState = usePromptEditorTitle({
    launchTitle: launch?.title,
    adapter,
    sourceKind,
    sourceId,
    setSessionError: htmlSession.setSessionError,
  });
  applyHydratedNameRef.current = titleState.applyHydratedName;

  const { leaveEditor, leaveWithoutSave, leavePending } = usePromptEditorLeave({
    adapter,
    launch,
    htmlRef: htmlSession.htmlRef,
    titleRef: titleState.titleRef,
    dirtyRef: htmlSession.dirtyRef,
    setSessionError: htmlSession.setSessionError,
    setLastSavedAt: htmlSession.setLastSavedAt,
    setSaving: htmlSession.setSaving,
  });

  const saveError = pickSaveError(htmlSession.sessionError);
  const loadError = pickLoadError(htmlSession.status, htmlSession.sessionError);
  const ready = htmlSession.status === "ready";
  const chrome = buildPromptEditorChromeLabels({
    status: htmlSession.status,
    ready,
    lastSavedAt: htmlSession.lastSavedAt,
    html: htmlSession.html,
    repoLabel: htmlSession.repoLabel,
    saveError,
    saving: htmlSession.saving,
    leavePending,
    dirty: htmlSession.dirty,
  });

  return {
    kindOk,
    sourceId,
    sourceKind,
    launch,
    html: htmlSession.html,
    status: htmlSession.status,
    ready,
    loadError,
    saveError,
    hydrateWarning: htmlSession.hydrateWarning,
    dismissHydrateWarning: htmlSession.dismissHydrateWarning,
    saving: htmlSession.saving,
    leavePending,
    ...chrome,
    title: titleState.title,
    onTitleCommit: titleState.onTitleCommit,
    worktreeId: htmlSession.worktreeId,
    onChange: htmlSession.onChange,
    onHydrateFallback: htmlSession.onHydrateFallback,
    leaveEditor,
    leaveWithoutSave,
    retryLoad: htmlSession.retryLoad,
    retrySave: htmlSession.retrySave,
  };
}
