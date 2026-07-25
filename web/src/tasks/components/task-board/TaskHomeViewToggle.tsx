import type { TaskHomeView } from "../../pages/taskHomeView";

type Props = {
  value: TaskHomeView;
  onChange: (view: TaskHomeView) => void;
};

const LIST_PANEL_ID = "task-list-panel";
const BOARD_PANEL_ID = "task-board-panel";

/**
 * List | Board control for Task Home. Tab-like segmented buttons with
 * aria-controls pointing at the active view's tabpanel.
 */
export function TaskHomeViewToggle({ value, onChange }: Props) {
  return (
    <div
      className="task-home-view-toggle"
      role="tablist"
      aria-label="Task view"
    >
      <button
        type="button"
        role="tab"
        id="task-home-view-tab-list"
        className={
          value === "list"
            ? "task-home-view-toggle__btn task-home-view-toggle__btn--active"
            : "task-home-view-toggle__btn"
        }
        aria-selected={value === "list"}
        aria-controls={LIST_PANEL_ID}
        tabIndex={value === "list" ? 0 : -1}
        onClick={() => onChange("list")}
        onKeyDown={(e) => {
          if (e.key === "ArrowRight" || e.key === "ArrowLeft") {
            e.preventDefault();
            onChange("board");
          }
        }}
      >
        List
      </button>
      <button
        type="button"
        role="tab"
        id="task-home-view-tab-board"
        className={
          value === "board"
            ? "task-home-view-toggle__btn task-home-view-toggle__btn--active"
            : "task-home-view-toggle__btn"
        }
        aria-selected={value === "board"}
        aria-controls={BOARD_PANEL_ID}
        tabIndex={value === "board" ? 0 : -1}
        onClick={() => onChange("board")}
        onKeyDown={(e) => {
          if (e.key === "ArrowRight" || e.key === "ArrowLeft") {
            e.preventDefault();
            onChange("list");
          }
        }}
      >
        Board
      </button>
    </div>
  );
}
