import { CustomSelect } from "@/components/custom-select";
import { TASK_LIST_PRIORITY_FILTER_OPTIONS } from "./taskListFilterSelectOptions";
import type { RefObject } from "react";

type Props = {
  priorityFilter: string;
  onPriorityFilterChange: (value: string) => void;
  projectFilter?: string;
  projectOptions?: Array<{ id: string; name: string }>;
  onProjectFilterChange?: (value: string) => void;
  titleSearch: string;
  onTitleSearchChange: (value: string) => void;
  searchInputRef?: RefObject<HTMLInputElement>;
  showSearchShortcutHint?: boolean;
};

export function TaskListFilters({
  priorityFilter,
  onPriorityFilterChange,
  projectFilter = "all",
  projectOptions = [],
  onProjectFilterChange,
  titleSearch,
  onTitleSearchChange,
  searchInputRef,
  showSearchShortcutHint = true,
}: Props) {
  const projectFilterOptions = [
    { value: "all", label: "All projects" },
    ...projectOptions.map((project) => ({
      value: project.id,
      label: project.name,
    })),
  ];

  return (
    <div
      className="task-list-filters task-list-filters--redesign"
      role="search"
      aria-label="Filter tasks"
    >
      <div className="task-list-filters__controls">
        <div className="task-list-filter-field">
          <CustomSelect
            id="task-list-filter-priority"
            label="Priority"
            compact
            dropdownVariant="toolbar"
            dropdownMinWidth={200}
            listboxName="Filter by priority"
            value={priorityFilter}
            options={TASK_LIST_PRIORITY_FILTER_OPTIONS}
            onChange={onPriorityFilterChange}
          />
        </div>
        {onProjectFilterChange ? (
          <div className="task-list-filter-field task-list-filter-field--project">
            <CustomSelect
              id="task-list-filter-project"
              label="Project"
              compact
              dropdownVariant="toolbar"
              dropdownMinWidth={220}
              listboxName="Filter by project"
              value={projectFilter}
              options={projectFilterOptions}
              onChange={onProjectFilterChange}
            />
          </div>
        ) : null}
      </div>
      <div className="field grow task-list-search-field">
        <label htmlFor="task-list-search-title" className="visually-hidden">
          Search titles
        </label>
        <div className="task-list-search-field__inner">
          <input
            ref={searchInputRef}
            id="task-list-search-title"
            type="search"
            value={titleSearch}
            onChange={(e) => onTitleSearchChange(e.target.value)}
            placeholder="Search tasks…"
            autoComplete="off"
          />
          {showSearchShortcutHint && !titleSearch.trim() ? (
            <kbd className="task-list-search-kbd-hint" aria-hidden="true">
              /
            </kbd>
          ) : null}
        </div>
      </div>
    </div>
  );
}
