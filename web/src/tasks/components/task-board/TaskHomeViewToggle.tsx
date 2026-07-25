import type { TaskHomeView } from "../../pages/taskHomeView";

type Props = {
  value: TaskHomeView;
  onChange: (view: TaskHomeView) => void;
};

/**
 * Segmented List | Board control for Task Home. Uses aria-pressed
 * buttons (Plan 3 may promote to full tablist).
 */
export function TaskHomeViewToggle({ value, onChange }: Props) {
  return (
    <div
      className="task-home-view-toggle"
      role="group"
      aria-label="Task view"
    >
      <button
        type="button"
        className={
          value === "list"
            ? "task-home-view-toggle__btn task-home-view-toggle__btn--active"
            : "task-home-view-toggle__btn"
        }
        aria-pressed={value === "list"}
        onClick={() => onChange("list")}
      >
        List
      </button>
      <button
        type="button"
        className={
          value === "board"
            ? "task-home-view-toggle__btn task-home-view-toggle__btn--active"
            : "task-home-view-toggle__btn"
        }
        aria-pressed={value === "board"}
        onClick={() => onChange("board")}
      >
        Board
      </button>
    </div>
  );
}
