import { AttemptDetailLoadedSection } from "../components/task-detail/attempt/AttemptDetailLoadedSection";
import {
  AttemptErrorSection,
  AttemptInvalidParamsSection,
  AttemptLoadingSection,
} from "../components/task-detail/attempt/AttemptDetailStates";
import { useTaskCycleDetailPageState } from "./attempt/useTaskCycleDetailPageState";

/**
 * Route page for a single task execution attempt. Reads `taskId` and `cycleId`
 * from the URL, loads attempt metadata, Cursor stream events, and audit
 * timeline data, and renders the phase rail plus tabbed activity panels.
 */
export function TaskCycleDetailPage() {
  const pageState = useTaskCycleDetailPageState();

  if (!pageState.paramsValid) {
    return <AttemptInvalidParamsSection />;
  }
  if (pageState.cycleQuery.isPending) {
    return <AttemptLoadingSection />;
  }
  if (pageState.cycleQuery.isError) {
    return (
      <AttemptErrorSection
        taskId={pageState.taskId}
        error={pageState.cycleQuery.error}
        onRetry={() => void pageState.cycleQuery.refetch()}
      />
    );
  }

  return (
    <AttemptDetailLoadedSection
      pageState={pageState}
      cycle={pageState.cycleQuery.data}
    />
  );
}
