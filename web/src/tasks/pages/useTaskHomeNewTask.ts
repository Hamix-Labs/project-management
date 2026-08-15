import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useTasksAppContext } from "../app/TasksAppProvider";
import { tasksNewPath } from "../composeRoutes";
import { decideCreateEntry } from "../create/decideCreateEntry";

export function useTaskHomeNewTask() {
  const app = useTasksAppContext();
  const navigate = useNavigate();
  const [resumeOpen, setResumeOpen] = useState(false);
  const [awaitingDrafts, setAwaitingDrafts] = useState(false);

  const applyDecision = useCallback(() => {
    const decision = decideCreateEntry({
      isPending: app.draftListLoading,
      isError: Boolean(app.draftListError),
      errorMessage: app.draftListError,
      draftCount: app.taskDrafts.length,
    });
    if (decision.kind === "wait") return false;
    if (decision.kind === "showPicker") {
      setResumeOpen(true);
    } else {
      setResumeOpen(false);
      navigate(tasksNewPath());
    }
    return true;
  }, [
    app.draftListError,
    app.draftListLoading,
    app.taskDrafts.length,
    navigate,
  ]);

  const openCreate = useCallback(() => {
    if (!applyDecision()) setAwaitingDrafts(true);
  }, [applyDecision]);

  useEffect(() => {
    if (!awaitingDrafts) return;
    if (applyDecision()) setAwaitingDrafts(false);
  }, [awaitingDrafts, applyDecision]);

  const closeResume = useCallback(() => {
    setResumeOpen(false);
  }, []);

  const startFresh = useCallback(() => {
    setResumeOpen(false);
    navigate(tasksNewPath());
  }, [navigate]);

  const resumeDraft = useCallback(
    (draftId: string) => {
      setResumeOpen(false);
      navigate(tasksNewPath({ draft: draftId }));
    },
    [navigate],
  );

  return {
    openCreate,
    awaitingDrafts,
    resumeOpen,
    closeResume,
    startFresh,
    resumeDraft,
    createBlocked: app.createModalOpen || awaitingDrafts || resumeOpen,
    drafts: app.taskDrafts,
    draftListError: app.draftListError,
    retryDraftList: app.retryDraftList,
    resumePending: app.resumeDraftPending,
    resumeError: app.resumeDraftError,
  };
}
