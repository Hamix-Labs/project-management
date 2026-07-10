import { useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import type { LegacyRef, MutableRefObject, RefObject } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useTaskDetailPrefetcher } from "@/app/hooks/usePrefetchOnIntent";
import type { Task } from "@/types";
import type { TaskWithDepth } from "../../../task-tree";
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
import { computeTaskListDisplayOrder } from "./taskListDisplayOrder";

/**
 * Matches the `task-list-row-fade-out` keyframe duration in
 * app-task-list-and-mentions.css (--duration-normal ≈ 200ms).
 * A hair longer than --duration-normal so we don't yank the row
 * off in the final frame. Kept in JS so the cleanup timer doesn't
 * fight the CSS fallback for users whose onAnimationEnd never
 * fires (e.g. tab backgrounded mid-exit).
 */
const ROW_EXIT_MS = 220;

function isTaskListRowNavExcluded(target: EventTarget | null): boolean {
  if (!(target instanceof Element)) return true;
  return Boolean(
    target.closest("a, button, input, select, textarea, label, [role='combobox']"),
  );
}

/**
 * Optional bulk-selection bindings. When omitted, the table renders
 * without the leftmost checkbox column (callers that don't want
 * bulk actions get the historical layout for free). When provided, the table renders
 * a header tri-state checkbox plus a per-row checkbox column;
 * the parent owns the state via `useTaskListSelection`.
 */
type BulkSelectionProps = {
  isSelected: (id: string) => boolean;
  onRowToggle: (id: string) => void;
  allVisibleSelected: boolean;
  someVisibleSelected: boolean;
  onToggleAllVisible: () => void;
};

type Props = {
  caption: string;
  /** Reflects background refetch while the table stays visible. */
  refreshing: boolean;
  tasks: TaskWithDepth[];
  filteredTasks: TaskWithDepth[];
  saving: boolean;
  emptyListAction?: EmptyStateAction;
  onEdit: (t: Task) => void;
  onRequestDelete: (t: DeleteTargetInput) => void;
  /**
   * Optional bulk-selection bindings (Stage 5 of task scheduling).
   * Omit to keep the legacy no-checkbox layout for callers that
   * don't need bulk actions.
   */
  selection?: BulkSelectionProps;
  /** Maps `task.project_id` to a label for the Project column (e.g. from `GET /projects`). */
  projectNameById?: Record<string, string>;
  showProjectColumn?: boolean;
  sortKey?: TaskListSortKey;
  sortDir?: TaskListSortDir;
  onSortChange?: (key: TaskListSortKey) => void;
};

type ExitingRow = { task: TaskWithDepth; timeoutId: number };

type TaskListRowRenderState = {
  task: TaskWithDepth;
  isEntering: boolean;
  isExiting: boolean;
  isFilterExit: boolean;
};

type TaskListRowAnimationRefs = {
  seenIdsRef: MutableRefObject<Set<string>>;
  exitingRef: MutableRefObject<Map<string, ExitingRow>>;
  filterExitingRef: MutableRefObject<Map<string, TaskWithDepth>>;
  displayOrderRef: MutableRefObject<TaskWithDepth[]>;
  prevFilteredRef: MutableRefObject<TaskWithDepth[]>;
};

function syncHeaderCheckboxIndeterminate(
  selection: BulkSelectionProps | undefined,
  headerCheckboxRef: RefObject<HTMLInputElement | null>,
): void {
  if (!selection || !headerCheckboxRef.current) return;
  headerCheckboxRef.current.indeterminate = selection.someVisibleSelected;
}

function scheduleRemovedTaskExits(
  prevOrder: TaskWithDepth[],
  tasksIds: Set<string>,
  exitingRef: MutableRefObject<Map<string, ExitingRow>>,
  seenIdsRef: MutableRefObject<Set<string>>,
  setExitingTick: (updater: (value: number) => number) => void,
): void {
  for (const pr of prevOrder) {
    if (tasksIds.has(pr.id)) continue;
    if (exitingRef.current.has(pr.id)) continue;
    const timeoutId = window.setTimeout(() => {
      exitingRef.current.delete(pr.id);
      seenIdsRef.current.delete(pr.id);
      setExitingTick((x) => x + 1);
    }, ROW_EXIT_MS);
    exitingRef.current.set(pr.id, { task: pr, timeoutId });
  }
}

function scheduleFilterRemovedRowExits(
  prevOrder: TaskWithDepth[],
  nextIds: Set<string>,
  tasksIds: Set<string>,
  filterExitingRef: MutableRefObject<Map<string, TaskWithDepth>>,
  displayOrderRef: MutableRefObject<TaskWithDepth[]>,
  seenIdsRef: MutableRefObject<Set<string>>,
  setExitingTick: (updater: (value: number) => number) => void,
): boolean {
  let scheduledFilterExit = false;
  for (const t of prevOrder) {
    if (nextIds.has(t.id)) continue;
    if (!tasksIds.has(t.id)) continue;
    if (filterExitingRef.current.has(t.id)) continue;
    filterExitingRef.current.set(t.id, t);
    window.setTimeout(() => {
      filterExitingRef.current.delete(t.id);
      displayOrderRef.current = displayOrderRef.current.filter(
        (row) => row.id !== t.id,
      );
      seenIdsRef.current.delete(t.id);
      setExitingTick((x) => x + 1);
    }, ROW_EXIT_MS);
    scheduledFilterExit = true;
  }
  return scheduledFilterExit;
}

