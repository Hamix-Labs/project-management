/** Skeleton placeholder while the cycles list query is pending. */
export function CyclesLoading() {
  return (
    <ul
      className="task-cycles-list task-cycles-list--loading"
      aria-busy="true"
      aria-label="Loading attempts"
    >
      <li className="task-cycle-row task-cycle-row--skeleton" />
      <li className="task-cycle-row task-cycle-row--skeleton" />
    </ul>
  );
}
