import { Link } from "react-router-dom";
import type { Task } from "@/types";
import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import type { CloseTargetInput } from "../../../hooks/useTaskCloseFlow";
import {
  canEditTask,
  PriorityBadge,
  StatusBadge,
} from "../../../task-display";
import { TASK_LIST_TAG_CHIP_LIMIT } from "../filters/taskListClientFilter";
import { TaskListDeleteGlyph, TaskListEditGlyph } from "./TaskListRowActionIcons";
import { formatInAppTimezone, useAppTimezone } from "@/shared/time/appTimezone";
import { formatRelativeTime } from "@/shared/time/relativeTime";
import { useNow } from "@/shared/useNow";
import type { TaskListRowRenderState } from "./taskListRowAnimations";
import type { BulkSelectionProps } from "./taskListTableSelection";

export type TaskListDataTableRowProps = {
  row: TaskListRowRenderState;
  showSelectionCol: boolean;
  showProjectColumn: boolean;
  selection: BulkSelectionProps | undefined;
  projectNameById: Record<string, string>;
  saving: boolean;
  onEdit: (t: Task) => void;
  onRequestClose: (t: CloseTargetInput) => void;
  prefetchTaskDetail: (id: string) => void;
  navigate: (path: string) => void;
};

function isTaskListRowNavExcluded(target: EventTarget | null): boolean {
  if (!(target instanceof Element)) return true;
  return Boolean(
    target.closest("a, button, input, select, textarea, label, [role='combobox']"),
  );
}

function TaskListRowTagChips({ tags }: { tags: string[] }) {
  const visible = tags.slice(0, TASK_LIST_TAG_CHIP_LIMIT);
  const overflow = tags.length - visible.length;
  return (
    <div className="task-list-row-tags" data-testid="task-list-row-tags">
      {visible.map((tag) => (
        <span key={tag} className="cell-pill task-list-tag-chip">
          {tag}
        </span>
      ))}
      {overflow > 0 ? (
        <span className="task-list-tag-overflow" aria-label={`${overflow} more tags`}>
          +{overflow}
        </span>
      ) : null}
    </div>
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
  onRequestClose,
  prefetchTaskDetail,
  navigate,
}: TaskListDataTableRowProps) {
  const projectLabel =
    showProjectColumn && t.project_id != null && t.project_id !== ""
      ? projectNameById[t.project_id]
      : undefined;
  const tagsUiEnabled = !isUiFeatureOmitted("taskTags");
  const rowTags = tagsUiEnabled ? (t.tags ?? []).filter(Boolean) : [];
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
            {rowTags.length > 0 ? <TaskListRowTagChips tags={rowTags} /> : null}
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
            aria-label={`Close task "${t.title}"`}
            onClick={() => onRequestClose(t)}
            disabled={saving || isExiting}
          >
            <TaskListDeleteGlyph />
          </button>
        </div>
      </td>
    </tr>
  );
}