function syncTaskListEnteringIds(
  filteredTasks: TaskWithDepth[],
  filteredIds: Set<string>,
  refs: TaskListRowAnimationRefs,
  setEnteringIds: (updater: (prev: Set<string>) => Set<string>) => void,
): void {
  const newlyEntering = new Set<string>();
  for (const t of filteredTasks) {
    if (!refs.seenIdsRef.current.has(t.id)) {
      newlyEntering.add(t.id);
      refs.seenIdsRef.current.add(t.id);
    }
    const pendingExit = refs.exitingRef.current.get(t.id);
    if (pendingExit) {
      clearTimeout(pendingExit.timeoutId);
      refs.exitingRef.current.delete(t.id);
    }
    refs.filterExitingRef.current.delete(t.id);
  }

  const clientExitIds = new Set(refs.filterExitingRef.current.keys());
  for (const id of Array.from(refs.seenIdsRef.current)) {
    if (filteredIds.has(id)) continue;
    if (refs.exitingRef.current.has(id)) continue;
    if (clientExitIds.has(id)) continue;
    refs.seenIdsRef.current.delete(id);
  }

  setEnteringIds((prevEntering) => {
    if (prevEntering.size === 0 && newlyEntering.size === 0) return prevEntering;
    return newlyEntering;
  });
}

function buildTaskListRowsToRender(
  renderOrder: TaskWithDepth[],
  filteredTasks: TaskWithDepth[],
  filteredMap: Map<string, TaskWithDepth>,
  filterExitIds: Set<string>,
  filterExitingRef: MutableRefObject<Map<string, TaskWithDepth>>,
  exitingRef: MutableRefObject<Map<string, ExitingRow>>,
  enteringIds: Set<string>,
): TaskListRowRenderState[] {
  const rowsToRender: TaskListRowRenderState[] = [];
  const processed = new Set<string>();
  for (const t of renderOrder) {
    const visible = filteredMap.get(t.id);
    if (visible) {
      rowsToRender.push({
        task: visible,
        isEntering: enteringIds.has(t.id),
        isExiting: false,
        isFilterExit: false,
      });
      processed.add(t.id);
      continue;
    }
    if (filterExitIds.has(t.id)) {
      const exitingTask = filterExitingRef.current.get(t.id) ?? t;
      rowsToRender.push({
        task: exitingTask,
        isEntering: false,
        isExiting: true,
        isFilterExit: true,
      });
      processed.add(t.id);
    }
  }
  for (const t of filteredTasks) {
    if (processed.has(t.id)) continue;
    rowsToRender.push({
      task: t,
      isEntering: enteringIds.has(t.id),
      isExiting: false,
      isFilterExit: false,
    });
    processed.add(t.id);
  }
  for (const { task } of exitingRef.current.values()) {
    if (processed.has(task.id)) continue;
    rowsToRender.push({
      task,
      isEntering: false,
      isExiting: true,
      isFilterExit: false,
    });
  }
  return rowsToRender;
}

function useTaskListRowAnimations(filteredTasks: TaskWithDepth[], tasks: TaskWithDepth[]) {
  const seenIdsRef = useRef<Set<string>>(new Set());
  const [enteringIds, setEnteringIds] = useState<Set<string>>(new Set());
  const exitingRef = useRef<Map<string, ExitingRow>>(new Map());
  const filterExitingRef = useRef<Map<string, TaskWithDepth>>(new Map());
  const displayOrderRef = useRef<TaskWithDepth[]>([]);
  const prevFilteredRef = useRef<TaskWithDepth[]>([]);
  const [exitingTick, setExitingTick] = useState(0);

  const filteredIds = useMemo(
    () => new Set(filteredTasks.map((t) => t.id)),
    [filteredTasks],
  );
  const tasksIds = useMemo(() => new Set(tasks.map((t) => t.id)), [tasks]);
  const animationRefs: TaskListRowAnimationRefs = {
    seenIdsRef,
    exitingRef,
    filterExitingRef,
    displayOrderRef,
    prevFilteredRef,
  };

  useLayoutEffect(() => {
    const prevOrder =
      displayOrderRef.current.length > 0
        ? displayOrderRef.current
        : prevFilteredRef.current;
    const nextIds = new Set(filteredTasks.map((t) => t.id));
    const scheduledFilterExit = scheduleFilterRemovedRowExits(
      prevOrder,
      nextIds,
      tasksIds,
      filterExitingRef,
      displayOrderRef,
      seenIdsRef,
      setExitingTick,
    );

    for (const t of filteredTasks) {
      filterExitingRef.current.delete(t.id);
    }

    scheduleRemovedTaskExits(
      prevOrder,
      tasksIds,
      exitingRef,
      seenIdsRef,
      setExitingTick,
    );

    displayOrderRef.current = computeTaskListDisplayOrder(
      prevOrder,
      filteredTasks,
      new Set(filterExitingRef.current.keys()),
      filterExitingRef.current,
    );
    prevFilteredRef.current = filteredTasks;
    if (scheduledFilterExit) {
      setExitingTick((x) => x + 1);
    }
  }, [filteredTasks, tasksIds, filteredIds]);

  useEffect(() => {
    syncTaskListEnteringIds(filteredTasks, filteredIds, animationRefs, setEnteringIds);
  }, [filteredTasks, filteredIds, tasksIds]);

  useEffect(() => {
    const exiting = exitingRef.current;
    return () => {
      for (const { timeoutId } of exiting.values()) {
        clearTimeout(timeoutId);
      }
      exiting.clear();
    };
  }, []);

  const filteredMap = useMemo(
    () => new Map(filteredTasks.map((t) => [t.id, t])),
    [filteredTasks],
  );
  const filterExitIds = useMemo(
    () => new Set(filterExitingRef.current.keys()),
    // eslint-disable-next-line react-hooks/exhaustive-deps -- ref map is the source of truth
    [exitingTick, filteredTasks],
  );
  const renderOrder =
    displayOrderRef.current.length > 0
      ? displayOrderRef.current
      : filteredTasks;
  const rowsToRender = buildTaskListRowsToRender(
    renderOrder,
    filteredTasks,
    filteredMap,
    filterExitIds,
    filterExitingRef,
    exitingRef,
    enteringIds,
  );
  void exitingTick;

  return rowsToRender;
}

