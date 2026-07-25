import { useQuery } from "@tanstack/react-query";
import { useMemo } from "react";
import { taskQueryKeys } from "@/lib/taskQueryKeys";
import { errorMessage } from "@/lib/errorMessage";
import { getTaskActivity } from "@/api/tasks";
import { timelineRangeCutoff } from "../components/task-timeline/timelineRange";
import type { TaskHomeView } from "../pages/taskHomeView";
import type { TimelineRangeId } from "../components/task-timeline/timelineTypes";
import type { TaskActivityEvent } from "@/types";

export type UseTasksActivityOptions = {
  view: TaskHomeView;
  /** When false, query stays disabled (same gate as other home queries). */
  dataEnabled?: boolean;
  range: TimelineRangeId;
};

/**
 * Fetches cross-task activity feed (`GET /tasks/activity`) gated on
 * `view === "timeline"`. Maps the active range to a `since` ISO timestamp
 * so the server filters by lookback window before the client groups by day.
 */
export function useTasksActivity({
  view,
  dataEnabled = true,
  range,
}: UseTasksActivityOptions) {
  const since = useMemo(() => {
    const cutoff = timelineRangeCutoff(range);
    return cutoff ? cutoff.toISOString() : undefined;
  }, [range]);

  const enabled = view === "timeline" && dataEnabled;

  const query = useQuery({
    queryKey: taskQueryKeys.activity(since, 0),
    queryFn: ({ signal }) => getTaskActivity({ signal, limit: 50, since }),
    enabled,
  });

  const events: TaskActivityEvent[] = useMemo(
    () => query.data?.events ?? [],
    [query.data?.events],
  );

  const total = query.data?.total ?? 0;
  const limit = query.data?.limit ?? 50;
  const offset = query.data?.offset ?? 0;

  return {
    events,
    total,
    hasMore: total > offset + events.length,
    truncated: total > limit,
    loading: enabled && query.isPending,
    error: query.isError ? errorMessage(query.error) : null,
    refetch: query.refetch,
  };
}
