import { listTaskEvents } from "@/api";
import { errorMessage } from "@/lib/errorMessage";
import { TaskTimelineSkeleton } from "@/tasks/components/skeletons";
import type { TaskCycleDetailPageState } from "@/tasks/pages/attempt/useTaskCycleDetailPageState";
import { AttemptAuditTimeline } from "./AttemptAuditTimeline";
import { LoadMoreRows } from "./LoadMoreRows";
import { PhaseFilteredEmpty } from "./PhaseFilteredEmpty";

type AuditActivityPanelProps = {
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
};

export function AuditActivityPanel({
  panelId,
  tabId,
  auditQuery,
  auditEvents,
  visibleAuditEvents,
  taskId,
  filterLabel,
  onClearPhaseFilter,
  onLoadMore,
}: AuditActivityPanelProps) {
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
        <PhaseFilteredEmpty
          filterLabel={filterLabel}
          kind="audit"
          onClearPhaseFilter={onClearPhaseFilter}
        />
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