type TaskListDataTableRowProps = {
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

function sortAriaValue(
  key: TaskListSortKey,
  activeKey: TaskListSortKey | undefined,
  dir: TaskListSortDir | undefined,
): "none" | "ascending" | "descending" {
  if (activeKey !== key) return "none";
  return dir === "asc" ? "ascending" : "descending";
}

function TaskListTableSortHeader({
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

function TaskListTableHeader({
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

function TaskListTableBody({
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

function TaskListDataTableRow({
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
    showProjectColumn &&
    t.project_id != null &&
    t.project_id !== ""
      ? projectNameById[t.project_id]
      : undefined;
  const titleSubtitle = taskListRowSubtitle({
    promptPreview,
  });
  const rowSelected =
    !isExiting && selection ? selection.isSelected(t.id) : false;
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
          className={["cell-title-link", "cell-title-link--cell"]
            .filter(Boolean)
            .join(" ")}
          aria-label={`Open task details: ${t.title}`}
        >
          <div className="cell-title-stack">
            <span className="cell-title-main">
              <span className="cell-title-text cell-title-text--primary">
                {t.title}
              </span>
              <span className="cell-title-open-hint" aria-hidden="true">
                →
              </span>
            </span>
            {titleSubtitle ? (
              <div className="cell-title-sub">{titleSubtitle}</div>
            ) : null}
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
              canEditTask(t.status)
                ? undefined
                : "Cannot edit while the task is in progress"
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

export function TaskListDataTable({
  caption,
  refreshing,
  tasks,
  filteredTasks,
  saving,
  emptyListAction,
  onEdit,
  onRequestDelete,
  selection,
  projectNameById = {},
  showProjectColumn = true,
  sortKey,
  sortDir,
  onSortChange,
}: Props) {
  const navigate = useNavigate();
  const prefetchTaskDetail = useTaskDetailPrefetcher();
  const headerCheckboxRef = useRef<HTMLInputElement | null>(null);
  const rowsToRender = useTaskListRowAnimations(filteredTasks, tasks);

  useEffect(() => {
    syncHeaderCheckboxIndeterminate(selection, headerCheckboxRef);
  }, [selection]);

  const showSelectionCol = Boolean(selection);
  const colSpan =
    (showSelectionCol ? 1 : 0) + 5 + (showProjectColumn ? 1 : 0);
  return (
    <div className="table-wrap task-list-table-wrap">
      <table className="task-list-table" aria-busy={refreshing}>
        <caption className="visually-hidden">{caption}</caption>
        <colgroup>
          {showSelectionCol ? <col className="task-list-col-select" /> : null}
          <col className="task-list-col-title" />
          <col className="task-list-col-status" />
          <col className="task-list-col-priority" />
          <col className="task-list-col-created" />
          {showProjectColumn ? <col className="task-list-col-project" /> : null}
          <col className="task-list-col-actions" />
        </colgroup>
        <TaskListTableHeader
          showSelectionCol={showSelectionCol}
          showProjectColumn={showProjectColumn}
          selection={selection}
          headerCheckboxRef={headerCheckboxRef}
          filteredTasksLength={filteredTasks.length}
          sortKey={sortKey}
          sortDir={sortDir}
          onSortChange={onSortChange}
        />
        <TaskListTableBody
          tasksLength={tasks.length}
          rowsToRender={rowsToRender}
          colSpan={colSpan}
          emptyListAction={emptyListAction}
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
      </table>
    </div>
  );
}
