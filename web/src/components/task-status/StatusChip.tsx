import type { HTMLAttributes } from "react";
import type { StatusTone } from "@/lib/taskStatusDisplay";

type Props = HTMLAttributes<HTMLSpanElement> & {
  tone: StatusTone;
  label: string;
  pulse?: boolean;
};

/** Presentational status chrome shared by task and cycle status badges. */
export function StatusChip({
  tone,
  label,
  pulse = false,
  className,
  ...rest
}: Props) {
  const toneClass = `task-status-badge--tone-${tone}`;
  const pulseClass = pulse ? "task-status-badge--pulse" : "";
  const rootClass = ["task-status-badge", toneClass, pulseClass, className]
    .filter(Boolean)
    .join(" ");

  return (
    <span className={rootClass} {...rest}>
      <span className="task-status-badge__dot" aria-hidden="true">
        <span className="task-status-badge__dot-core" />
      </span>
      {label}
    </span>
  );
}
