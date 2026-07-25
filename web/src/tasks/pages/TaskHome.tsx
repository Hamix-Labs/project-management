import { useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import { Button } from "@/components/ui";
import { runViewTransition } from "@/lib/runViewTransition";
import { TaskListSection } from "../components/task-list";
import { TaskBoardSection } from "../components/task-board/TaskBoardSection";
import { TaskHomeViewToggle } from "../components/task-board/TaskHomeViewToggle";
import { TaskTimelineSection } from "../components/task-timeline/TaskTimelineSection";
import { useTasksAppList, useTasksAppModals } from "../app/TasksAppProvider";
import { useTasksBoard } from "../hooks/useTasksBoard";
import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import { useProjects } from "@/hooks/useProjects";
import {
  applyTaskHomeView,
  parseTaskHomeView,
  type TaskHomeView,
} from "./taskHomeView";

export function TaskHome() {
  useDocumentTitle(undefined);
  const list = useTasksAppList();
  const modals = useTasksAppModals();
  const [searchParams, setSearchParams] = useSearchParams();
  const view = parseTaskHomeView(searchParams.get("view"));
  /** CSS enter fallback when View Transitions API is unavailable. */
  const [viewSwapEnter, setViewSwapEnter] = useState(false);
  const projectsUiEnabled = !isUiFeatureOmitted("projects");
  const projects = useProjects({
    includeArchived: false,
    limit: 100,
    enabled: projectsUiEnabled,
  });
  const { openCreateModal, createModalOpen } = modals;

  const board = useTasksBoard({
    view,
    dataEnabled: list.homeDataReady,
    bootstrapSettled: true,
  });

  const createIntent = searchParams.get("create");
  const projectIntent = projectsUiEnabled
    ? (searchParams.get("project")?.trim() ?? "")
    : "";

  useEffect(() => {
    if (createIntent !== "1" || !projectIntent) return;
    openCreateModal({ projectID: projectIntent });
    setSearchParams({}, { replace: true });
  }, [openCreateModal, createIntent, projectIntent, setSearchParams]);

  const onViewChange = useCallback(
    (next: TaskHomeView) => {
      const apply = () => {
        setSearchParams(applyTaskHomeView(searchParams, next), {
          replace: true,
        });
      };
      if (runViewTransition(apply)) return;
      setViewSwapEnter(true);
      apply();
    },
    [searchParams, setSearchParams],
  );

  /** Row-level busy state for the list only; excludes create so modal typing does not re-render the table. */
  const listSaving = list.patchPending || list.deletePending;

  const projectFilterOptions = projectsUiEnabled
    ? (projects.data?.projects ?? [])
    : [];

  const listSectionProps = useMemo(
    () => ({
      tasks: list.tasks,
      rootTasksOnPage: list.rootTasksOnPage,
      loading: list.loading,
      refreshing: list.listRefreshing,
      saving: listSaving,
      hideBackgroundRefreshHint: list.sseLive,
      listPage: list.taskListPage,
      listPageSize: list.taskListPageSize,
      projectFilterOptions,
      showProjectColumn: projectsUiEnabled,
      onListPageChange: list.setTaskListPage,
      onListFiltersChange: list.resetTaskListPage,
      hasNextPage: list.hasNextTaskPage,
      hasPrevPage: list.hasPrevTaskPage,
      onEdit: modals.openEdit,
      onRequestDelete: modals.requestDelete,
      taskStats: list.taskStats ?? null,
    }),
    [
      list.tasks,
      list.rootTasksOnPage,
      list.loading,
      list.listRefreshing,
      listSaving,
      list.sseLive,
      list.taskListPage,
      list.taskListPageSize,
      projectFilterOptions,
      projectsUiEnabled,
      list.setTaskListPage,
      list.resetTaskListPage,
      list.hasNextTaskPage,
      list.hasPrevTaskPage,
      modals.openEdit,
      modals.requestDelete,
      list.taskStats,
    ],
  );

  const listActions = useMemo(
    () => (
      <>
        <TaskHomeViewToggle value={view} onChange={onViewChange} />
        <Button
          variant="secondary"
          className="task-home-new-template-btn"
          onClick={() => modals.openTemplateCreateModal()}
          disabled={createModalOpen}
        >
          <svg
            className="task-home-new-template-btn__icon"
            width="16"
            height="16"
            viewBox="0 0 16 16"
            fill="none"
            aria-hidden="true"
          >
            <path
              d="M3 2.5h7l3 3V13.5H3V2.5Z"
              stroke="currentColor"
              strokeWidth="1.2"
              strokeLinejoin="round"
            />
            <path
              d="M10 2.5V5.5H13"
              stroke="currentColor"
              strokeWidth="1.2"
              strokeLinejoin="round"
            />
            <path
              d="M8 8.5V11.5"
              stroke="currentColor"
              strokeWidth="1.2"
              strokeLinecap="round"
            />
            <path
              d="M6.5 10H9.5"
              stroke="currentColor"
              strokeWidth="1.2"
              strokeLinecap="round"
            />
          </svg>
          New template
        </Button>
        <Button
          variant="primary"
          className="task-home-new-task-btn"
          onClick={() => openCreateModal()}
          disabled={createModalOpen}
        >
          + New task
        </Button>
      </>
    ),
    [
      view,
      onViewChange,
      openCreateModal,
      modals.openTemplateCreateModal,
      createModalOpen,
    ],
  );

  const emptyAction = useMemo(
    () => ({
      label: "New task",
      onClick: () => openCreateModal(),
      disabled: createModalOpen,
    }),
    [openCreateModal, createModalOpen],
  );

  return (
    <div className="task-detail-content--enter">
      <div
        key={view}
        className={
          viewSwapEnter
            ? "task-home-view-swap task-home-view-swap--enter"
            : "task-home-view-swap"
        }
      >
        {view === "board" ? (
          <TaskBoardSection
            tasks={board.tasks}
            loading={board.loading}
            refreshing={board.refreshing}
            hideBackgroundRefreshHint={list.sseLive}
            error={board.error}
            truncated={board.truncated}
            onRetry={() => void board.refetch()}
            projectFilterOptions={projectFilterOptions}
            showProjectColumn={projectsUiEnabled}
            actions={listActions}
            emptyListAction={emptyAction}
          />
        ) : view === "timeline" ? (
          <TaskTimelineSection actions={listActions} />
        ) : (
          <TaskListSection
            {...listSectionProps}
            actions={listActions}
          />
        )}
      </div>
    </div>
  );
}
