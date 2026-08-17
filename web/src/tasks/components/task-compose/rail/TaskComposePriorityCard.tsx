import { PRIORITIES, type Priority, type PriorityChoice } from "@/types";
import { ComposeRailSectionTitle } from "../TaskComposeBriefCard";

type Props = {
  value: PriorityChoice;
  disabled?: boolean;
  onChange: (p: PriorityChoice) => void;
};

const LABELS: Record<Priority, string> = {
  low: "Low",
  medium: "Medium",
  high: "High",
  critical: "Critical",
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
      <ComposeRailSectionTitle icon={<ZapIcon />}>
        Priority
      </ComposeRailSectionTitle>
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
              {LABELS[p]}
            </button>
          );
        })}
      </div>
    </section>
  );
}

function ZapIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
    </svg>
  );
}
