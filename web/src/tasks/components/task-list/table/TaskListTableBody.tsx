import type { Task } from "@/types";
import type { CloseTargetInput } from "../../../hooks/useTaskCloseFlow";
import {
  EmptyState,
  EmptyStateFilterGlyph,
  type EmptyStateAction,
} from "@/shared/EmptyState";
import type { TaskListRowRenderState } from "./taskListRowAnimations";
import type { BulkSelectionProps } from "./taskListTableSelection";
import { TaskListDataTableRow } from "./TaskListDataTableRow";

export function TaskListTableBody({
  tasksLength,
  rowsToRender,
  colSpan,
  emptyListAction,
  showSelectionCol,
  showProjectColumn,
  selection,
  projectNameById,
  saving,
  onEdit,
  onRequestClose,
  prefetchTaskDetail,
  navigate,
}: {
  tasksLength: number;
  rowsToRender: TaskListRowRenderState[];
  colSpan: number;
  emptyListAction?: EmptyStateAction;
  showSelectionCol: boolean;
  showProjectColumn: boolean;
  selection: BulkSelectionProps | undefined;
  projectNameById: Record<string, string>;
  saving: boolean;
  onEdit: (t: Task) => void;
  onRequestClose: (t: CloseTargetInput) => void;
  prefetchTaskDetail: (id: string) => void;
  navigate: (path: string) => void;
}) {
  return (
    <tbody className="task-list-tbody">
      {tasksLength === 0 ? (
        <tr className="task-list-empty-row">
          <td colSpan={colSpan} className="task-list-empty-cell">
            <EmptyState
              className="empty-state--in-table empty-state--task-list-fresh"
              title="No tasks yet"
              description=""
              hideIcon
              action={emptyListAction}
            />
          </td>
        </tr>
      ) : rowsToRender.length === 0 ? (
        <tr className="task-list-empty-row">
          <td colSpan={colSpan} className="task-list-empty-cell">
            <EmptyState
              className="empty-state--in-table"
              icon={<EmptyStateFilterGlyph />}
              title="No matching tasks"
              description=""
              hideIcon={false}
            />
          </td>
        </tr>
      ) : (
        rowsToRender.map((row) => (
          <TaskListDataTableRow
            key={row.task.id}
            row={row}
            showSelectionCol={showSelectionCol}
            showProjectColumn={showProjectColumn}
            selection={selection}
            projectNameById={projectNameById}
            saving={saving}
            onEdit={onEdit}
            onRequestClose={onRequestClose}
            prefetchTaskDetail={prefetchTaskDetail}
            navigate={navigate}
          />
        ))
      )}
    </tbody>
  );
}
