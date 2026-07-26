import { useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { useTaskDetailPrefetcher } from "@/app/hooks/usePrefetchOnIntent";
import type { Task } from "@/types";
import type { TaskWithDepth } from "../../../task-tree";
import type { CloseTargetInput } from "../../../hooks/useTaskCloseFlow";
import type {
  TaskListSortDir,
  TaskListSortKey,
} from "../filters/taskListSort";
import type { EmptyStateAction } from "@/shared/EmptyState";
import { TaskListTableBody } from "./TaskListTableBody";
import { TaskListTableHeader } from "./TaskListTableHeader";
import {
  syncHeaderCheckboxIndeterminate,
  type BulkSelectionProps,
} from "./taskListTableSelection";
import { useTaskListRowAnimations } from "./taskListRowAnimations";

type Props = {
  caption: string;
  refreshing: boolean;
  tasks: TaskWithDepth[];
  filteredTasks: TaskWithDepth[];
  saving: boolean;
  emptyListAction?: EmptyStateAction;
  onEdit: (t: Task) => void;
  onRequestClose: (t: CloseTargetInput) => void;
  selection?: BulkSelectionProps;
  projectNameById?: Record<string, string>;
  showProjectColumn?: boolean;
  showTagsColumn?: boolean;
  sortKey?: TaskListSortKey;
  sortDir?: TaskListSortDir;
  onSortChange?: (key: TaskListSortKey) => void;
};

export function TaskListDataTable({
  caption,
  refreshing,
  tasks,
  filteredTasks,
  saving,
  emptyListAction,
  onEdit,
  onRequestClose,
  selection,
  projectNameById = {},
  showProjectColumn = true,
  showTagsColumn = true,
  sortKey,
  sortDir,
  onSortChange,
}: Props) {
  const navigate = useNavigate();
  const prefetchTaskDetail = useTaskDetailPrefetcher();
  const headerCheckboxRef = useRef<HTMLInputElement | null>(null);
  const rowsToRender = useTaskListRowAnimations(filteredTasks, tasks);

  useEffect(() => {
    syncHeaderCheckboxIndeterminate(selection, headerCheckboxRef);
  }, [selection]);

  const showSelectionCol = Boolean(selection);
  const colSpan =
    (showSelectionCol ? 1 : 0) +
    5 +
    (showProjectColumn ? 1 : 0) +
    (showTagsColumn ? 1 : 0);
  return (
    <div className="table-wrap task-list-table-wrap">
      <table className="task-list-table" aria-busy={refreshing}>
        <caption className="visually-hidden">{caption}</caption>
        <colgroup>
          {showSelectionCol ? <col className="task-list-col-select" /> : null}
          <col className="task-list-col-title" />
          <col className="task-list-col-status" />
          <col className="task-list-col-priority" />
          {showTagsColumn ? <col className="task-list-col-tags" /> : null}
          <col className="task-list-col-created" />
          {showProjectColumn ? <col className="task-list-col-project" /> : null}
          <col className="task-list-col-actions" />
        </colgroup>
        <TaskListTableHeader
          showSelectionCol={showSelectionCol}
          showProjectColumn={showProjectColumn}
          showTagsColumn={showTagsColumn}
          selection={selection}
          headerCheckboxRef={headerCheckboxRef}
          filteredTasksLength={filteredTasks.length}
          sortKey={sortKey}
          sortDir={sortDir}
          onSortChange={onSortChange}
        />
        <TaskListTableBody
          tasksLength={tasks.length}
          rowsToRender={rowsToRender}
          colSpan={colSpan}
          emptyListAction={emptyListAction}
          showSelectionCol={showSelectionCol}
          showProjectColumn={showProjectColumn}
          showTagsColumn={showTagsColumn}
          selection={selection}
          projectNameById={projectNameById}
          saving={saving}
          onEdit={onEdit}
          onRequestClose={onRequestClose}
          prefetchTaskDetail={prefetchTaskDetail}
          navigate={navigate}
        />
      </table>
    </div>
  );
}

export type { BulkSelectionProps };
