import { useCallback, useMemo, useRef, useState } from "react";
import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import {
  filterTasksByTag,
  filterTasksForListView,
  uniqueSortedTagsFromTasks,
  type TaskListClientPriorityFilter,
} from "../task-list/filters/taskListClientFilter";
import { useTaskListSearchShortcut } from "../task-list/hooks/useTaskListSearchShortcut";
import type { TaskWithDepth } from "../../task-tree";

type Args = {
  tasks: TaskWithDepth[];
  projectFilterOptions: Array<{ id: string; name: string }>;
  showProjectColumn: boolean;
  smoothTransitions: boolean;
};

/** Client filters for the board (no status tabs — columns own status). */
export function useTaskBoardFilters({
  tasks,
  projectFilterOptions,
  showProjectColumn,
  smoothTransitions,
}: Args) {
  const tagsUiEnabled = !isUiFeatureOmitted("taskTags");
  const [priorityFilter, setPriorityFilter] =
    useState<TaskListClientPriorityFilter>("all");
  const [projectFilter, setProjectFilter] = useState("all");
  const [tagFilter, setTagFilter] = useState("all");
  const [titleSearch, setTitleSearch] = useState("");
  const searchInputRef = useRef<HTMLInputElement>(null);
  useTaskListSearchShortcut(searchInputRef, smoothTransitions);

  const projectNameById = useMemo(() => {
    const m: Record<string, string> = {};
    for (const p of projectFilterOptions) {
      m[p.id] = p.name;
    }
    return m;
  }, [projectFilterOptions]);

  const tagFilterOptions = useMemo(() => {
    if (!tagsUiEnabled) return [];
    return uniqueSortedTagsFromTasks(tasks);
  }, [tagsUiEnabled, tasks]);

  const filteredTasks = useMemo(() => {
    const base = filterTasksForListView(
      tasks,
      "open",
      "all",
      priorityFilter,
      titleSearch,
    );
    const scoped =
      projectFilter === "all"
        ? base
        : projectFilter === "none"
          ? base.filter((task) => !task.project_id)
          : base.filter((task) => task.project_id === projectFilter);
    return filterTasksByTag(scoped, tagsUiEnabled ? tagFilter : "all");
  }, [
    tasks,
    priorityFilter,
    titleSearch,
    projectFilter,
    tagFilter,
    tagsUiEnabled,
  ]);

  const hasClientFilters =
    priorityFilter !== "all" ||
    projectFilter !== "all" ||
    tagFilter !== "all" ||
    titleSearch.trim() !== "";

  const onPriorityChange = useCallback((v: string) => {
    setPriorityFilter(v as TaskListClientPriorityFilter);
  }, []);

  return {
    tagsUiEnabled,
    priorityFilter,
    projectFilter,
    setProjectFilter,
    tagFilter,
    setTagFilter,
    titleSearch,
    setTitleSearch,
    searchInputRef,
    projectNameById,
    tagFilterOptions,
    filteredTasks,
    hasClientFilters,
    onPriorityChange,
    showProjectColumn,
  };
}
