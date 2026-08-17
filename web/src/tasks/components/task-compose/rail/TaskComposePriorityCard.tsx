import { PRIORITIES, type Priority, type PriorityChoice } from "@/types";

type Props = {
  value: PriorityChoice;
  disabled?: boolean;
  onChange: (p: PriorityChoice) => void;
};

const LABELS: Record<Priority, string> = {
  low: "Low",
  medium: "Medium",
  high: "High",
  critical: "Urgent",
};

/**
 * Segmented priority control for the compose rail.
 * Values stay on the domain Priority enum (critical, not "urgent").
 */
export function TaskComposePriorityCard({
  value,
  disabled = false,
  onChange,
}: Props) {
  return (
    <section
      className="compose-handoff__section compose-priority"
      aria-label="Priority"
    >
      <h2 className="compose-handoff__title">Priority</h2>
      <div
        className="compose-priority__segments"
        role="radiogroup"
        aria-label="Priority"
      >
        {PRIORITIES.map((p) => {
          const active = value === p;
          return (
            <button
              key={p}
              type="button"
              role="radio"
              aria-checked={active}
              disabled={disabled}
              data-active={active ? "true" : "false"}
              data-priority={p}
              className="compose-priority__segment"
              onClick={() => onChange(p)}
            >
              <span className="compose-priority__dot" aria-hidden="true" />
              {LABELS[p]}
            </button>
          );
        })}
      </div>
    </section>
  );
}
