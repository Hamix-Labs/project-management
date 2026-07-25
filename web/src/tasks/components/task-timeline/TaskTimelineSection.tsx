import { useMemo, useState, type ReactNode } from "react";
import { TaskListSectionHeading } from "../task-list/section/TaskListSectionHeading";
import { groupTimelineEvents } from "./groupTimelineEvents";
import { TIMELINE_FIXTURES } from "./timelineFixtures";
import { TimelineEventItem } from "./TimelineEventItem";
import { TimelineRangeDropdown } from "./TimelineRangeDropdown";
import {
  DEFAULT_TIMELINE_RANGE,
  timelineRangeLabel,
} from "./timelineRange";
import type {
  TimelineEvent,
  TimelineFilterId,
  TimelineRangeId,
} from "./timelineTypes";

type Props = {
  actions?: ReactNode;
  /** Override fixtures (tests). */
  events?: TimelineEvent[];
  /** Fixed "now" for stable grouping in tests. */
  now?: Date;
};

const FILTERS: { id: TimelineFilterId; label: string }[] = [
  { id: "all", label: "All events" },
  { id: "tasks", label: "Tasks" },
  { id: "verification", label: "Verification" },
];

function emptyCopy(
  filter: TimelineFilterId,
  rangeId: TimelineRangeId,
): string {
  const category =
    filter === "all"
      ? "activity"
      : filter === "tasks"
        ? "task activity"
        : "verification activity";
  const range = timelineRangeLabel(rangeId).toLowerCase();
  if (rangeId === "all") {
    return `No ${category} to show.`;
  }
  return `No ${category} over the ${range}.`;
}

export function TaskTimelineSection({
  actions,
  events = TIMELINE_FIXTURES,
  now,
}: Props) {
  const [filter, setFilter] = useState<TimelineFilterId>("all");
  const [range, setRange] = useState<TimelineRangeId>(DEFAULT_TIMELINE_RANGE);
  const clock = now ?? new Date();

  const groups = useMemo(
    () => groupTimelineEvents(events, filter, range, clock),
    [events, filter, range, clock],
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
        <div
          className="task-home-timeline-filters"
          role="group"
          aria-label="Event category"
        >
          {FILTERS.map((f) => (
            <button
              key={f.id}
              type="button"
              className={
                filter === f.id
                  ? "task-home-timeline-filters__btn task-home-timeline-filters__btn--active"
                  : "task-home-timeline-filters__btn"
              }
              aria-pressed={filter === f.id}
              onClick={() => setFilter(f.id)}
            >
              {f.label}
            </button>
          ))}
        </div>
        <TimelineRangeDropdown value={range} onChange={setRange} />
      </div>

      <div className="task-home-timeline__feed">
        {total === 0 ? (
          <p className="task-home-timeline__empty" role="status">
            {emptyCopy(filter, range)}
          </p>
        ) : (
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
        )}
      </div>
    </section>
  );
}
