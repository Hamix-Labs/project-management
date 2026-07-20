import type { Task } from "@/types";
import type { TaskWithDepth } from "../../../task-tree";
import type { DeleteTargetInput } from "../../../hooks/useTaskDeleteFlow";
import type { EmptyStateAction } from "@/shared/EmptyState";
import { TaskListDataTable } from "../table/TaskListDataTable";
import { TaskListTableSkeleton } from "../table/TaskListTableSkeleton";
import { TaskPager } from "../pager/TaskPager";
import { taskListPagerSummary } from "../pager/taskListPagerSummary";
import type { useTaskListSectionFilters } from "./useTaskListSectionFilters";
import type { useTaskListSectionBulkActions } from "./useTaskListSectionBulkActions";

type Filters = ReturnType<typeof useTaskListSectionFilters>;
type Bulk = ReturnType<typeof useTaskListSectionBulkActions>;

type Props = {
  caption: string;
  loading: boolean;
  showLoadingLine: boolean;
  refreshing: boolean;
  hideBackgroundRefreshHint: boolean;
  saving: boolean;
  tasks: TaskWithDepth[];
  rootTasksOnPage: number;
  listPage: number;
  listPageSize: number;
  hasNextPage: boolean;
  hasPrevPage: boolean;
  showTaskPager: boolean;
  showProjectColumn: boolean;
  emptyListAction?: EmptyStateAction;
  onEdit: (t: Task) => void;
  onRequestDelete: (t: DeleteTargetInput) => void;
  onListPageChange: (page: number) => void;
  filters: Filters;
  bulk: Bulk;
};

export function TaskListTableRegion({
  caption,
  loading,
  showLoadingLine,
  refreshing,
  hideBackgroundRefreshHint,
  saving,
  tasks,
  rootTasksOnPage,
  listPage,
  listPageSize,
  hasNextPage,
  hasPrevPage,
  showTaskPager,
  showProjectColumn,
  emptyListAction,
  onEdit,
  onRequestDelete,
  onListPageChange,
  filters,
  bulk,
}: Props) {
  return (
    <>
      {refreshing && !loading && !hideBackgroundRefreshHint ? (
        <p className="sync-hint task-list-phase-msg" aria-live="polite" role="status">
          Syncing with server…
        </p>
      ) : null}
      {loading && showLoadingLine ? (
        <TaskListTableSkeleton caption={caption} />
      ) : null}
      {!loading ? (
        <div className="task-list-content task-list-content--enter">
          <TaskListDataTable
            caption={caption}
            refreshing={refreshing}
            tasks={tasks}
            filteredTasks={filters.filteredTasks}
            saving={saving}
            emptyListAction={emptyListAction}
            onEdit={onEdit}
            onRequestDelete={onRequestDelete}
            projectNameById={filters.projectNameById}
            showProjectColumn={showProjectColumn}
            sortKey={filters.sortKey}
            sortDir={filters.sortDir}
            onSortChange={filters.handleSortChange}
            selection={{
              isSelected: bulk.selection.isSelected,
              onRowToggle: bulk.selection.toggle,
              allVisibleSelected: bulk.selection.allVisibleSelected,
              someVisibleSelected: bulk.selection.someVisibleSelected,
              onToggleAllVisible: bulk.selection.toggleAllVisible,
            }}
          />
          {bulk.bulkErrorBanner ? (
            <p
              className="err task-list-bulk-error"
              role="alert"
              data-testid="task-list-bulk-error"
            >
              {bulk.bulkErrorBanner}
            </p>
          ) : null}
          {showTaskPager ? (
            <TaskPager
              navLabel="Task list pages"
              summary={taskListPagerSummary({
                tasksLength: tasks.length,
                listPage,
                listPageSize,
                rootTasksOnPage,
                hasNextPage,
              })}
              onPrev={() => onListPageChange(listPage - 1)}
              onNext={() => onListPageChange(listPage + 1)}
              disablePrev={!hasPrevPage}
              disableNext={!hasNextPage}
            />
          ) : null}
        </div>
      ) : null}
    </>
  );
}
