export function BoardActivePill({ count }: { count: number }) {
  return (
    <span className="task-board-active-pill">
      <span className="task-board-active-pill__dot" aria-hidden="true" />
      {count} active
    </span>
  );
}
