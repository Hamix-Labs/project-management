import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import { Button } from "@/components/ui";
import { runViewTransition } from "@/lib/runViewTransition";
import type { Task } from "@/types";
import { DraftResumeModal } from "../components/draft-resume";
import { TaskListSection } from "../components/task-list";
import { TaskBoardSection } from "../components/task-board/TaskBoardSection";
import { TaskHomeViewToggle } from "../components/task-board/TaskHomeViewToggle";
import { useTasksAppList, useTasksAppModals } from "../app/TasksAppProvider";
import { useTasksBoard } from "../hooks/useTasksBoard";
import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import { useProjects } from "@/hooks/useProjects";
import { taskEditPath, tasksNewPath, templatesNewPath } from "../composeRoutes";
import {
  applyTaskHomeView,
  parseTaskHomeView,
  type TaskHomeView,
} from "./taskHomeView";
import { useTaskHomeNewTask } from "./useTaskHomeNewTask";

export function TaskHome() {
  useDocumentTitle(undefined);
  const list = useTasksAppList();
  const modals = useTasksAppModals();
  const navigate = useNavigate();
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
  const newTask = useTaskHomeNewTask();

  const board = useTasksBoard({
    view,
    dataEnabled: list.homeDataReady,
    bootstrapSettled: true,
    worktreeFamilyId: list.worktreeFamilyId,
  });

  const createIntent = searchParams.get("create");
  const projectIntent = projectsUiEnabled
    ? (searchParams.get("project")?.trim() ?? "")
    : "";

  useEffect(() => {
    if (createIntent !== "1" || !projectIntent) return;
    navigate(tasksNewPath({ project: projectIntent }), { replace: true });
  }, [navigate, createIntent, projectIntent]);

  const openTemplateCreate = useCallback(() => {
    navigate(templatesNewPath());
  }, [navigate]);

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
  const listSaving = list.patchPending || list.closePending;

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
      worktreeFamilyFilter: list.worktreeFamilyId,
      onWorktreeFamilyFilterChange: list.setWorktreeFamilyId,
      onEdit: (task: Task) => navigate(taskEditPath(task.id)),
      onRequestClose: modals.requestClose,
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
      list.worktreeFamilyId,
      list.setWorktreeFamilyId,
      navigate,
      modals.requestClose,
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
          onClick={() => openTemplateCreate()}
          disabled={newTask.createBlocked}
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
          onClick={() => newTask.openCreate()}
          disabled={newTask.createBlocked}
          loading={newTask.awaitingDrafts}
        >
          + New task
        </Button>
      </>
    ),
    [
      view,
      onViewChange,
      newTask.openCreate,
      newTask.createBlocked,
      newTask.awaitingDrafts,
      openTemplateCreate,
    ],
  );

  const emptyAction = useMemo(
    () => ({
      label: "New task",
      onClick: () => newTask.openCreate(),
      disabled: newTask.createBlocked,
    }),
    [newTask.openCreate, newTask.createBlocked],
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
            worktreeFamilyFilter={list.worktreeFamilyId}
            onWorktreeFamilyFilterChange={list.setWorktreeFamilyId}
            actions={listActions}
            emptyListAction={emptyAction}
          />
        ) : (
          <TaskListSection
            {...listSectionProps}
            actions={listActions}
          />
        )}
      </div>
      {newTask.resumeOpen ? (
        <DraftResumeModal
          drafts={newTask.drafts}
          onClose={newTask.closeResume}
          onStartFresh={newTask.startFresh}
          onResume={newTask.resumeDraft}
          loadError={newTask.draftListError}
          onRetryLoad={() => {
            void newTask.retryDraftList();
          }}
          resumePending={newTask.resumePending}
          resumeError={newTask.resumeError}
        />
      ) : null}
    </div>
  );
}
