import { Link } from "react-router-dom";
import { formatTimelineClock, timelineDateGroupLabel } from "./groupTimelineEvents";
import { TimelineKindGlyph } from "./TimelineGlyphs";
import type { TimelineEvent, TimelineKind } from "./timelineTypes";

type Props = {
  event: TimelineEvent;
  last: boolean;
  now?: Date;
};

const NODE_TONE: Record<TimelineKind, string> = {
  "verification-passed": "task-home-timeline-node--success",
  "verification-failed": "task-home-timeline-node--danger",
  "agent-started": "task-home-timeline-node--brand",
  "agent-finished": "task-home-timeline-node--brand",
  "status-changed": "task-home-timeline-node--neutral",
  "task-created": "task-home-timeline-node--neutral",
  "review-approved": "task-home-timeline-node--success",
  comment: "task-home-timeline-node--neutral",
};

export function TimelineEventItem({
  event,
  last,
  now = new Date(),
}: Props) {
  const at = new Date(event.at);
  const groupLabel = timelineDateGroupLabel(at, now);
  const clock = formatTimelineClock(at);
  const nodeClass = `task-home-timeline-node ${NODE_TONE[event.kind]}`;

  const eventLink =
    event.taskId && event.seq != null
      ? `/tasks/${encodeURIComponent(event.taskId)}/events/${event.seq}`
      : null;

  return (
    <li className="task-home-timeline-item">
      {!last ? (
        <span className="task-home-timeline-item__spine" aria-hidden="true" />
      ) : null}
      <span className={nodeClass} aria-hidden="true">
        <TimelineKindGlyph
          kind={event.kind}
          className="task-home-timeline-node__icon"
        />
      </span>
      <div className="task-home-timeline-item__body">
        <div className="task-home-timeline-item__meta">
          {eventLink ? (
            <Link
              className="task-home-timeline-item__time-link"
              to={eventLink}
            >
              <time dateTime={event.at}>
                {groupLabel}, {clock}
              </time>
            </Link>
          ) : (
            <time dateTime={event.at}>
              {groupLabel}, {clock}
            </time>
          )}
          {event.taskRef && event.taskId ? (
            <Link
              className="task-home-timeline-item__task-ref"
              to={`/tasks/${encodeURIComponent(event.taskId)}`}
            >
              {event.taskRef}
            </Link>
          ) : event.taskRef ? (
            <span className="task-home-timeline-item__task-ref">
              {event.taskRef}
            </span>
          ) : null}
        </div>
        <p className="task-home-timeline-item__title">
          <span className="task-home-timeline-item__title-lead">
            {event.title}
          </span>
          {event.highlight ? (
            <>
              {" "}
              <span className="task-home-timeline-item__title-highlight">
                {event.highlight}
              </span>
            </>
          ) : null}
        </p>
        <p className="task-home-timeline-item__detail">{event.detail}</p>
        {event.meta && event.meta.length > 0 ? (
          <div className="task-home-timeline-item__chips">
            {event.meta.map((chip) => (
              <span key={chip} className="task-home-timeline-chip">
                {chip}
              </span>
            ))}
          </div>
        ) : null}
      </div>
    </li>
  );
}
