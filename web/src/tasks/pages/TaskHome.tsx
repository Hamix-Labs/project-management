import { useEffect, useMemo } from "react";
import { useSearchParams } from "react-router-dom";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import { Button } from "@/components/ui";
import { TaskListSection } from "../components/task-list";
import { useTasksAppList, useTasksAppModals } from "../app/TasksAppProvider";
import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import { useProjects } from "@/projects";

export function TaskHome() {
  useDocumentTitle(undefined);
  const list = useTasksAppList();
  const modals = useTasksAppModals();
  const [searchParams, setSearchParams] = useSearchParams();
  const projectsUiEnabled = !isUiFeatureOmitted("projects");
  const projects = useProjects({
    includeArchived: false,
    limit: 100,
    enabled: projectsUiEnabled,
  });
  const { openCreateModal, createModalOpen } = modals;

  const createIntent = searchParams.get("create");
  const projectIntent = projectsUiEnabled
    ? (searchParams.get("project")?.trim() ?? "")
    : "";

  useEffect(() => {
    if (createIntent !== "1" || !projectIntent) return;
    openCreateModal({ projectID: projectIntent });
    setSearchParams({}, { replace: true });
  }, [openCreateModal, createIntent, projectIntent, setSearchParams]);

  /** Row-level busy state for the list only; excludes create so modal typing does not re-render the table. */
  const listSaving = list.patchPending || list.deletePending;

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
      projectFilterOptions: projectsUiEnabled
        ? (projects.data?.projects ?? [])
        : [],
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
      projectsUiEnabled,
      projects.data?.projects,
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
    [openCreateModal, modals.openTemplateCreateModal, createModalOpen],
  );

  return (
    <div className="task-detail-content--enter">
      <TaskListSection {...listSectionProps} actions={listActions} />
    </div>
  );
}
