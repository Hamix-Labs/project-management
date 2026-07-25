import { CustomSelect } from "@/components/custom-select";
import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import {
  TASK_LIST_PRIORITY_FILTER_OPTIONS,
  taskListStatusFilterOptions,
} from "./taskListFilterSelectOptions";
import type { RefObject } from "react";

type Props = {
  /** When false, hides the status dropdown (Closed tab owns that axis). */
  showStatusFilter?: boolean;
  statusFilter?: string;
  onStatusFilterChange?: (value: string) => void;
  priorityFilter: string;
  onPriorityFilterChange: (value: string) => void;
  projectFilter?: string;
  projectOptions?: Array<{ id: string; name: string }>;
  onProjectFilterChange?: (value: string) => void;
  tagFilter?: string;
  tagOptions?: string[];
  onTagFilterChange?: (value: string) => void;
  titleSearch: string;
  onTitleSearchChange: (value: string) => void;
  searchInputRef?: RefObject<HTMLInputElement>;
  showSearchShortcutHint?: boolean;
};

export function TaskListFilters({
  showStatusFilter = true,
  statusFilter = "all",
  onStatusFilterChange,
  priorityFilter,
  onPriorityFilterChange,
  projectFilter = "all",
  projectOptions = [],
  onProjectFilterChange,
  tagFilter = "all",
  tagOptions = [],
  onTagFilterChange,
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

  const tagFilterSelectOptions = [
    { value: "all", label: "All tags" },
    ...tagOptions.map((tag) => ({
      value: tag,
      label: tag,
    })),
  ];

  const scheduleUiEnabled = !isUiFeatureOmitted("schedule");
  const statusOptions = taskListStatusFilterOptions({
    includeScheduled: scheduleUiEnabled,
  });

  return (
    <div
      className="task-list-filters task-list-filters--redesign"
      role="search"
      aria-label="Filter tasks"
    >
      <div className="task-list-filters__controls">
        {showStatusFilter && onStatusFilterChange ? (
          <div className="task-list-filter-field">
            <CustomSelect
              id="task-list-filter-status"
              label="Status"
              compact
              dropdownVariant="toolbar"
              dropdownMinWidth={220}
              listboxName="Filter by status"
              value={statusFilter}
              options={statusOptions}
              onChange={onStatusFilterChange}
            />
          </div>
        ) : null}
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
        {onTagFilterChange && tagOptions.length > 0 ? (
          <div className="task-list-filter-field task-list-filter-field--tag">
            <CustomSelect
              id="task-list-filter-tag"
              label="Tag"
              compact
              dropdownVariant="toolbar"
              dropdownMinWidth={200}
              listboxName="Filter by tag"
              value={tagFilter}
              options={tagFilterSelectOptions}
              onChange={onTagFilterChange}
            />
          </div>
        ) : null}
      </div>
      <div className="field grow task-list-search-field">
        <label htmlFor="task-list-search-title" className="visually-hidden">
          Search titles
        </label>
        <div className="task-list-search-field__inner">
          <svg
            className="task-list-search-field__icon"
            width="14"
            height="14"
            viewBox="0 0 14 14"
            fill="none"
            aria-hidden="true"
          >
            <circle
              cx="6"
              cy="6"
              r="4.5"
              stroke="currentColor"
              strokeWidth="1.2"
            />
            <path
              d="M9.5 9.5L13 13"
              stroke="currentColor"
              strokeWidth="1.2"
              strokeLinecap="round"
            />
          </svg>
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
