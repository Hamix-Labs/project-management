import {
  createContext,
  useContext,
  useMemo,
  type ReactNode,
} from "react";
import type { useTasksApp } from "../hooks/useTasksApp";

export type TasksAppContextValue = ReturnType<typeof useTasksApp>;

const TasksAppContext = createContext<TasksAppContextValue | null>(null);

export function TasksAppProvider({
  value,
  children,
}: {
  value: TasksAppContextValue;
  children: ReactNode;
}) {
  return (
    <TasksAppContext.Provider value={value}>{children}</TasksAppContext.Provider>
  );
}

export function useTasksAppContext(): TasksAppContextValue {
  const ctx = useContext(TasksAppContext);
  if (!ctx) {
    throw new Error("useTasksAppContext must be used within TasksAppProvider");
  }
  return ctx;
}

export function useTasksAppList() {
  const app = useTasksAppContext();
  return useMemo(
    () => ({
      tasks: app.tasks,
      rootTasksOnPage: app.rootTasksOnPage,
      loading: app.loading,
      listRefreshing: app.listRefreshing,
      patchPending: app.patchPending,
      closePending: app.closePending,
      sseLive: app.sseLive,
      taskListPage: app.taskListPage,
      setTaskListPage: app.setTaskListPage,
      resetTaskListPage: app.resetTaskListPage,
      taskListPageSize: app.taskListPageSize,
      hasNextTaskPage: app.hasNextTaskPage,
      hasPrevTaskPage: app.hasPrevTaskPage,
      openEdit: app.openEdit,
      requestClose: app.requestClose,
      taskStats: app.taskStats,
      taskStatsLoading: app.taskStatsLoading,
      homeDataReady: app.homeDataReady,
    }),
    [
      app.tasks,
      app.rootTasksOnPage,
      app.loading,
      app.listRefreshing,
      app.patchPending,
      app.closePending,
      app.sseLive,
      app.taskListPage,
      app.setTaskListPage,
      app.resetTaskListPage,
      app.taskListPageSize,
      app.hasNextTaskPage,
      app.hasPrevTaskPage,
      app.openEdit,
      app.requestClose,
      app.taskStats,
      app.taskStatsLoading,
      app.homeDataReady,
    ],
  );
}

export function useTasksAppModals() {
  const app = useTasksAppContext();
  return useMemo(
    () => ({
      closeTarget: app.closeTarget,
      requestClose: app.requestClose,
      cancelClose: app.cancelClose,
      confirmClose: app.confirmClose,
      closePending: app.closePending,
      closeError: app.closeError,
      closeSuccess: app.closeSuccess,
      closeVariables: app.closeVariables,
      reopen: app.reopen,
      reopenPending: app.reopenPending,
      reopenError: app.reopenError,
      changeModelTask: app.changeModelTask,
      changeModelDraft: app.changeModelDraft,
      setChangeModelDraft: app.setChangeModelDraft,
      openChangeModel: app.openChangeModel,
      closeChangeModel: app.closeChangeModel,
      submitChangeModel: app.submitChangeModel,
      patchPending: app.patchPending,
      patchError: app.patchError,
      openCreateModal: app.openCreateModal,
      openTemplateCreateModal: app.openTemplateCreateModal,
      createModalOpen: app.createModalOpen,
      openEdit: app.openEdit,
      closeEdit: app.closeEdit,
      closeCreateModal: app.closeCreateModal,
    }),
    [
      app.closeTarget,
      app.requestClose,
      app.cancelClose,
      app.confirmClose,
      app.closePending,
      app.closeError,
      app.closeSuccess,
      app.closeVariables,
      app.reopen,
      app.reopenPending,
      app.reopenError,
      app.changeModelTask,
      app.changeModelDraft,
      app.setChangeModelDraft,
      app.openChangeModel,
      app.closeChangeModel,
      app.submitChangeModel,
      app.patchPending,
      app.patchError,
      app.openCreateModal,
      app.openTemplateCreateModal,
      app.createModalOpen,
      app.openEdit,
      app.closeEdit,
      app.closeCreateModal,
    ],
  );
}

export function useTasksAppMeta() {
  const app = useTasksAppContext();
  return useMemo(
    () => ({
      error: app.error,
      saving: app.saving,
      sseLive: app.sseLive,
    }),
    [app.error, app.saving, app.sseLive],
  );
}
