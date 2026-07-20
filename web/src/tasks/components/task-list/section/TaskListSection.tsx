import {
  memo,
  useEffect,
  useMemo,
  useRef,
  type ReactNode,
} from "react";
import { useDelayedTrue } from "@/lib/useDelayedTrue";
import { formatTaskListHeadingSummary } from "./taskListHeadingSummary";
import type { Task, TaskStatsResponse } from "@/types";
import type { TaskWithDepth } from "../../../task-tree";
import type { DeleteTargetInput } from "../../../hooks/useTaskDeleteFlow";
import type { EmptyStateAction } from "@/shared/EmptyState";
import { useAppTimezone } from "@/shared/time/appTimezone";
import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import { useTaskListSectionFilters } from "./useTaskListSectionFilters";
import { useTaskListSectionBulkActions } from "./useTaskListSectionBulkActions";
import { TaskListToolbar } from "./TaskListToolbar";
import { TaskListTableRegion } from "./TaskListTableRegion";
import { TaskListBulkLayer } from "./TaskListBulkLayer";

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
      <TaskListToolbar
        loading={loading}
        actions={actions}
        headingSummary={headingSummary}
        taskStats={taskStats}
        showProjectColumn={showProjectColumn}
        projectFilterOptions={projectFilterOptions}
        filters={filters}
      />
      <TaskListTableRegion
        caption={TASK_LIST_TABLE_CAPTION}
        loading={loading}
        showLoadingLine={showLoadingLine}
        refreshing={refreshing}
        hideBackgroundRefreshHint={hideBackgroundRefreshHint}
        saving={saving}
        tasks={tasks}
        rootTasksOnPage={rootTasksOnPage}
        listPage={listPage}
        listPageSize={listPageSize}
        hasNextPage={hasNextPage}
        hasPrevPage={hasPrevPage}
        showTaskPager={showTaskPager}
        showProjectColumn={showProjectColumn}
        emptyListAction={emptyListAction}
        onEdit={onEdit}
        onRequestDelete={onRequestDelete}
        onListPageChange={onListPageChange}
        filters={filters}
        bulk={bulk}
      />
      <TaskListBulkLayer
        scheduleUiEnabled={scheduleUiEnabled}
        appTimezone={appTimezone}
        bulk={bulk}
      />
    </section>
  );
});
