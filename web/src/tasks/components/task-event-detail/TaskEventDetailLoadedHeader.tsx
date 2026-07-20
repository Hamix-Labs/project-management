import { Link } from "react-router-dom";
import { CopyableId } from "@/shared/CopyableId";
import type { TaskEvent } from "@/types";
import { eventTypeLabel, eventTypeNeedsUserInput } from "../../task-events";

export function TaskEventDetailLoadedHeader(props: {
  taskId: string;
  event: TaskEvent;
  needsInput: boolean;
  awaitingUser: boolean;
}) {
  const { taskId, event: ev, needsInput, awaitingUser } = props;
  return (
    <>
      <nav className="task-detail-nav" aria-label="Event navigation">
        <Link to="/" className="pd__back project-context-back-link">
          <span aria-hidden="true">&#8249;</span>
          All tasks
        </Link>
        <Link
          to={`/tasks/${encodeURIComponent(taskId)}`}
          className="pd__back project-context-back-link"
        >
          <span aria-hidden="true">&#8249;</span>
          Task
        </Link>
      </nav>

      <header className="task-event-detail-header">
        <h2 className="task-detail-title term-arrow">
          <span>Event #{ev.seq}</span>
        </h2>
        <p
          className="task-event-detail-stance"
          role="status"
          data-stance={needsInput ? "needs-user" : "informational"}
          data-awaiting-response={awaitingUser ? "true" : undefined}
        >
          {needsInput
            ? awaitingUser
              ? "Agent needs input"
              : "You replied — waiting on agent"
            : "Informational"}
        </p>
        <p className="task-event-detail-task-id">
          <span className="task-event-detail-task-id-label">Task</span>{" "}
          <CopyableId value={ev.task_id} />
        </p>
      </header>

      <dl className="task-event-detail-dl task-event-detail-dl--readable">
        <div>
          <dt>Type</dt>
          <dd>
            <code
              className="task-timeline-type-pill"
              data-event-type={ev.type}
              data-needs-user={needsInput ? "true" : undefined}
              title={eventTypeLabel(ev.type)}
            >
              {ev.type}
            </code>
          </dd>
        </div>
        <div>
          <dt>When</dt>
          <dd>
            <time dateTime={ev.at}>{new Date(ev.at).toLocaleString()}</time>
          </dd>
        </div>
        <div>
          <dt>By</dt>
          <dd className="task-timeline-by">{ev.by}</dd>
        </div>
      </dl>
    </>
  );
}
