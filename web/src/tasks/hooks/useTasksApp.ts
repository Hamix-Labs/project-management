import { useCallback, useEffect, useMemo, useRef, type FormEvent } from "react";
import type { Task } from "@/types";
import { useTaskDeleteFlow } from "./useTaskDeleteFlow";
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
 * Thin facade: create flow + home list + edit/delete composers.
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
    newProjectContextItemIDs,
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
    deleteTarget,
    requestDelete,
    cancelDelete,
    confirmDelete,
    deletePending,
    deleteError,
    deleteSuccess,
    deleteVariables,
    resetError: resetDeleteError,
  } = useTaskDeleteFlow({
    onDeleted: (deletedId) => {
      if (editingTaskIdRef.current === deletedId) {
        closeCreateModal();
      }
    },
  });

  useEffect(() => {
    if (!deleteTarget) resetDeleteError();
  }, [deleteTarget, resetDeleteError]);

  const list = useTasksHomeList({ dataEnabled, bootstrapSettled });

  const edit = useTaskEditFlow({
    editingTaskId,
    closeCreateModal,
    beginEditSession,
    newTitle,
    newPrompt,
    newPriority,
    newProjectID,
    newProjectContextItemIDs,
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
    deletePending;

  const error = useMemo(() => {
    if (list.listError) return list.listError;
    if (createFlowError) return createFlowError;
    if (edit.patchError) return edit.patchError;
    if (deleteError) return deleteError;
    return edit.editTitleRequiredError;
  }, [
    list.listError,
    createFlowError,
    edit.patchError,
    deleteError,
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
    newProjectContextItemIDs,
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
    deletePending,
    deleteSuccess,
    deleteVariables,
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
    deleteTarget,
    requestDelete,
    cancelDelete,
    confirmDelete,
    deleteError,
    taskListPage: list.taskListPage,
    setTaskListPage: list.setTaskListPage,
    resetTaskListPage: list.resetTaskListPage,
    taskListPageSize: list.taskListPageSize,
    hasNextTaskPage: list.hasNextTaskPage,
    hasPrevTaskPage: list.hasPrevTaskPage,
    /** True when home list/board queries may run (route + bootstrap). */
    homeDataReady,
  };
}

// Re-export for callers that type against Task from this module historically.
export type { Task };
