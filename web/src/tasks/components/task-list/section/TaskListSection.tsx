import {
  memo,
  useEffect,
  useMemo,
  useRef,
  type ReactNode,
} from "react";
import { useDelayedTrue } from "@/lib/useDelayedTrue";
import { TaskListDataTable } from "../table/TaskListDataTable";
import { TaskListFilters } from "../filters/TaskListFilters";
import { TaskListStatusTabs } from "../filters/TaskListStatusTabs";
import { TaskListSectionHeading } from "./TaskListSectionHeading";
import { formatTaskListHeadingSummary } from "./taskListHeadingSummary";
import { TaskPager } from "../pager/TaskPager";
import type { Task, TaskStatsResponse } from "@/types";
import type { TaskWithDepth } from "../../../task-tree";
import type { DeleteTargetInput } from "../../../hooks/useTaskDeleteFlow";
import type { EmptyStateAction } from "@/shared/EmptyState";
import { taskListPagerSummary } from "../pager/taskListPagerSummary";
import { TaskListTableSkeleton } from "../table/TaskListTableSkeleton";
import { useAppTimezone } from "@/shared/time/appTimezone";
import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import {
  TaskBulkDeleteConfirmModal,
  TaskBulkRescheduleModal,
  TaskListBulkActionBar,
} from "../bulk";
import {
  useTaskListSectionFilters,
} from "./useTaskListSectionFilters";
import { useTaskListSectionBulkActions } from "./useTaskListSectionBulkActions";

type Props = {
  tasks: TaskWithDepth[];
  /** Tasks returned for this list page (for pager copy). */
  rootTasksOnPage: number;
  loading: boolean;
  /** Background refetch in progress (list still visible). */
  refreshing: boolean;
  /** A create/update/delete request is in flight. */
  saving: boolean;
  /**
   * When true, hide the background “Syncing with server…” line (e.g. live SSE
   * already drives refetches; avoids duplicate status with the header).
   */
  hideBackgroundRefreshHint?: boolean;
  /** Zero-based server list page (see `GET /tasks` offset). */
  listPage: number;
  listPageSize: number;
  projectFilterOptions?: Array<{ id: string; name: string }>;
  /** When false, hides the project filter and table column (launch omission). */
  showProjectColumn?: boolean;
  onListPageChange: (page: number) => void;
  /** Reset to first server page when filters change. */
  onListFiltersChange: () => void;
  hasNextPage: boolean;
  hasPrevPage: boolean;
  /**
   * When true (default), the loading line waits briefly before appearing. Set false in tests.
   * List “syncing” is smoothed in `useTasksApp` (hysteresis on refetch).
   */
  smoothTransitions?: boolean;
  onEdit: (t: Task) => void;
  /**
   * Opens in-app delete confirmation (do not call `window.confirm` from the
   * table).
   */
  onRequestDelete: (t: DeleteTargetInput) => void;
  /** Primary action when the server returned no tasks (e.g. open create modal). */
  emptyListAction?: EmptyStateAction;
  /** Optional toolbar on the title row (e.g. home “New task”). */
  actions?: ReactNode;
  /**
   * Optional `GET /tasks/stats` projection. When supplied AND total > 0 the
   * section renders a quiet inline stats strip below the heading (counts
   * for ready / critical / scheduled / etc). Pass `null` while the stats
   * query is loading or has errored — the strip self-hides on falsy data.
   */
  taskStats?: TaskStatsResponse | null;
};

const LOADING_STATUS_DELAY_MS = 220;

const TASK_LIST_TABLE_CAPTION =
  "All tasks: title with context line, status, priority, created time, project, and row actions.";

