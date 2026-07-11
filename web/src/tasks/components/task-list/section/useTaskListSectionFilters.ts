import {
  useCallback,
  useMemo,
  useRef,
  useState,
  type RefObject,
} from "react";
import type { TaskWithDepth } from "../../../task-tree";
import {
  filterTasksForListView,
  type TaskListClientPriorityFilter,
  type TaskListClientStatusFilter,
} from "../filters/taskListClientFilter";
import {
  sortTasksForListView,
  type TaskListSortDir,
  type TaskListSortKey,
} from "../filters/taskListSort";
import { useTaskListSearchShortcut } from "../hooks/useTaskListSearchShortcut";

type UseTaskListSectionFiltersArgs = {
  tasks: TaskWithDepth[];
  projectFilterOptions: Array<{ id: string; name: string }>;
  showProjectColumn: boolean;
  onListFiltersChange: () => void;
  smoothTransitions: boolean;
};

export function useTaskListSectionFilters({
  tasks,
  projectFilterOptions,
  showProjectColumn,
  onListFiltersChange,
  smoothTransitions,
}: UseTaskListSectionFiltersArgs) {
  const projectNameById = useMemo(() => {
    const m: Record<string, string> = {};
    for (const p of projectFilterOptions) {
      m[p.id] = p.name;
    }
    return m;
  }, [projectFilterOptions]);

  const [statusFilter, setStatusFilter] =
    useState<TaskListClientStatusFilter>("all");
  const [priorityFilter, setPriorityFilter] =
    useState<TaskListClientPriorityFilter>("all");
  const [projectFilter, setProjectFilter] = useState("all");
  const [titleSearch, setTitleSearch] = useState("");
  const [sortKey, setSortKey] = useState<TaskListSortKey>("created_at");
  const [sortDir, setSortDir] = useState<TaskListSortDir>("desc");
  const searchInputRef = useRef<HTMLInputElement>(null);
  useTaskListSearchShortcut(searchInputRef, smoothTransitions);

  const filteredTasks = useMemo(() => {
    const base = filterTasksForListView(
      tasks,
      statusFilter,
      priorityFilter,
      titleSearch,
    );
    const scoped =
      projectFilter === "all"
        ? base
        : projectFilter === "none"
          ? base.filter((task) => !task.project_id)
          : base.filter((task) => task.project_id === projectFilter);
    return sortTasksForListView(scoped, sortKey, sortDir, projectNameById);
  }, [
    tasks,
    statusFilter,
    priorityFilter,
    titleSearch,
    projectFilter,
    sortKey,
    sortDir,
    projectNameById,
  ]);

  const handleSortChange = useCallback((key: TaskListSortKey) => {
    if (sortKey === key) {
      setSortDir((d) => (d === "asc" ? "desc" : "asc"));
      return;
    }
    setSortKey(key);
    setSortDir(key === "created_at" ? "desc" : "asc");
  }, [sortKey]);

  const visibleIds = useMemo(
    () => filteredTasks.map((t) => t.id),
    [filteredTasks],
  );

  return {
    projectNameById,
    statusFilter,
    setStatusFilter,
    priorityFilter,
    setPriorityFilter,
    projectFilter,
    setProjectFilter,
    titleSearch,
    setTitleSearch,
    sortKey,
    sortDir,
    searchInputRef: searchInputRef as RefObject<HTMLInputElement>,
    filteredTasks,
    handleSortChange,
    visibleIds,
    showProjectColumn,
    onListFiltersChange,
  };
}
