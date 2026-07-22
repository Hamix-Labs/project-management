import type { HTMLAttributes } from "react";
import type { Status } from "@/types";
import { STATUS_META } from "@/lib/taskStatusDisplay";

type Props = HTMLAttributes<HTMLSpanElement> & {
  status: Status;
};

export function StatusBadge({ status, className, ...rest }: Props) {
  const meta = STATUS_META[status];
  const toneClass = `task-status-badge--tone-${meta.tone}`;
  const pulseClass = meta.pulse ? "task-status-badge--pulse" : "";
  const rootClass = ["task-status-badge", toneClass, pulseClass, className]
    .filter(Boolean)
    .join(" ");

  return (
    <span className={rootClass} {...rest}>
      <span className="task-status-badge__dot" aria-hidden="true">
        <span className="task-status-badge__dot-core" />
      </span>
      {meta.label}
    </span>
  );
}
