import { Link } from "react-router-dom";

export function AttemptDetailNavigation({ taskId }: { taskId: string }) {
  return (
    <nav
      className="task-detail-nav task-attempt-nav"
      aria-label="Attempt navigation"
    >
      <Link
        to="/"
        className="pd__back project-context-back-link task-attempt-nav-link"
      >
        All tasks
      </Link>
      <span className="task-attempt-nav-separator" aria-hidden="true">
        /
      </span>
      <Link
        to={`/tasks/${encodeURIComponent(taskId)}`}
        className="pd__back project-context-back-link task-attempt-nav-link"
      >
        Task
      </Link>
    </nav>
  );
}
