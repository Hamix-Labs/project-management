import { useState, type ReactNode } from "react";

type Props = {
  title: string;
  headingId: string;
  /** Optional count pill shown next to the title. */
  count?: number | null;
  /** Initial open state; defaults to open. */
  defaultOpen?: boolean;
  className?: string;
  children: ReactNode;
  /** Outer landmark element. Defaults to `section`. */
  as?: "div" | "section";
  "data-testid"?: string;
};

/**
 * Full-width task-detail accordion: uppercase title + optional count on
 * the left, chevron on the right. Native details/summary so keyboard and
 * AT get disclosure behavior without a custom widget.
 */
export function TaskDetailCollapsibleSection({
  title,
  headingId,
  count = null,
  defaultOpen = true,
  className = "",
  children,
  as: Wrapper = "section",
  "data-testid": testId,
}: Props) {
  const [open, setOpen] = useState(defaultOpen);
  const sectionClass = ["task-detail-section", className]
    .filter(Boolean)
    .join(" ");

  return (
    <Wrapper
      className={sectionClass}
      aria-labelledby={headingId}
      data-testid={testId}
    >
      <details
        className="task-detail-collapsible"
        open={open}
        onToggle={(e) => {
          const next = (e.currentTarget as HTMLDetailsElement).open;
          if (next !== open) setOpen(next);
        }}
      >
        <summary className="task-detail-collapsible-summary">
          <span className="task-detail-collapsible-title-cluster">
            <h3 id={headingId} className="task-detail-section-heading">
              <span>{title}</span>
              {count != null ? (
                <span className="task-detail-section-count" aria-hidden="true">
                  {count}
                </span>
              ) : null}
            </h3>
          </span>
          <span
            className="task-detail-collapsible-chevron"
            aria-hidden="true"
          >
            ▾
          </span>
        </summary>
        <div className="task-detail-collapsible-body">{children}</div>
      </details>
    </Wrapper>
  );
}
