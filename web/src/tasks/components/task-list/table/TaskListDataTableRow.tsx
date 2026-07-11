import type { LegacyRef, RefObject } from "react";
import { Link } from "react-router-dom";
import type { Task } from "@/types";
import type { DeleteTargetInput } from "../../../hooks/useTaskDeleteFlow";
import {
  canEditTask,
  PriorityBadge,
  StatusBadge,
} from "../../../task-display";
import { TaskListDeleteGlyph, TaskListEditGlyph } from "./TaskListRowActionIcons";
import { taskListRowSubtitle } from "./taskListRowSubtitle";
import type {
  TaskListSortDir,
  TaskListSortKey,
} from "../filters/taskListSort";
import { previewTextFromPrompt } from "@/lib/promptFormat";
import { formatInAppTimezone, useAppTimezone } from "@/shared/time/appTimezone";
import { formatRelativeTime } from "@/shared/time/relativeTime";
import { useNow } from "@/shared/useNow";
import {
  EmptyState,
  EmptyStateFilterGlyph,
  type EmptyStateAction,
} from "@/shared/EmptyState";
import type { TaskListRowRenderState } from "./taskListRowAnimations";

export type BulkSelectionProps = {
  isSelected: (id: string) => boolean;
  onRowToggle: (id: string) => void;
  allVisibleSelected: boolean;
  someVisibleSelected: boolean;
  onToggleAllVisible: () => void;
};

export type TaskListDataTableRowProps = {
  row: TaskListRowRenderState;
  showSelectionCol: boolean;
  showProjectColumn: boolean;
  selection: BulkSelectionProps | undefined;
  projectNameById: Record<string, string>;
  saving: boolean;
  onEdit: (t: Task) => void;
  onRequestDelete: (t: DeleteTargetInput) => void;
  prefetchTaskDetail: (id: string) => void;
  navigate: (path: string) => void;
};

function isTaskListRowNavExcluded(target: EventTarget | null): boolean {
  if (!(target instanceof Element)) return true;
  return Boolean(
    target.closest("a, button, input, select, textarea, label, [role='combobox']"),
  );
}

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

export function TaskListTableHeader({
  showSelectionCol,
  showProjectColumn,
  selection,
  headerCheckboxRef,
  filteredTasksLength,
  sortKey,
  sortDir,
  onSortChange,
}: {
  showSelectionCol: boolean;
  showProjectColumn: boolean;
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
  onRequestDelete,
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
  onRequestDelete: (t: DeleteTargetInput) => void;
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
            onRequestDelete={onRequestDelete}
            prefetchTaskDetail={prefetchTaskDetail}
            navigate={navigate}
          />
        ))
      )}
    </tbody>
  );
}

export function TaskListDataTableRow({
  row: { task: t, isEntering, isExiting, isFilterExit },
  showSelectionCol,
  showProjectColumn,
  selection,
  projectNameById,
  saving,
  onEdit,
  onRequestDelete,
  prefetchTaskDetail,
  navigate,
}: TaskListDataTableRowProps) {
  const promptPreview = previewTextFromPrompt(t.initial_prompt);
  const projectLabel =
    showProjectColumn && t.project_id != null && t.project_id !== ""
      ? projectNameById[t.project_id]
      : undefined;
  const titleSubtitle = taskListRowSubtitle({ promptPreview });
  const rowSelected = !isExiting && selection ? selection.isSelected(t.id) : false;
  const rowClass = [
    "task-list-row",
    isEntering ? "task-list-row--enter" : "",
    isExiting ? "task-list-row--exit" : "",
    isFilterExit ? "task-list-row--filter-exit" : "",
    !isExiting ? "task-list-row--navigable" : "",
  ]
    .filter(Boolean)
    .join(" ");
  const taskHref = `/tasks/${t.id}`;
  const onIntent = isExiting ? undefined : () => prefetchTaskDetail(t.id);
  const appTimezone = useAppTimezone();
  const now = useNow({ intervalMs: 60_000 });
  const createdLabel = t.created_at
    ? formatRelativeTime(t.created_at, new Date(now))
    : "";
  const createdTitle = t.created_at
    ? formatInAppTimezone(t.created_at, appTimezone)
    : undefined;

  return (
    <tr
      key={t.id}
      className={rowClass}
      data-selected={rowSelected ? "true" : undefined}
      aria-hidden={isExiting ? "true" : undefined}
      onPointerEnter={onIntent}
      onFocus={onIntent}
      onClick={
        isExiting
          ? undefined
          : (e) => {
              if (isTaskListRowNavExcluded(e.target)) return;
              navigate(taskHref);
            }
      }
    >
      {showSelectionCol && selection ? (
        <td className="task-list-select-col">
          <input
            type="checkbox"
            className="task-list-select-checkbox"
            aria-label={
              rowSelected
                ? `Deselect task "${t.title}"`
                : `Select task "${t.title}"`
            }
            checked={rowSelected}
            onChange={() => selection.onRowToggle(t.id)}
            data-testid={`task-list-select-row-${t.id}`}
            disabled={isExiting}
          />
        </td>
      ) : null}
      <td className="cell-title">
        <Link
          to={taskHref}
          className={["cell-title-link", "cell-title-link--cell"].filter(Boolean).join(" ")}
          aria-label={`Open task details: ${t.title}`}
        >
          <div className="cell-title-stack">
            <span className="cell-title-main">
              <span className="cell-title-text cell-title-text--primary">{t.title}</span>
              <span className="cell-title-open-hint" aria-hidden="true">
                →
              </span>
            </span>
            {titleSubtitle ? <div className="cell-title-sub">{titleSubtitle}</div> : null}
          </div>
        </Link>
      </td>
      <td className="cell-status">
        <StatusBadge status={t.status} />
      </td>
      <td className="cell-priority">
        <PriorityBadge priority={t.priority} />
      </td>
      <td className="cell-created">
        {createdLabel ? (
          <time dateTime={t.created_at} title={createdTitle}>
            {createdLabel}
          </time>
        ) : (
          <span className="task-list-created-empty">—</span>
        )}
      </td>
      {showProjectColumn ? (
        <td className="cell-project">
          {projectLabel ? (
            <span className="task-list-project-name">{projectLabel}</span>
          ) : (
            <span className="task-list-project-empty">—</span>
          )}
        </td>
      ) : null}
      <td className="cell-actions">
        <div className="task-list-row-actions">
          <button
            type="button"
            className="task-list-icon-btn task-list-icon-btn--edit"
            aria-label={
              canEditTask(t.status)
                ? `Edit task "${t.title}"`
                : `Cannot edit task "${t.title}" while in progress`
            }
            title={
              canEditTask(t.status) ? undefined : "Cannot edit while the task is in progress"
            }
            onClick={() => onEdit(t)}
            disabled={saving || isExiting || !canEditTask(t.status)}
          >
            <TaskListEditGlyph />
          </button>
          <button
            type="button"
            className="task-list-icon-btn task-list-icon-btn--delete"
            aria-label={`Delete task "${t.title}"`}
            onClick={() => onRequestDelete(t)}
            disabled={saving || isExiting}
          >
            <TaskListDeleteGlyph />
          </button>
        </div>
      </td>
    </tr>
  );
}

export function syncHeaderCheckboxIndeterminate(
  selection: BulkSelectionProps | undefined,
  headerCheckboxRef: RefObject<HTMLInputElement | null>,
): void {
  if (!selection || !headerCheckboxRef.current) return;
  headerCheckboxRef.current.indeterminate = selection.someVisibleSelected;
}
