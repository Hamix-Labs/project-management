import type { ReactNode } from "react";
import type { TaskStatsResponse } from "@/types";
import { TaskListFilters } from "../filters/TaskListFilters";
import { TaskListStatusTabs } from "../filters/TaskListStatusTabs";
import { TaskListSectionHeading } from "./TaskListSectionHeading";
import type { useTaskListSectionFilters } from "./useTaskListSectionFilters";

type Filters = ReturnType<typeof useTaskListSectionFilters>;

type Props = {
  loading: boolean;
  actions?: ReactNode;
  headingSummary: string | null;
  taskStats?: TaskStatsResponse | null;
  showProjectColumn: boolean;
  projectFilterOptions: Array<{ id: string; name: string }>;
  filters: Filters;
};

export function TaskListToolbar({
  loading,
  actions,
  headingSummary,
  taskStats,
  showProjectColumn,
  projectFilterOptions,
  filters,
}: Props) {
  return (
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
  );
}
