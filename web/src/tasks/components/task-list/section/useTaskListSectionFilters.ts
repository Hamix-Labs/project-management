import {
  useCallback,
  useMemo,
  useRef,
  useState,
  type RefObject,
} from "react";
import type { TaskWithDepth } from "../../../task-tree";
import {
  filterTasksByTag,
  filterTasksForListView,
  uniqueSortedTagsFromTasks,
  type TaskListClientPriorityFilter,
  type TaskListClientStatusFilter,
  type TaskListLifecycleFilter,
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
  /** When false, tag filter stays at "all" and options are empty (launch gate). */
  tagsUiEnabled: boolean;
  onListFiltersChange: () => void;
  smoothTransitions: boolean;
};

export function useTaskListSectionFilters({
  tasks,
  projectFilterOptions,
  showProjectColumn,
  tagsUiEnabled,
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

  const [lifecycleFilter, setLifecycleFilter] =
    useState<TaskListLifecycleFilter>("open");
  const [statusFilter, setStatusFilter] =
    useState<TaskListClientStatusFilter>("all");
  const [priorityFilter, setPriorityFilter] =
    useState<TaskListClientPriorityFilter>("all");
  const [projectFilter, setProjectFilter] = useState("all");
  const [tagFilter, setTagFilter] = useState("all");
  const [titleSearch, setTitleSearch] = useState("");
  const [sortKey, setSortKey] = useState<TaskListSortKey>("created_at");
  const [sortDir, setSortDir] = useState<TaskListSortDir>("desc");
  const searchInputRef = useRef<HTMLInputElement>(null);
  useTaskListSearchShortcut(searchInputRef, smoothTransitions);

  const tagFilterOptions = useMemo(() => {
    if (!tagsUiEnabled) return [];
    return uniqueSortedTagsFromTasks(tasks);
  }, [tagsUiEnabled, tasks]);

  const filteredTasks = useMemo(() => {
    const base = filterTasksForListView(
      tasks,
      lifecycleFilter,
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
    const tagged = filterTasksByTag(
      scoped,
      tagsUiEnabled ? tagFilter : "all",
    );
    return sortTasksForListView(tagged, sortKey, sortDir, projectNameById);
  }, [
    tasks,
    lifecycleFilter,
    statusFilter,
    priorityFilter,
    titleSearch,
    projectFilter,
    tagFilter,
    tagsUiEnabled,
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

  const handleLifecycleChange = useCallback(
    (next: TaskListLifecycleFilter) => {
      setLifecycleFilter(next);
      if (next === "closed") {
        setStatusFilter("all");
      }
    },
    [],
  );

  const visibleIds = useMemo(
    () => filteredTasks.map((t) => t.id),
    [filteredTasks],
  );

  return {
    projectNameById,
    lifecycleFilter,
    setLifecycleFilter: handleLifecycleChange,
    statusFilter,
    setStatusFilter,
    priorityFilter,
    setPriorityFilter,
    projectFilter,
    setProjectFilter,
    tagFilter,
    setTagFilter,
    tagFilterOptions,
    titleSearch,
    setTitleSearch,
    sortKey,
    sortDir,
    searchInputRef: searchInputRef as RefObject<HTMLInputElement>,
    filteredTasks,
    handleSortChange,
    visibleIds,
    showProjectColumn,
    tagsUiEnabled,
    onListFiltersChange,
  };
}
