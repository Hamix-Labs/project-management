import { useCallback, useEffect, useMemo, useRef, useState, type FormEvent } from "react";
import type { Task } from "@/types";
import { useTaskCloseFlow } from "./useTaskCloseFlow";
import { useTaskCreateFlow } from "./useTaskCreateFlow";
import { useTasksHomeList } from "./useTasksHomeList";
import { useTaskEditFlow } from "./useTaskEditFlow";

export type UseTasksAppOptions = {
  /** Whether the task change SSE stream is connected; owned by `App` via `useTaskEventStream`. */
  sseLive: boolean;
  /**
   * Whether the home-list / stats queries should be active. When the
   * user is on a route that does not consume `app.tasks` / `app.taskStats`
   * (e.g. `/settings`), passing `false` suspends the queries — the
   * cache remains populated for the next visit but no new GETs fire
   * while the route is mounted. Defaults to `true` so callers that
   * do not pass this stay on the historical eager-fetch behaviour.
   */
  dataEnabled?: boolean;
  /**
   * When false, settings/list/stats stay disabled until `useBootstrap`
   * settles so the aggregate seed can win over parallel GETs. Defaults
   * to `true` for harnesses that do not run bootstrap.
   */
  bootstrapSettled?: boolean;
};

/**
 * Thin facade: create flow + home list + edit/close composers.
 * Settings cache ownership lives in bootstrap / settings vertical (F-05-06).
 */
export function useTasksApp({
  sseLive,
  dataEnabled = true,
  bootstrapSettled = true,
}: UseTasksAppOptions) {
  const homeDataReady = dataEnabled && bootstrapSettled;
  const {
    createFlowError,
    editingTaskId,
    closeCreateModal,
    newTitle,
    newPrompt,
    newPriority,
    newProjectID,
    newTagsCsv,
    newMilestone,
    newTaskCursorModel,
    newSchedule,
    composeStatus,
    beginEditSession,
    createModalOpen,
    ...createFlow
  } = useTaskCreateFlow();

  const editingTaskIdRef = useRef<string | null>(null);
  editingTaskIdRef.current = editingTaskId;

  const {
    closeTarget,
    requestClose,
    cancelClose,
    confirmClose,
    closePending,
    closeError,
    closeSuccess,
    closeVariables,
    resetCloseError,
    reopen,
    reopenPending,
    reopenError,
    resetReopenError,
  } = useTaskCloseFlow({
    onClosed: (closedId) => {
      // Keep the compose modal open on close — closing a task should
      // never surprise-close an unrelated editor. Historically the
      // delete flow closed the modal because the task row vanished;
      // with close the row still exists so there is no orphan risk.
      if (editingTaskIdRef.current === closedId) {
        closeCreateModal();
      }
    },
  });

  useEffect(() => {
    if (!closeTarget) resetCloseError();
  }, [closeTarget, resetCloseError]);

  const [worktreeFamilyId, setWorktreeFamilyIdState] = useState("all");
  const setWorktreeFamilyId = useCallback((next: string) => {
    setWorktreeFamilyIdState(next);
  }, []);

  const list = useTasksHomeList({
    dataEnabled,
    bootstrapSettled,
    worktreeFamilyId,
  });

  useEffect(() => {
    list.resetTaskListPage();
    // Reset page when the family filter changes; list identity is stable enough.
    // eslint-disable-next-line react-hooks/exhaustive-deps -- only family id should reset page
  }, [worktreeFamilyId]);

  const edit = useTaskEditFlow({
    editingTaskId,
    closeCreateModal,
    beginEditSession,
    newTitle,
    newPrompt,
    newPriority,
    newProjectID,
    newTagsCsv,
    newMilestone,
    newTaskCursorModel,
    newSchedule,
    composeStatus,
    createModalOpen,
  });

  const saving =
    createFlow.createPending ||
    createFlow.templateSavePending ||
    edit.patchPending ||
    closePending ||
    reopenPending;

  const error = useMemo(() => {
    if (list.listError) return list.listError;
    if (createFlowError) return createFlowError;
    if (edit.patchError) return edit.patchError;
    if (closeError) return closeError;
    if (reopenError) return reopenError;
    return edit.editTitleRequiredError;
  }, [
    list.listError,
    createFlowError,
    edit.patchError,
    closeError,
    reopenError,
    edit.editTitleRequiredError,
  ]);

  const submitComposeModal = useCallback(
    (e: FormEvent) => {
      if (editingTaskId) {
        edit.submitEdit(e);
        return;
      }
      if (createFlow.composeTarget === "template") {
        void createFlow.submitTemplate(e);
        return;
      }
      void createFlow.submitCreate(e);
    },
    [
      editingTaskId,
      edit.submitEdit,
      createFlow.composeTarget,
      createFlow.submitTemplate,
      createFlow.submitCreate,
    ],
  );

  return {
    ...createFlow,
    closeCreateModal,
    editingTaskId,
    composeStatus,
    newTitle,
    newPrompt,
    newPriority,
    newProjectID,
    newTagsCsv,
    newMilestone,
    newTaskCursorModel,
    newSchedule,
    createModalOpen,
    beginEditSession,
    tasks: list.tasks,
    rootTasksOnPage: list.rootTasksOnPage,
    loading: list.loading,
    listRefreshing: list.listRefreshing,
    saving,
    patchPending: edit.patchPending,
    patchError: edit.patchError,
    closePending,
    closeSuccess,
    closeVariables,
    reopen,
    reopenPending,
    reopenError,
    resetReopenError,
    error,
    sseLive,
    taskStats: list.taskStats,
    taskStatsLoading: list.taskStatsLoading,
    changeModelTask: edit.changeModelTask,
    changeModelDraft: edit.changeModelDraft,
    setChangeModelDraft: edit.setChangeModelDraft,
    openChangeModel: edit.openChangeModel,
    closeChangeModel: edit.closeChangeModel,
    submitChangeModel: edit.submitChangeModel,
    openEdit: edit.openEdit,
    closeEdit: edit.closeEdit,
    submitEdit: edit.submitEdit,
    submitComposeModal,
    editFormError: edit.editTitleRequiredError,
    closeTarget,
    requestClose,
    cancelClose,
    confirmClose,
    closeError,
    taskListPage: list.taskListPage,
    setTaskListPage: list.setTaskListPage,
    resetTaskListPage: list.resetTaskListPage,
    taskListPageSize: list.taskListPageSize,
    hasNextTaskPage: list.hasNextTaskPage,
    hasPrevTaskPage: list.hasPrevTaskPage,
    worktreeFamilyId,
    setWorktreeFamilyId,
    /** True when home list/board queries may run (route + bootstrap). */
    homeDataReady,
  };
}

// Re-export for callers that type against Task from this module historically.
export type { Task };
