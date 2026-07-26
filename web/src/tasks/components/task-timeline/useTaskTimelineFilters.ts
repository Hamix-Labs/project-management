import { useCallback, useMemo, useRef, useState } from "react";
import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import {
  filterTasksByTag,
  uniqueSortedTagsFromTasks,
  type TaskListClientPriorityFilter,
} from "../task-list/filters/taskListClientFilter";
import { useTaskListSearchShortcut } from "../task-list/hooks/useTaskListSearchShortcut";
import type { TimelineEvent } from "./timelineTypes";

type Args = {
  events: TimelineEvent[];
  projectFilterOptions: Array<{ id: string; name: string }>;
  showProjectColumn: boolean;
  smoothTransitions?: boolean;
};

/** Client filters for the Timeline (board parity: priority / project / tag / search). */
export function useTaskTimelineFilters({
  events,
  projectFilterOptions,
  showProjectColumn,
  smoothTransitions = true,
}: Args) {
  const tagsUiEnabled = !isUiFeatureOmitted("taskTags");
  const [priorityFilter, setPriorityFilter] =
    useState<TaskListClientPriorityFilter>("all");
  const [projectFilter, setProjectFilter] = useState("all");
  const [tagFilter, setTagFilter] = useState("all");
  const [titleSearch, setTitleSearch] = useState("");
  const searchInputRef = useRef<HTMLInputElement>(null);
  useTaskListSearchShortcut(searchInputRef, smoothTransitions);

  const tagFilterOptions = useMemo(() => {
    if (!tagsUiEnabled) return [];
    return uniqueSortedTagsFromTasks(
      events.map((e) => ({ tags: e.taskTags })),
    );
  }, [tagsUiEnabled, events]);

  const filteredEvents = useMemo(() => {
    const q = titleSearch.trim().toLowerCase();
    let next = events;
    if (priorityFilter !== "all") {
      next = next.filter((e) => e.taskPriority === priorityFilter);
    }
    if (projectFilter !== "all") {
      next =
        projectFilter === "none"
          ? next.filter((e) => !e.taskProjectId)
          : next.filter((e) => e.taskProjectId === projectFilter);
    }
    if (q) {
      next = next.filter((e) =>
        (e.taskTitle ?? "").toLowerCase().includes(q),
      );
    }
    return filterTasksByTag(
      next.map((e) => ({ ...e, tags: e.taskTags })),
      tagsUiEnabled ? tagFilter : "all",
    );
  }, [
    events,
    priorityFilter,
    projectFilter,
    tagFilter,
    titleSearch,
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
    tagFilterOptions,
    projectFilterOptions,
    showProjectColumn,
    filteredEvents,
    hasClientFilters,
    onPriorityChange,
  };
}
