import { useMemo, useState, type ReactNode } from "react";
import { TaskListSectionHeading } from "../task-list/section/TaskListSectionHeading";
import { groupTimelineEvents } from "./groupTimelineEvents";
import { TimelineEventItem } from "./TimelineEventItem";
import { TimelineRangeDropdown } from "./TimelineRangeDropdown";
import {
  DEFAULT_TIMELINE_RANGE,
  timelineRangeLabel,
} from "./timelineRange";
import { mapActivityEventsToTimeline } from "./activityMapper";
import { useTasksActivity } from "../../hooks/useTasksActivity";
import type { TimelineEvent, TimelineRangeId } from "./timelineTypes";
import type { TaskHomeView } from "../../pages/taskHomeView";

type Props = {
  actions?: ReactNode;
  view: TaskHomeView;
  dataEnabled?: boolean;
  /** Override events (tests). */
  events?: TimelineEvent[];
  /** Fixed "now" for stable grouping in tests. */
  now?: Date;
};

function emptyCopy(rangeId: TimelineRangeId): string {
  const range = timelineRangeLabel(rangeId).toLowerCase();
  if (rangeId === "all") {
    return "No activity to show.";
  }
  return `No activity over the ${range}.`;
}

const SKELETON_ROWS = 5;

function TimelineSkeleton() {
  return (
    <div
      className="task-home-timeline__skeleton"
      role="status"
      aria-label="Loading activity"
    >
      {Array.from({ length: SKELETON_ROWS }, (_, i) => (
        <div key={i} className="task-home-timeline-skeleton-row">
          <span className="task-home-timeline-skeleton-row__node" />
          <div className="task-home-timeline-skeleton-row__body">
            <span className="task-home-timeline-skeleton-row__line task-home-timeline-skeleton-row__line--short" />
            <span className="task-home-timeline-skeleton-row__line" />
            <span className="task-home-timeline-skeleton-row__line task-home-timeline-skeleton-row__line--medium" />
          </div>
        </div>
      ))}
    </div>
  );
}

export function TaskTimelineSection({
  actions,
  view,
  dataEnabled,
  events: eventsOverride,
  now,
}: Props) {
  const [range, setRange] = useState<TimelineRangeId>(DEFAULT_TIMELINE_RANGE);
  const clock = now ?? new Date();

  const activity = useTasksActivity({
    view,
    dataEnabled,
    range,
  });

  const mappedEvents = useMemo(
    () => mapActivityEventsToTimeline(activity.events),
    [activity.events],
  );

  const events = eventsOverride ?? mappedEvents;

  const groups = useMemo(
    () => groupTimelineEvents(events, clock),
    [events, clock],
  );
  const total = groups.reduce((n, g) => n + g.events.length, 0);

  return (
    <section
      className="panel task-list-section-panel task-list-section--redesign task-home-timeline"
      aria-labelledby="task-timeline-heading"
      id="task-timeline-panel"
      role="tabpanel"
    >
      <div className="task-list-toolbar">
        <TaskListSectionHeading
          title="Timeline"
          titleId="task-timeline-heading"
          actions={actions}
          description="A chronological view of project activity across tasks, agents, and verification."
        />
      </div>

      <div className="task-home-timeline__toolbar">
        <TimelineRangeDropdown value={range} onChange={setRange} />
      </div>

      <div className="task-home-timeline__feed">
        {activity.loading ? (
          <TimelineSkeleton />
        ) : activity.error ? (
          <div className="task-home-timeline__error" role="alert">
            <p className="task-home-timeline__error-msg">
              Could not load activity: {activity.error}
            </p>
            <button
              type="button"
              className="task-home-timeline__retry-btn"
              onClick={() => void activity.refetch()}
            >
              Retry
            </button>
          </div>
        ) : total === 0 ? (
          <p className="task-home-timeline__empty" role="status">
            {emptyCopy(range)}
          </p>
        ) : (
          <>
            <div className="task-home-timeline__groups">
              {groups.map((group) => (
                <section
                  key={group.label}
                  className="task-home-timeline-group"
                  aria-label={group.label}
                >
                  <h3 className="task-home-timeline-group__label">
                    {group.label}
                  </h3>
                  <ul className="task-home-timeline-group__list">
                    {group.events.map((event, i) => (
                      <TimelineEventItem
                        key={event.id}
                        event={event}
                        last={i === group.events.length - 1}
                        now={clock}
                      />
                    ))}
                  </ul>
                </section>
              ))}
            </div>
            {activity.truncated ? (
              <p className="task-home-timeline__truncation" role="status">
                Showing the {activity.events.length} most recent events. Narrow
                the date range to see more.
              </p>
            ) : null}
          </>
        )}
      </div>
    </section>
  );
}
