import type { ChangeEvent } from "react";

type Props = {
  enabled: boolean;
  disabled: boolean;
  onChange: (enabled: boolean) => void;
  /** Visible title; defaults to the create-modal wording. */
  label?: string;
  readyHint?: string;
  pausedHint?: string;
};

/**
 * Autonomy gate for the create-task modal.
 *
 * When `enabled` is false the parent flow submits the task with
 * `status: "on_hold"`. The agent worker picks up only `status: "ready"`
 * tasks (ReadyForAgentPickup, pkgs/tasks/store/internal/tasks/readiness.go),
 * so on-hold tasks sit untouched until the operator resumes them from
 * the task detail page.
 *
 * Sits between the primary fields and the "More options" details block
 * because it changes the most fundamental thing about the new task —
 * whether the agent is allowed to start working on it. Hiding this
 * behind "More options" buries a primary intent.
 */
export function TaskCreateModalAutonomyToggle({
  enabled,
  disabled,
  onChange,
  label = "Autonomous execution",
  readyHint = "Created as ready. The agent picks it up when no other task is running.",
  pausedHint = "Created paused until you resume from the task page.",
}: Props) {
  function handle(e: ChangeEvent<HTMLInputElement>) {
    onChange(e.target.checked);
  }
  return (
    <section className="task-create-autonomy" aria-label={label}>
      <label
        className="task-create-autonomy__row"
        htmlFor="task-create-autonomy-toggle"
      >
        <span className="task-create-autonomy__text">
          <span className="task-create-autonomy__label">{label}</span>
          <span className="task-create-autonomy__hint">
            {enabled ? readyHint : pausedHint}
          </span>
        </span>
        <span
          className="task-create-autonomy__switch"
          data-checked={enabled ? "true" : "false"}
          aria-hidden="true"
        >
          <span className="task-create-autonomy__switch-thumb" />
        </span>
        <input
          id="task-create-autonomy-toggle"
          type="checkbox"
          className="task-create-autonomy__input visually-hidden"
          role="switch"
          aria-checked={enabled}
          checked={enabled}
          disabled={disabled}
          onChange={handle}
        />
      </label>
    </section>
  );
}
