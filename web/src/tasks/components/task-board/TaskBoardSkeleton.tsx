import { BOARD_COLUMNS } from "./boardColumns";

/** Four column placeholders matching denser board layout. */
export function TaskBoardSkeleton() {
  return (
    <div
      className="task-board-track task-board-track--skeleton"
      aria-hidden="true"
    >
      {BOARD_COLUMNS.map((col) => (
        <div
          key={col.id}
          className="task-board-column task-board-column--skeleton"
        >
          <div className="skeleton-block task-board-skeleton__head" />
          <div className="skeleton-block task-board-skeleton__card" />
          <div className="skeleton-block task-board-skeleton__card" />
          <div className="skeleton-block task-board-skeleton__card" />
        </div>
      ))}
    </div>
  );
}