export const TaskListSection = memo(function TaskListSection({
  tasks,
  rootTasksOnPage,
  loading,
  refreshing,
  saving,
  hideBackgroundRefreshHint = false,
  listPage,
  listPageSize,
  projectFilterOptions = [],
  showProjectColumn = true,
  onListPageChange,
  onListFiltersChange,
  hasNextPage,
  hasPrevPage,
  smoothTransitions = true,
  onEdit,
  onRequestDelete,
  emptyListAction,
  actions,
  taskStats,
}: Props) {
  const scheduleUiEnabled = !isUiFeatureOmitted("schedule");
  const statusDelayMs = smoothTransitions ? LOADING_STATUS_DELAY_MS : 0;
  const showLoadingLine = useDelayedTrue(loading, statusDelayMs);
  const appTimezone = useAppTimezone();

  const filters = useTaskListSectionFilters({
    tasks,
    projectFilterOptions,
    showProjectColumn,
    onListFiltersChange,
    smoothTransitions,
  });

  const bulk = useTaskListSectionBulkActions({
    filteredTasks: filters.filteredTasks,
    visibleIds: filters.visibleIds,
    scheduleUiEnabled,
  });

  const skipFiltersResetOnMount = useRef(true);
  const { clearSelection } = bulk.selection;

  useEffect(() => {
    if (skipFiltersResetOnMount.current) {
      skipFiltersResetOnMount.current = false;
      return;
    }
    onListFiltersChange();
    clearSelection();
  }, [
    filters.statusFilter,
    filters.priorityFilter,
    filters.projectFilter,
    filters.titleSearch,
    filters.sortKey,
    filters.sortDir,
    onListFiltersChange,
    clearSelection,
  ]);

  const showTaskPager =
    !loading && (hasPrevPage || hasNextPage || tasks.length === listPageSize);

  const headingSummary = useMemo(
    () => formatTaskListHeadingSummary(filters.filteredTasks.length, taskStats),
    [filters.filteredTasks.length, taskStats],
  );

  return (
    <section
      className="panel task-list-section-panel task-list-section--redesign"
      aria-labelledby="task-list-heading"
      id="task-list-panel"
      role="tabpanel"
    >
      <div className="task-list-toolbar">
        <TaskListSectionHeading actions={actions} summary={headingSummary} />
        {!loading ? (
          <>
            <TaskListStatusTabs
              value={filters.statusFilter}
              onChange={filters.setStatusFilter}
              stats={taskStats}
            />
            <TaskListFilters
              priorityFilter={filters.priorityFilter}
              onPriorityFilterChange={(v) =>
                filters.setPriorityFilter(v as typeof filters.priorityFilter)
              }
              projectFilter={filters.projectFilter}
              projectOptions={showProjectColumn ? projectFilterOptions : []}
              onProjectFilterChange={
                showProjectColumn ? filters.setProjectFilter : undefined
              }
              titleSearch={filters.titleSearch}
              onTitleSearchChange={filters.setTitleSearch}
              searchInputRef={filters.searchInputRef}
            />
          </>
        ) : null}
      </div>
      {refreshing && !loading && !hideBackgroundRefreshHint ? (
        <p className="sync-hint task-list-phase-msg" aria-live="polite" role="status">
          Syncing with server…
        </p>
      ) : null}
      {loading && showLoadingLine ? (
        <TaskListTableSkeleton caption={TASK_LIST_TABLE_CAPTION} />
      ) : null}
      {!loading ? (
        <div className="task-list-content task-list-content--enter">
          <TaskListDataTable
            caption={TASK_LIST_TABLE_CAPTION}
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
      <TaskListBulkActionBar
        selectedCount={bulk.selection.selectedVisibleIds.length}
        scheduledCount={bulk.selectedScheduledIds.length}
        rescheduleDisabled={bulk.selectedIncludesDone}
        showScheduleActions={scheduleUiEnabled}
        busy={bulk.bulkSchedule.isPending || bulk.bulkDelete.isPending}
        onReschedule={bulk.openRescheduleModal}
        onClearSchedule={bulk.handleClearSchedule}
        onDelete={bulk.openBulkDeleteModal}
        onCancel={bulk.handleCancelSelection}
      />
      {bulk.bulkDeleteModalOpen && bulk.selectedRowsForBulkDelete.length > 0 ? (
        <TaskBulkDeleteConfirmModal
          tasks={bulk.selectedRowsForBulkDelete}
          busy={bulk.bulkDelete.isPending}
          error={bulk.bulkDeleteError}
          onCancel={bulk.closeBulkDelete}
          onConfirm={bulk.handleBulkDeleteConfirm}
        />
      ) : null}
      {bulk.rescheduleModalOpen ? (
        <TaskBulkRescheduleModal
          selectedCount={bulk.selection.selectedVisibleIds.length}
          appTimezone={appTimezone}
          busy={bulk.bulkSchedule.isPending}
          error={bulk.bulkErrorBanner}
          onClose={bulk.closeReschedule}
          onSubmit={bulk.handleRescheduleSubmit}
        />
      ) : null}
    </section>
  );
});
