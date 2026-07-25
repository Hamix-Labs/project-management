import type { ReactNode } from "react";

type Props = {
  /**
   * Optional summary under the title when a string, or inline beside the
   * title when a React node (e.g. board active pill).
   */
  summary?: ReactNode;
  /** Optional description under the title row. */
  description?: string;
  /** Optional toolbar on the title row (e.g. home “New task”). */
  actions?: ReactNode;
  /** Section title; defaults to “All tasks”. */
  title?: string;
  /** `id` for the heading (a11y labelledby target). */
  titleId?: string;
};

export function TaskListSectionHeading({
  summary,
  description,
  actions,
  title = "All tasks",
  titleId = "task-list-heading",
}: Props) {
  const inlineSummary =
    summary != null && typeof summary !== "string" ? summary : null;
  const textSummary = typeof summary === "string" ? summary : null;

  return (
    <div className="task-list-section-head">
      <div className="task-list-section-head__text">
        <div className="task-list-section-head__title-row">
          <h2 id={titleId} className="task-list-section-title">
            {title}
          </h2>
          {inlineSummary}
        </div>
        {textSummary ? (
          <p className="task-list-section-summary">{textSummary}</p>
        ) : null}
        {description ? (
          <p className="task-list-section-description">{description}</p>
        ) : null}
      </div>
      {actions ? (
        <div className="task-list-section-actions">{actions}</div>
      ) : null}
    </div>
  );
}
