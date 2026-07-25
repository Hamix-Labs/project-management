import { useMemo, type ReactNode } from "react";
import { useDelayedTrue } from "@/lib/useDelayedTrue";
import { EmptyState, EmptyStateFilterGlyph } from "@/shared/EmptyState";
import { Button } from "@/components/ui";
import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import { TaskListFilters } from "../task-list/filters/TaskListFilters";
import {
  filterTasksByTag,
  filterTasksForListView,
  uniqueSortedTagsFromTasks,
  type TaskListClientPriorityFilter,
} from "../task-list/filters/taskListClientFilter";
import { TaskListSectionHeading } from "../task-list/section/TaskListSectionHeading";
import { useTaskListSearchShortcut } from "../task-list/hooks/useTaskListSearchShortcut";
import type { TaskWithDepth } from "../../task-tree";
import { BOARD_COLUMNS } from "./boardColumns";
import { BOARD_ACTIVE_CAP } from "./boardConstants";
import { groupTasksByBoardColumn } from "./groupTasksByBoardColumn";
import { TaskBoardColumn } from "./TaskBoardColumn";
import { TaskBoardSkeleton } from "./TaskBoardSkeleton";
import { useCallback, useRef, useState } from "react";

type Props = {
  tasks: TaskWithDepth[];
  loading: boolean;
  refreshing: boolean;
  hideBackgroundRefreshHint?: boolean;
  error: string | null;
  truncated: boolean;
  onRetry: () => void;
  projectFilterOptions?: Array<{ id: string; name: string }>;
  showProjectColumn?: boolean;
  actions?: ReactNode;
  emptyListAction?: { label: string; onClick: () => void; disabled?: boolean };
  smoothTransitions?: boolean;
};

const LOADING_STATUS_DELAY_MS = 220;

export function TaskBoardSection({
  tasks,
  loading,
  refreshing,
  hideBackgroundRefreshHint = false,
  error,
  truncated,
  onRetry,
  projectFilterOptions = [],
  showProjectColumn = true,
  actions,
  emptyListAction,
  smoothTransitions = true,
}: Props) {
  const tagsUiEnabled = !isUiFeatureOmitted("taskTags");
  const statusDelayMs = smoothTransitions ? LOADING_STATUS_DELAY_MS : 0;
  const showSkeleton = useDelayedTrue(loading, statusDelayMs);

  const [priorityFilter, setPriorityFilter] =
    useState<TaskListClientPriorityFilter>("all");
  const [projectFilter, setProjectFilter] = useState("all");
  const [tagFilter, setTagFilter] = useState("all");
  const [titleSearch, setTitleSearch] = useState("");
  const searchInputRef = useRef<HTMLInputElement>(null);
  useTaskListSearchShortcut(searchInputRef, smoothTransitions);

  const projectNameById = useMemo(() => {
    const m: Record<string, string> = {};
    for (const p of projectFilterOptions) {
      m[p.id] = p.name;
    }
    return m;
  }, [projectFilterOptions]);

  const tagFilterOptions = useMemo(() => {
    if (!tagsUiEnabled) return [];
    return uniqueSortedTagsFromTasks(tasks);
  }, [tagsUiEnabled, tasks]);

  const filteredTasks = useMemo(() => {
    const base = filterTasksForListView(
      tasks,
      "all",
      priorityFilter,
      titleSearch,
    );
    const scoped =
      projectFilter === "all"
        ? base
        : projectFilter === "none"
          ? base.filter((task) => !task.project_id)
          : base.filter((task) => task.project_id === projectFilter);
    return filterTasksByTag(scoped, tagsUiEnabled ? tagFilter : "all");
  }, [
    tasks,
    priorityFilter,
    titleSearch,
    projectFilter,
    tagFilter,
    tagsUiEnabled,
  ]);

  const groups = useMemo(
    () => groupTasksByBoardColumn(filteredTasks),
    [filteredTasks],
  );

  const hasClientFilters =
    priorityFilter !== "all" ||
    projectFilter !== "all" ||
    tagFilter !== "all" ||
    titleSearch.trim() !== "";

  const onPriorityChange = useCallback((v: string) => {
    setPriorityFilter(v as TaskListClientPriorityFilter);
  }, []);

  return (
    <section
      className="panel task-list-section-panel task-list-section--redesign task-board-section"
      aria-labelledby="task-board-heading"
      id="task-board-panel"
      role="tabpanel"
    >
      <div className="task-list-toolbar">
        <TaskListSectionHeading
          title="Board"
          titleId="task-board-heading"
          actions={actions}
          summary={loading ? undefined : `${filteredTasks.length} active`}
        />
        {!loading ? (
          <TaskListFilters
            priorityFilter={priorityFilter}
            onPriorityFilterChange={onPriorityChange}
            projectFilter={projectFilter}
            projectOptions={showProjectColumn ? projectFilterOptions : []}
            onProjectFilterChange={
              showProjectColumn ? setProjectFilter : undefined
            }
            tagFilter={tagFilter}
            tagOptions={tagFilterOptions}
            onTagFilterChange={tagsUiEnabled ? setTagFilter : undefined}
            titleSearch={titleSearch}
            onTitleSearchChange={setTitleSearch}
            searchInputRef={searchInputRef}
          />
        ) : null}
      </div>

      {truncated && !loading && !error ? (
        <p className="task-board-truncation" role="status">
          Showing the first {BOARD_ACTIVE_CAP} active tasks. Older active
          tasks may be missing until server-side filtering lands.
        </p>
      ) : null}

      {showSkeleton ? <TaskBoardSkeleton /> : null}

      {!showSkeleton && error ? (
        <div className="task-board-error" role="alert">
          <p className="err">{error}</p>
          <Button
            variant="secondary"
            type="button"
            onClick={() => void onRetry()}
          >
            Retry
          </Button>
        </div>
      ) : null}

      {!showSkeleton &&
      !error &&
      refreshing &&
      !hideBackgroundRefreshHint ? (
        <p className="sync-hint task-list-phase-msg">Syncing with server…</p>
      ) : null}

      {!showSkeleton && !error && filteredTasks.length === 0 ? (
        <EmptyState
          title={hasClientFilters ? "No matching tasks" : "No active tasks"}
          description={
            hasClientFilters
              ? "Try clearing filters to see active work."
              : "Tasks in progress will show up here. Done tasks are hidden from the board."
          }
          icon={hasClientFilters ? <EmptyStateFilterGlyph /> : undefined}
          action={!hasClientFilters ? emptyListAction : undefined}
          className="empty-state--task-list-fresh"
        />
      ) : null}

      {!showSkeleton && !error && filteredTasks.length > 0 ? (
        <div className="task-board-track">
          {BOARD_COLUMNS.map((column) => (
            <TaskBoardColumn
              key={column.id}
              column={column}
              tasks={groups[column.id]}
              projectNameById={projectNameById}
              showProject={showProjectColumn}
            />
          ))}
        </div>
      ) : null}
    </section>
  );
}
