import type { TaskHomeView } from "../../pages/taskHomeView";

type Props = {
  value: TaskHomeView;
  onChange: (view: TaskHomeView) => void;
};

const LIST_PANEL_ID = "task-list-panel";
const BOARD_PANEL_ID = "task-board-panel";

function ListIcon() {
  return (
    <svg
      className="task-home-view-toggle__icon"
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M3 4h10M3 8h10M3 12h10"
        stroke="currentColor"
        strokeWidth="1.4"
        strokeLinecap="round"
      />
    </svg>
  );
}

function BoardIcon() {
  return (
    <svg
      className="task-home-view-toggle__icon"
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      aria-hidden="true"
    >
      <rect
        x="2.5"
        y="2.5"
        width="4.5"
        height="4.5"
        rx="1"
        stroke="currentColor"
        strokeWidth="1.3"
      />
      <rect
        x="9"
        y="2.5"
        width="4.5"
        height="4.5"
        rx="1"
        stroke="currentColor"
        strokeWidth="1.3"
      />
      <rect
        x="2.5"
        y="9"
        width="4.5"
        height="4.5"
        rx="1"
        stroke="currentColor"
        strokeWidth="1.3"
      />
      <rect
        x="9"
        y="9"
        width="4.5"
        height="4.5"
        rx="1"
        stroke="currentColor"
        strokeWidth="1.3"
      />
    </svg>
  );
}

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
        <ListIcon />
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
        <BoardIcon />
        Board
      </button>
    </div>
  );
}
