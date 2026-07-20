import { useEffect } from "react";
import { listTaskEvents } from "@/api";
import { phaseLabel } from "@/tasks/cycleDisplay/cyclesViewModel";
import type { TaskCycleDetail, TaskCycleStreamEvent } from "@/types";
import { activityCountCaption } from "@/tasks/pages/attempt/filterActivityByPhase";
import type { TaskCycleDetailPageState } from "@/tasks/pages/attempt/useTaskCycleDetailPageState";
import { AuditActivityPanel } from "./AuditActivityPanel";
import { CursorActivityPanel } from "./CursorActivityPanel";

const STREAM_VISIBLE_INITIAL = 6;
const AUDIT_VISIBLE_INITIAL = 6;
const LOAD_MORE_STEP = 6;

type AttemptActivitySectionProps = {
  pageState: TaskCycleDetailPageState;
  cycle: TaskCycleDetail;
  streamEvents: TaskCycleStreamEvent[];
  allStreamCount: number;
  auditEvents: NonNullable<
    Awaited<ReturnType<typeof listTaskEvents>>["events"]
  >;
  allAuditCount: number;
  showPhaseBadge: boolean;
  filterPhaseSeq: number | null;
  onClearPhaseFilter: () => void;
};

export function AttemptActivitySection({
  pageState,
  cycle,
  streamEvents,
  allStreamCount,
  auditEvents,
  allAuditCount,
  showPhaseBadge,
  filterPhaseSeq,
  onClearPhaseFilter,
}: AttemptActivitySectionProps) {
  const {
    activityTab,
    setActivityTab,
    cursorTabId,
    auditTabId,
    cursorPanelId,
    auditPanelId,
    visibleStreamCount,
    setVisibleStreamCount,
    visibleAuditCount,
    setVisibleAuditCount,
    streamQuery,
    taskId,
  } = pageState;

  useEffect(() => {
    setVisibleStreamCount(STREAM_VISIBLE_INITIAL);
    setVisibleAuditCount(AUDIT_VISIBLE_INITIAL);
  }, [filterPhaseSeq, setVisibleStreamCount, setVisibleAuditCount]);

  const filteredPhase = filterPhaseSeq
    ? cycle.phases.find((p) => p.phase_seq === filterPhaseSeq)
    : undefined;
  const filterLabel =
    filteredPhase && filterPhaseSeq
      ? `${phaseLabel(filteredPhase.phase)} #${filterPhaseSeq}`
      : null;
  const streamCountCaption = activityCountCaption(streamEvents.length, allStreamCount);
  const auditCountCaption = activityCountCaption(auditEvents.length, allAuditCount);
  const visibleStreamEvents = streamEvents.slice(0, visibleStreamCount);
  const visibleAuditEvents = auditEvents.slice(0, visibleAuditCount);

  return (
    <section
      className="task-attempt-section task-attempt-section--activity"
      aria-labelledby="attempt-activity-heading"
    >
      <div className="task-attempt-section-heading-row">
        <h3 className="task-detail-subheading" id="attempt-activity-heading">
          <span>Activity</span>
        </h3>
        <div
          className="task-attempt-activity-tabs"
          role="tablist"
          aria-label="Attempt activity views"
        >
          <button
            type="button"
            role="tab"
            id={cursorTabId}
            aria-selected={activityTab === "cursor"}
            aria-controls={cursorPanelId}
            className={
              activityTab === "cursor"
                ? "task-attempt-activity-tab task-attempt-activity-tab--active"
                : "task-attempt-activity-tab"
            }
            onClick={() => setActivityTab("cursor")}
            title={streamCountCaption}
          >
            Cursor
            <span className="task-attempt-activity-tab-count">
              {streamEvents.length}
            </span>
          </button>
          <button
            type="button"
            role="tab"
            id={auditTabId}
            aria-selected={activityTab === "audit"}
            aria-controls={auditPanelId}
            className={
              activityTab === "audit"
                ? "task-attempt-activity-tab task-attempt-activity-tab--active"
                : "task-attempt-activity-tab"
            }
            onClick={() => setActivityTab("audit")}
            title={auditCountCaption}
          >
            Audit
            <span className="task-attempt-activity-tab-count">
              {auditEvents.length}
            </span>
          </button>
        </div>
      </div>

      {activityTab === "cursor" ? (
        <CursorActivityPanel
          panelId={cursorPanelId}
          tabId={cursorTabId}
          streamQuery={streamQuery}
          streamEvents={streamEvents}
          visibleStreamEvents={visibleStreamEvents}
          showPhaseBadge={showPhaseBadge}
          filterLabel={filterLabel}
          onClearPhaseFilter={onClearPhaseFilter}
          onLoadMore={() => setVisibleStreamCount((n) => n + LOAD_MORE_STEP)}
        />
      ) : (
        <AuditActivityPanel
          panelId={auditPanelId}
          tabId={auditTabId}
          auditQuery={pageState.auditQuery}
          auditEvents={auditEvents}
          visibleAuditEvents={visibleAuditEvents}
          taskId={taskId}
          filterLabel={filterLabel}
          onClearPhaseFilter={onClearPhaseFilter}
          onLoadMore={() => setVisibleAuditCount((n) => n + LOAD_MORE_STEP)}
        />
      )}
    </section>
  );
}
