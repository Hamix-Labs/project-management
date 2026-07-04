import type { TaskStatsResponse } from "@/types";
import {
  taskListStatusTabCount,
  taskListStatusTabs,
  type TaskListStatusTab,
} from "./taskListStatusTabConfig";
import type { TaskListClientStatusFilter } from "./taskListClientFilter";

type Props = {
  value: TaskListClientStatusFilter;
  onChange: (value: TaskListClientStatusFilter) => void;
  stats: TaskStatsResponse | null | undefined;
};

export function TaskListStatusTabs({ value, onChange, stats }: Props) {
  const tabs = taskListStatusTabs().filter((tab) => {
    if (tab.value === "all") return true;
    if (tab.value === value) return true;
    const count = taskListStatusTabCount(tab.value, stats);
    if (count === null) return true;
    return count > 0;
  });

  return (
    <div
      className="task-list-status-tabs"
      role="tablist"
      aria-label="Filter tasks by status"
    >
      {tabs.map((tab) => (
        <StatusTabButton
          key={tab.value}
          tab={tab}
          selected={value === tab.value}
          count={taskListStatusTabCount(tab.value, stats)}
          onSelect={() => onChange(tab.value)}
        />
      ))}
    </div>
  );
}

function StatusTabButton({
  tab,
  selected,
  count,
  onSelect,
}: {
  tab: TaskListStatusTab;
  selected: boolean;
  count: number | null;
  onSelect: () => void;
}) {
  const id = `task-list-status-tab-${tab.value}`;
  return (
    <button
      type="button"
      id={id}
      role="tab"
      aria-selected={selected}
      aria-controls="task-list-panel"
      className={
        selected
          ? "task-list-status-tab task-list-status-tab--active"
          : "task-list-status-tab"
      }
      onClick={onSelect}
    >
      <span className="task-list-status-tab__label">{tab.label}</span>
      {count !== null ? (
        <span className="task-list-status-tab__count" aria-hidden="true">
          {count}
        </span>
      ) : (
        <span className="task-list-status-tab__count" aria-hidden="true">
          —
        </span>
      )}
    </button>
  );
}
