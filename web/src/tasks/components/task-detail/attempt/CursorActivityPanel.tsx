import { errorMessage } from "@/lib/errorMessage";
import type { TaskCycleStreamEvent } from "@/types";
import type { UseTaskCycleStreamResult } from "@/tasks/hooks/useTaskCycles";
import { PhaseFilteredEmpty } from "./PhaseFilteredEmpty";
import { LoadMoreRows } from "./LoadMoreRows";
import { StreamEventRow } from "./StreamEventRow";

type CursorActivityPanelProps = {
  panelId: string;
  tabId: string;
  streamQuery: UseTaskCycleStreamResult;
  streamEvents: TaskCycleStreamEvent[];
  visibleStreamEvents: TaskCycleStreamEvent[];
  showPhaseBadge: boolean;
  filterLabel: string | null;
  onClearPhaseFilter: () => void;
  onLoadMore: () => void;
};

export function CursorActivityPanel({
  panelId,
  tabId,
  streamQuery,
  streamEvents,
  visibleStreamEvents,
  showPhaseBadge,
  filterLabel,
  onClearPhaseFilter,
  onLoadMore,
}: CursorActivityPanelProps) {
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
        <PhaseFilteredEmpty
          filterLabel={filterLabel}
          kind="cursor"
          onClearPhaseFilter={onClearPhaseFilter}
        />
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
