import type { LegacyRef, RefObject } from "react";
import type {
  TaskListSortDir,
  TaskListSortKey,
} from "../filters/taskListSort";
import type { BulkSelectionProps } from "./taskListTableSelection";
import { TaskListTableSortHeader } from "./TaskListTableSortHeader";

export function TaskListTableHeader({
  showSelectionCol,
  showProjectColumn,
  showTagsColumn,
  selection,
  headerCheckboxRef,
  filteredTasksLength,
  sortKey,
  sortDir,
  onSortChange,
}: {
  showSelectionCol: boolean;
  showProjectColumn: boolean;
  showTagsColumn: boolean;
  selection: BulkSelectionProps | undefined;
  headerCheckboxRef: RefObject<HTMLInputElement | null>;
  filteredTasksLength: number;
  sortKey?: TaskListSortKey;
  sortDir?: TaskListSortDir;
  onSortChange?: (key: TaskListSortKey) => void;
}) {
  return (
    <thead>
      <tr>
        {showSelectionCol && selection ? (
          <th scope="col" className="task-list-select-col">
            <input
              ref={headerCheckboxRef as LegacyRef<HTMLInputElement>}
              type="checkbox"
              className="task-list-select-checkbox"
              aria-label={
                selection.allVisibleSelected
                  ? "Deselect all visible tasks"
                  : "Select all visible tasks"
              }
              checked={selection.allVisibleSelected}
              onChange={selection.onToggleAllVisible}
              data-testid="task-list-select-all"
              disabled={filteredTasksLength === 0}
            />
          </th>
        ) : null}
        <TaskListTableSortHeader
          label="Title"
          sortKey="title"
          activeSortKey={sortKey}
          sortDir={sortDir}
          onSortChange={onSortChange}
        />
        <TaskListTableSortHeader
          label="Status"
          sortKey="status"
          activeSortKey={sortKey}
          sortDir={sortDir}
          onSortChange={onSortChange}
        />
        <TaskListTableSortHeader
          label="Priority"
          sortKey="priority"
          activeSortKey={sortKey}
          sortDir={sortDir}
          onSortChange={onSortChange}
        />
        {showTagsColumn ? <th scope="col">Tags</th> : null}
        <TaskListTableSortHeader
          label="Created"
          sortKey="created_at"
          activeSortKey={sortKey}
          sortDir={sortDir}
          onSortChange={onSortChange}
        />
        {showProjectColumn ? (
          <TaskListTableSortHeader
            label="Project"
            sortKey="project"
            activeSortKey={sortKey}
            sortDir={sortDir}
            onSortChange={onSortChange}
          />
        ) : null}
        <th scope="col">Actions</th>
      </tr>
    </thead>
  );
}
