import type { ReactNode } from "react";

type Props = {
  /** Optional one-line summary under the title (e.g. "15 shown · 7 ready"). */
  summary?: string;
  /** Optional toolbar on the title row (e.g. home “New task”). */
  actions?: ReactNode;
  /** Section title; defaults to “All tasks”. */
  title?: string;
  /** `id` for the heading (a11y labelledby target). */
  titleId?: string;
};

export function TaskListSectionHeading({
  summary,
  actions,
  title = "All tasks",
  titleId = "task-list-heading",
}: Props) {
  return (
    <div className="task-list-section-head">
      <div className="task-list-section-head__text">
        <h2 id={titleId} className="task-list-section-title">
          {title}
        </h2>
        {summary ? (
          <p className="task-list-section-summary">{summary}</p>
        ) : null}
      </div>
      {actions ? (
        <div className="task-list-section-actions">{actions}</div>
      ) : null}
    </div>
  );
}
