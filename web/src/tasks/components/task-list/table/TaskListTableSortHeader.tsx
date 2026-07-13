import type {
  TaskListSortDir,
  TaskListSortKey,
} from "../filters/taskListSort";

function sortAriaValue(
  key: TaskListSortKey,
  activeKey: TaskListSortKey | undefined,
  dir: TaskListSortDir | undefined,
): "none" | "ascending" | "descending" {
  if (activeKey !== key) return "none";
  return dir === "asc" ? "ascending" : "descending";
}

export function TaskListTableSortHeader({
  label,
  sortKey,
  activeSortKey,
  sortDir,
  onSortChange,
}: {
  label: string;
  sortKey: TaskListSortKey;
  activeSortKey?: TaskListSortKey;
  sortDir?: TaskListSortDir;
  onSortChange?: (key: TaskListSortKey) => void;
}) {
  if (!onSortChange) {
    return <th scope="col">{label}</th>;
  }
  const active = activeSortKey === sortKey;
  const icon = !active ? "↕" : sortDir === "asc" ? "↑" : "↓";
  return (
    <th scope="col">
      <button
        type="button"
        className="task-list-table-sort-btn"
        aria-sort={sortAriaValue(sortKey, activeSortKey, sortDir)}
        onClick={() => onSortChange(sortKey)}
      >
        {label}
        <span className="task-list-sort-icon" aria-hidden="true">
          {icon}
        </span>
      </button>
    </th>
  );
}
