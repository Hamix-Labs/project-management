import { useMemo, type ReactNode } from "react";
import { useDelayedTrue } from "@/lib/useDelayedTrue";
import { EmptyState, EmptyStateFilterGlyph } from "@/shared/EmptyState";
import { Button } from "@/components/ui";
import { useTaskDetailPrefetcher } from "@/app/hooks/usePrefetchOnIntent";
import { TaskListFilters } from "../task-list/filters/TaskListFilters";
import { TaskListSectionHeading } from "../task-list/section/TaskListSectionHeading";
import type { TaskWithDepth } from "../../task-tree";
import { BOARD_COLUMNS } from "./boardColumns";
import { BOARD_ACTIVE_CAP } from "./boardConstants";
import { groupTasksByBoardColumn } from "./groupTasksByBoardColumn";
import { TaskBoardColumn } from "./TaskBoardColumn";
import { TaskBoardSkeleton } from "./TaskBoardSkeleton";
import { BoardActivePill } from "./BoardActivePill";
import { useTaskBoardFilters } from "./useTaskBoardFilters";

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
  const statusDelayMs = smoothTransitions ? LOADING_STATUS_DELAY_MS : 0;
  const showSkeleton = useDelayedTrue(loading, statusDelayMs);
  const filters = useTaskBoardFilters({
    tasks,
    projectFilterOptions,
    showProjectColumn,
    smoothTransitions,
  });
  const prefetchTaskDetail = useTaskDetailPrefetcher();

  const groups = useMemo(
    () => groupTasksByBoardColumn(filters.filteredTasks),
    [filters.filteredTasks],
  );

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
          summary={
            loading ? undefined : (
              <BoardActivePill count={filters.filteredTasks.length} />
            )
          }
          description="Track engineering work across every stage."
        />
        {!loading ? (
          <TaskListFilters
            priorityFilter={filters.priorityFilter}
            onPriorityFilterChange={filters.onPriorityChange}
            projectFilter={filters.projectFilter}
            projectOptions={showProjectColumn ? projectFilterOptions : []}
            onProjectFilterChange={
              showProjectColumn ? filters.setProjectFilter : undefined
            }
            tagFilter={filters.tagFilter}
            tagOptions={filters.tagFilterOptions}
            onTagFilterChange={
              filters.tagsUiEnabled ? filters.setTagFilter : undefined
            }
            titleSearch={filters.titleSearch}
            onTitleSearchChange={filters.setTitleSearch}
            searchInputRef={filters.searchInputRef}
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

      {!showSkeleton && !error && filters.filteredTasks.length === 0 ? (
        <EmptyState
          title={
            filters.hasClientFilters ? "No matching tasks" : "No active tasks"
          }
          description={
            filters.hasClientFilters
              ? "Try clearing filters to see active work."
              : "Tasks in progress will show up here. Done tasks are hidden from the board."
          }
          icon={
            filters.hasClientFilters ? <EmptyStateFilterGlyph /> : undefined
          }
          action={!filters.hasClientFilters ? emptyListAction : undefined}
          className="empty-state--task-list-fresh"
        />
      ) : null}

      {!showSkeleton && !error && filters.filteredTasks.length > 0 ? (
        <div className="task-board-track">
          {BOARD_COLUMNS.map((column) => (
            <TaskBoardColumn
              key={column.id}
              column={column}
              tasks={groups[column.id]}
              projectNameById={filters.projectNameById}
              showProject={showProjectColumn}
              showTags={filters.tagsUiEnabled}
              prefetchTaskDetail={prefetchTaskDetail}
            />
          ))}
        </div>
      ) : null}
    </section>
  );
}
