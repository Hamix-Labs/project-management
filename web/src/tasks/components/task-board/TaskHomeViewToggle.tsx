import type { ReactElement } from "react";
import type { TaskHomeView } from "../../pages/taskHomeView";

type Props = {
  value: TaskHomeView;
  onChange: (view: TaskHomeView) => void;
};

const LIST_PANEL_ID = "task-list-panel";
const BOARD_PANEL_ID = "task-board-panel";

const VIEW_ORDER: TaskHomeView[] = ["list", "board"];

const PANEL_BY_VIEW: Record<TaskHomeView, string> = {
  list: LIST_PANEL_ID,
  board: BOARD_PANEL_ID,
};

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

function neighborView(current: TaskHomeView, delta: 1 | -1): TaskHomeView {
  const idx = VIEW_ORDER.indexOf(current);
  const next = (idx + delta + VIEW_ORDER.length) % VIEW_ORDER.length;
  return VIEW_ORDER[next]!;
}

type TabDef = {
  view: TaskHomeView;
  label: string;
  icon: () => ReactElement;
};

const TABS: TabDef[] = [
  { view: "list", label: "List", icon: ListIcon },
  { view: "board", label: "Board", icon: BoardIcon },
];

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
      {TABS.map((tab) => {
        const Icon = tab.icon;
        const selected = value === tab.view;
        return (
          <button
            key={tab.view}
            type="button"
            role="tab"
            id={`task-home-view-tab-${tab.view}`}
            className={
              selected
                ? "task-home-view-toggle__btn task-home-view-toggle__btn--active"
                : "task-home-view-toggle__btn"
            }
            aria-selected={selected}
            aria-controls={PANEL_BY_VIEW[tab.view]}
            tabIndex={selected ? 0 : -1}
            onClick={() => onChange(tab.view)}
            onKeyDown={(e) => {
              if (e.key === "ArrowRight") {
                e.preventDefault();
                onChange(neighborView(value, 1));
              } else if (e.key === "ArrowLeft") {
                e.preventDefault();
                onChange(neighborView(value, -1));
              }
            }}
          >
            <Icon />
            {tab.label}
          </button>
        );
      })}
    </div>
  );
}
