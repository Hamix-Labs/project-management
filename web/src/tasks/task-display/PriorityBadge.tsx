import type { Priority } from "@/types";
import { PRIORITY_META } from "./priorityMeta";

type Props = {
  priority: Priority;
  className?: string;
};

const BAR_LEVELS = [1, 2, 3, 4] as const;

export function PriorityBadge({ priority, className }: Props) {
  const meta = PRIORITY_META[priority];
  const rootClass = [
    "task-priority-badge",
    `task-priority-badge--tone-${meta.tone}`,
    className,
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <span className={rootClass}>
      <span className="task-priority-badge__meter" aria-hidden="true">
        {BAR_LEVELS.map((level) => (
          <span
            key={level}
            className={
              level <= meta.weight
                ? "task-priority-badge__bar task-priority-badge__bar--filled"
                : "task-priority-badge__bar"
            }
            data-level={level}
          />
        ))}
      </span>
      {meta.label}
    </span>
  );
}
