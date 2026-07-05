import { useEffect } from "react";
import { listTaskEvents } from "@/api";
import { errorMessage } from "@/lib/errorMessage";
import { phaseLabel } from "@/observability";
import { EmptyState } from "@/shared/EmptyState";
import type { TaskCycleDetail, TaskCycleStreamEvent } from "@/types";
import type { UseTaskCycleStreamResult } from "@/tasks/hooks/useTaskCycles";
import { TaskTimelineSkeleton } from "@/tasks/components/skeletons";
import { agentProgressKindDescriptor } from "@/tasks/cycleDisplay/agentProgressDisplay";
import { activityCountCaption } from "@/tasks/pages/attempt/filterActivityByPhase";
import type { TaskCycleDetailPageState } from "@/tasks/pages/attempt/useTaskCycleDetailPageState";
import { AttemptAuditTimeline } from "./AttemptAuditTimeline";
import { PhaseSeqBadge } from "./AttemptPhaseSeqBadge";

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

function CursorActivityPanel({
  panelId,
  tabId,
  streamQuery,
  streamEvents,
  visibleStreamEvents,
  showPhaseBadge,
  filterLabel,
  onClearPhaseFilter,
  onLoadMore,
}: {
  panelId: string;
  tabId: string;
  streamQuery: UseTaskCycleStreamResult;
  streamEvents: TaskCycleStreamEvent[];
  visibleStreamEvents: TaskCycleStreamEvent[];
  showPhaseBadge: boolean;
  filterLabel: string | null;
  onClearPhaseFilter: () => void;
  onLoadMore: () => void;
}) {
  return (
    <div
      role="tabpanel"
      id={panelId}
      aria-labelledby={tabId}
      className="task-attempt-activity-panel"
    >
      {streamQuery.isError ? (
        <div className="err" role="alert">
          <p>
            {errorMessage(streamQuery.error, "Could not load stream events.")}
          </p>
        </div>
      ) : streamEvents.length === 0 ? (
        filterLabel ? (
          <EmptyState
            title={`No Cursor output for ${filterLabel}`}
            description="Try another phase or show all activity."
            density="compact"
            hideIcon
            action={{ label: "Show all phases", onClick: onClearPhaseFilter }}
          />
        ) : (
          <EmptyState
            title="No Cursor output yet"
            description="Stream lines appear here as the agent runs."
            density="compact"
            hideIcon
          />
        )
      ) : (
        <>
          <ol
            className={
              showPhaseBadge
                ? "task-attempt-stream-list task-attempt-stream-list--numbered"
                : "task-attempt-stream-list"
            }
          >
            {visibleStreamEvents.map((ev) => (
              <StreamEventRow
                key={ev.id}
                ev={ev}
                showPhaseBadge={showPhaseBadge}
              />
            ))}
          </ol>
          <LoadMoreRows
            shown={visibleStreamEvents.length}
            total={streamEvents.length}
            itemLabel="updates"
            onLoadMore={onLoadMore}
          />
        </>
      )}
    </div>
  );
}

function AuditActivityPanel({
  panelId,
  tabId,
  auditQuery,
  auditEvents,
  visibleAuditEvents,
  taskId,
  filterLabel,
  onClearPhaseFilter,
  onLoadMore,
}: {
  panelId: string;
  tabId: string;
  auditQuery: TaskCycleDetailPageState["auditQuery"];
  auditEvents: NonNullable<
    Awaited<ReturnType<typeof listTaskEvents>>["events"]
  >;
  visibleAuditEvents: NonNullable<
    Awaited<ReturnType<typeof listTaskEvents>>["events"]
  >;
  taskId: string;
  filterLabel: string | null;
  onClearPhaseFilter: () => void;
  onLoadMore: () => void;
}) {
  return (
    <div
      role="tabpanel"
      id={panelId}
      aria-labelledby={tabId}
      className="task-attempt-activity-panel"
    >
      {auditQuery.isPending ? (
        <TaskTimelineSkeleton />
      ) : auditQuery.isError ? (
        <div className="err" role="alert">
          <p>{errorMessage(auditQuery.error, "Could not load audit events.")}</p>
          <div className="task-detail-error-actions">
            <button
              type="button"
              className="secondary"
              onClick={() => void auditQuery.refetch()}
            >
              Try again
            </button>
          </div>
        </div>
      ) : auditEvents.length === 0 ? (
        filterLabel ? (
          <EmptyState
            title={`No audit events for ${filterLabel}`}
            description="Try another phase or show all activity."
            density="compact"
            hideIcon
            action={{ label: "Show all phases", onClick: onClearPhaseFilter }}
          />
        ) : (
          <EmptyState
            title="No audit events yet"
            description="System events for this attempt appear here."
            density="compact"
            hideIcon
          />
        )
      ) : (
        <>
          <AttemptAuditTimeline
            events={visibleAuditEvents}
            taskId={taskId}
            ariaLabelledBy={tabId}
          />
          <LoadMoreRows
            shown={visibleAuditEvents.length}
            total={auditEvents.length}
            itemLabel="events"
            onLoadMore={onLoadMore}
          />
        </>
      )}
    </div>
  );
}

function LoadMoreRows({
  shown,
  total,
  itemLabel,
  onLoadMore,
}: {
  shown: number;
  total: number;
  itemLabel: string;
  onLoadMore: () => void;
}) {
  if (shown >= total) {
    return (
      <p className="task-attempt-count muted">
        All {total} {itemLabel} shown.
      </p>
    );
  }
  return (
    <div className="task-attempt-load-more">
      <p className="task-attempt-count muted">
        {shown} of {total} {itemLabel}
      </p>
      <button type="button" className="secondary" onClick={onLoadMore}>
        Load more
      </button>
    </div>
  );
}

function StreamEventRow({
  ev,
  showPhaseBadge,
}: {
  ev: TaskCycleStreamEvent;
  showPhaseBadge: boolean;
}) {
  const preview = ev.message || ev.tool || "Agent reported progress.";
  const kind = agentProgressKindDescriptor(ev.kind, ev.subtype, ev.tool);
  return (
    <li className="task-attempt-stream-row">
      <details className="task-attempt-stream-details">
        <summary className="task-attempt-stream-summary">
          <time className="task-attempt-stream-time" dateTime={ev.at}>
            {new Date(ev.at).toLocaleTimeString(undefined, {
              hour: "numeric",
              minute: "2-digit",
            })}
          </time>
          <span className="task-attempt-stream-label">
            <span
              className={`task-attempt-stream-kind task-attempt-stream-kind--${kind.tone}`}
              title={kind.title}
            >
              {kind.label}
            </span>
            <span className="task-attempt-stream-message" title={preview}>
              {preview}
            </span>
          </span>
          {showPhaseBadge ? <PhaseSeqBadge seq={ev.phase_seq} /> : null}
        </summary>
        <div className="task-attempt-stream-detail-panel">
          <dl className="task-attempt-stream-detail-list">
            {ev.tool ? (
              <div>
                <dt>Tool</dt>
                <dd>{ev.tool}</dd>
              </div>
            ) : null}
            <div>
              <dt>Phase</dt>
              <dd>#{ev.phase_seq}</dd>
            </div>
          </dl>
          <div className="task-attempt-stream-detail-block">
            <h4>Raw payload</h4>
            <pre>{JSON.stringify(ev.payload, null, 2)}</pre>
          </div>
        </div>
      </details>
    </li>
  );
}
