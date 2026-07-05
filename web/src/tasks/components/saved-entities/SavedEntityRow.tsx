import type { ReactNode } from "react";
import { formatRelativeTime } from "@/shared/time/relativeTime";
import { isSavedEntityRowActionExcluded } from "../../hooks/useDeleteWithExitAnimation";
import { TaskListDeleteGlyph } from "../task-list/table/TaskListRowActionIcons";

type Props = {
  name: string;
  lastEdited?: string;
  renderNow: Date;
  isDeleting: boolean;
  rowDisabled: boolean;
  isExiting: boolean;
  resumeLabel: string;
  deleteLabel: string;
  deletingLabel: string;
  onOpen: () => void;
  onDelete: () => void;
  leading?: ReactNode;
  subline?: ReactNode;
  trailing?: ReactNode;
  timePrefix?: string;
};

export function SavedEntityRow({
  name,
  lastEdited,
  renderNow,
  isDeleting,
  rowDisabled,
  isExiting,
  resumeLabel,
  deleteLabel,
  deletingLabel,
  onOpen,
  onDelete,
  leading,
  subline,
  trailing,
  timePrefix = "Edited",
}: Props) {
  const relative = lastEdited ? formatRelativeTime(lastEdited, renderNow) : null;

  return (
    <li
      className={[
        "draft-row",
        rowDisabled ? "" : "draft-row--interactive",
        isExiting ? "draft-row--exit" : "",
      ]
        .filter(Boolean)
        .join(" ")}
      onClick={(e) => {
        if (rowDisabled || isSavedEntityRowActionExcluded(e.target)) return;
        onOpen();
      }}
      onKeyDown={(e) => {
        if (rowDisabled || isSavedEntityRowActionExcluded(e.target)) return;
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onOpen();
        }
      }}
      tabIndex={rowDisabled ? undefined : 0}
      aria-label={resumeLabel}
    >
      {leading}
      <div className="draft-row__meta">
        <span className="draft-row__name" title={name}>
          {name}
        </span>
        {lastEdited && relative ? (
          <time className="draft-row__time" dateTime={lastEdited} title={lastEdited}>
            {timePrefix} {relative}
          </time>
        ) : null}
        {subline}
      </div>
      <div className="draft-row__actions">
        {trailing ?? (
          <div className="task-list-row-actions">
            <button
              type="button"
              className="task-list-icon-btn task-list-icon-btn--delete"
              aria-label={isDeleting ? deletingLabel : deleteLabel}
              onClick={() => void onDelete()}
              disabled={rowDisabled}
              aria-busy={isDeleting || undefined}
            >
              <TaskListDeleteGlyph />
            </button>
          </div>
        )}
      </div>
    </li>
  );
}
