import type { TaskCycleDetail } from "@/types";
import {
  filterAuditEventsByPhase,
  filterStreamEventsByPhase,
} from "@/tasks/pages/attempt/filterActivityByPhase";
import { useAttemptPhaseFilter } from "@/tasks/pages/attempt/useAttemptPhaseFilter";
import type { TaskCycleDetailPageState } from "@/tasks/pages/attempt/useTaskCycleDetailPageState";
import {
  filterAuditEventsForCycle,
  sortStreamEventsNewestFirst,
} from "./attemptActivityHelpers";
import { AttemptActivitySection } from "./AttemptActivitySection";
import { AttemptCommitsSection } from "./AttemptCommitsSection";
import { AttemptDetailHeader } from "./AttemptDetailHeader";
import { AttemptDetailNavigation } from "./AttemptDetailNavigation";
import { AttemptPhasesSection } from "./AttemptPhasesSection";
import { buildAttemptTimelineDisplay } from "./attemptTimelineDisplay";

type Props = {
  pageState: TaskCycleDetailPageState;
  cycle: TaskCycleDetail;
};

export function AttemptDetailLoadedSection({ pageState, cycle }: Props) {
  const timelineDisplay = buildAttemptTimelineDisplay(cycle, pageState.now);
  const phaseFilter = useAttemptPhaseFilter(cycle.phases);
  const allStreamEvents = sortStreamEventsNewestFirst(pageState.streamQuery.events);
  const allAuditEvents = filterAuditEventsForCycle(
    pageState.auditQuery.data?.events,
    pageState.cycleId,
  );
  const streamEvents = filterStreamEventsByPhase(
    allStreamEvents,
    phaseFilter.filterPhaseSeq,
  );
  const auditEvents = filterAuditEventsByPhase(
    allAuditEvents,
    phaseFilter.filterPhaseSeq,
  );

  return (
    <section className="panel task-detail-panel task-attempt-detail task-detail-content--enter">
      <AttemptDetailNavigation taskId={pageState.taskId} />
      <AttemptDetailHeader cycle={cycle} timelineDisplay={timelineDisplay} />
      <AttemptPhasesSection
        taskId={pageState.taskId}
        cycleId={pageState.cycleId}
        cycle={cycle}
        timelineDisplay={timelineDisplay}
        filterPhaseSeq={phaseFilter.filterPhaseSeq}
        onSelectPhase={phaseFilter.setFilterPhaseSeq}
        phaseFilterEnabled={timelineDisplay.showPhaseBadge}
      />
      <AttemptCommitsSection
        taskId={pageState.taskId}
        cycleId={pageState.cycleId}
      />
      <AttemptActivitySection
        pageState={pageState}
        cycle={cycle}
        streamEvents={streamEvents}
        allStreamCount={allStreamEvents.length}
        auditEvents={auditEvents}
        allAuditCount={allAuditEvents.length}
        showPhaseBadge={timelineDisplay.showPhaseBadge}
        filterPhaseSeq={phaseFilter.filterPhaseSeq}
        onClearPhaseFilter={phaseFilter.clearFilter}
      />
    </section>
  );
}
