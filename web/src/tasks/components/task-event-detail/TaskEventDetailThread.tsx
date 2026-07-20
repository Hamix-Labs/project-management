import type { TaskEvent } from "@/types";

export function TaskEventDetailThread(props: {
  thread: NonNullable<TaskEvent["response_thread"]>;
}) {
  const { thread } = props;
  if (thread.length === 0) {
    return null;
  }
  return (
    <div
      className="task-event-detail-thread"
      role="log"
      aria-label="Conversation on this event"
    >
      <h3 className="task-detail-subheading" id="task-event-thread-heading">
        <span>Conversation</span>
      </h3>
      <ul
        className="task-event-detail-thread-list"
        aria-labelledby="task-event-thread-heading"
      >
        {thread.map((m, i) => (
          <li
            key={`${m.at}-${i}`}
            className={`task-event-detail-thread-item task-event-detail-thread-item--${m.by}`}
          >
            <article className="task-event-detail-thread-bubble">
              <header className="task-event-detail-thread-meta">
                <span className="task-event-detail-thread-by">
                  {m.by === "agent" ? "Agent" : "You"}
                </span>
                <span
                  className="task-event-detail-thread-meta-sep"
                  aria-hidden="true"
                >
                  ·
                </span>
                <time
                  className="task-event-detail-thread-time"
                  dateTime={m.at}
                >
                  {new Date(m.at).toLocaleString()}
                </time>
              </header>
              <p className="task-event-detail-thread-body">{m.body}</p>
            </article>
          </li>
        ))}
      </ul>
    </div>
  );
}
