import { Link } from "react-router-dom";
import { errorMessage } from "@/lib/errorMessage";

export function AttemptInvalidParamsSection() {
  return (
    <section className="panel task-detail-panel task-detail-content--enter">
      <div className="err" role="alert">
        <p>Missing task or attempt id in the URL.</p>
        <div className="task-detail-error-actions">
          <Link to="/" className="pd__back project-context-back-link">
            <span aria-hidden="true">&#8249;</span>
            All tasks
          </Link>
        </div>
      </div>
    </section>
  );
}

export function AttemptLoadingSection() {
  return (
    <section className="panel task-detail-panel task-attempt-detail task-detail-content--enter">
      <p className="muted" role="status" aria-busy="true">
        Loading attempt…
      </p>
    </section>
  );
}

export function AttemptErrorSection({
  taskId,
  error,
  onRetry,
}: {
  taskId: string;
  error: Error;
  onRetry: () => void;
}) {
  return (
    <section className="panel task-detail-panel task-detail-content--enter">
      <div className="err" role="alert">
        <p>{errorMessage(error, "Could not load attempt.")}</p>
        <div className="task-detail-error-actions">
          <button type="button" className="secondary" onClick={onRetry}>
            Try again
          </button>
          <Link
            to={`/tasks/${encodeURIComponent(taskId)}`}
            className="pd__back project-context-back-link"
          >
            <span aria-hidden="true">&#8249;</span>
            Task
          </Link>
        </div>
      </div>
    </section>
  );
}
